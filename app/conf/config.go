package conf

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
	"github.com/openlyinc/pointy"
	"github.com/p4tin/goaws/app"
	"github.com/p4tin/goaws/app/common"
)

var envs map[string]app.Environment

func LoadYamlConfig(filename string, env string) []string {
	ports := []string{"4100"}

	if filename == "" {
		filename, _ = filepath.Abs("./conf/goaws.yaml")
	}
	log.Warnf("Loading config file: %s", filename)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return ports
	}

	err = yaml.Unmarshal(yamlFile, &envs)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return ports
	}
	if env == "" {
		env = "Local"
	}

	if envs[env].Region == "" {
		app.CurrentEnvironment.Region = "local"
	}

	app.CurrentEnvironment = envs[env]

	if envs[env].Port != "" {
		ports = []string{envs[env].Port}
	} else if envs[env].SqsPort != "" && envs[env].SnsPort != "" {
		ports = []string{envs[env].SqsPort, envs[env].SnsPort}
		app.CurrentEnvironment.Port = envs[env].SqsPort
	}

	common.LogMessages = false
	common.LogFile = "./goaws_messages.log"

	if envs[env].LogToFile == true {
		common.LogMessages = true
		if envs[env].LogFile != "" {
			common.LogFile = envs[env].LogFile
		}
	}

	if app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout == 0 {
		app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout = 30
	}

	if app.CurrentEnvironment.AccountID == "" {
		app.CurrentEnvironment.AccountID = "queue"
	}

	if app.CurrentEnvironment.Host == "" {
		app.CurrentEnvironment.Host = "localhost"
		app.CurrentEnvironment.Port = "4100"
	}

	app.SyncQueues.Lock()
	app.SyncTopics.Lock()
	for _, queue := range envs[env].Queues {
		queueUrl := "http://" + app.CurrentEnvironment.Host + ":" + app.CurrentEnvironment.Port +
			"/" + app.CurrentEnvironment.AccountID + "/" + queue.Name
		if app.CurrentEnvironment.Region != "" {
			queueUrl = "http://" + app.CurrentEnvironment.Region + "." + app.CurrentEnvironment.Host + ":" +
				app.CurrentEnvironment.Port + "/" + app.CurrentEnvironment.AccountID + "/" + queue.Name
		}
		queueArn := "arn:aws:sqs:" + app.CurrentEnvironment.Region + ":" + app.CurrentEnvironment.AccountID + ":" + queue.Name

		if queue.ReceiveMessageWaitTimeSeconds == 0 {
			queue.ReceiveMessageWaitTimeSeconds = app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds
		}

		now := time.Now()
		app.SyncQueues.Queues[queue.Name] = &app.Queue{
			Name:                queue.Name,
			TimeoutSecs:         app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout,
			Arn:                 queueArn,
			URL:                 queueUrl,
			ReceiveWaitTimeSecs: queue.ReceiveMessageWaitTimeSeconds,
			IsFIFO:              app.HasFIFOQueueName(queue.Name),
			Created:             now,
			LastModified:        now,
			Tags:                make(map[string]string),
		}
	}

	for _, topic := range envs[env].Topics {
		topicArn := "arn:aws:sns:" + app.CurrentEnvironment.Region + ":" + app.CurrentEnvironment.AccountID + ":" + topic.Name

		newTopic := &app.Topic{
			Name:                    topic.Name,
			Arn:                     topicArn,
			EffectiveDeliveryPolicy: CreateDefaultDeliveryPolicy(),
			Policy:                  CreateDefaultTopicPolicy(topicArn),
		}
		newTopic.Subscriptions = make([]*app.Subscription, 0, 0)

		for _, subs := range topic.Subscriptions {
			var newSub *app.Subscription
			if strings.Contains(subs.Protocol, "http") {
				newSub = createHttpSubscription(subs)
			} else {
				//Queue does not exist yet, create it.
				newSub = createSqsSubscription(subs, topicArn)
			}
			if subs.FilterPolicy != "" {
				filterPolicy := &app.FilterPolicy{}
				err = json.Unmarshal([]byte(subs.FilterPolicy), filterPolicy)
				if err != nil {
					log.Errorf("err: %s", err)
					return ports
				}
				newSub.FilterPolicy = filterPolicy
			}

			newTopic.Subscriptions = append(newTopic.Subscriptions, newSub)
		}
		app.SyncTopics.Topics[topic.Name] = newTopic
		app.SyncTopics.ArnTopics[topicArn] = newTopic
	}

	app.SyncQueues.Unlock()
	app.SyncTopics.Unlock()

	return ports
}

func CreateDefaultDeliveryPolicy() *app.DeliveryPolicy {
	return &app.DeliveryPolicy{
		DefaultHealthyRetryPolicy: &app.RetryPolicy{
			NumberNoDelayRetries:  pointy.Int(0),
			NumberMinDelayRetries: pointy.Int(0),
			MinimumDelayTarget:    pointy.Int(20),
			NumberMaxDelayRetries: pointy.Int(0),
			MaximumDelayTarget:    pointy.Int(20),
			NumberRetries:         pointy.Int(3),
			BackoffFunction:       pointy.String("linear"),
		},
		Guaranteed: pointy.Bool(false),
	}
}

func CreateDefaultTopicPolicy(topicArn string) *app.TopicAttributePolicy {
	return &app.TopicAttributePolicy{
		Version: "2008-10-17",
		Id:      "__default_policy_ID",
		Statement: &[]app.AWSStatement{
			app.AWSStatement{
				Sid:    "__default_statement_ID",
				Effect: pointy.String("Allow"),
				Principal: app.AWSPrincipal{
					AWS: "*",
				},
				Action: []string{
					"SNS:GetTopicAttributes",
					"SNS:SetTopicAttributes",
					"SNS:AddPermission",
					"SNS:RemovePermission",
					"SNS:DeleteTopic",
					"SNS:Subscribe",
					"SNS:ListSubscriptionsByTopic",
					"SNS:Publish",
					"SNS:Receive",
				},
				Resource: topicArn,
				Condition: &map[string]interface{}{
					"StringEquals": map[string]string{
						"AWS:SourceOwner": app.CurrentEnvironment.AccountID,
					},
				},
			},
		},
	}
}

func createHttpSubscription(configSubscription app.EnvSubsciption) *app.Subscription {
	newSub := &app.Subscription{EndPoint: configSubscription.EndPoint, Protocol: configSubscription.Protocol, TopicArn: configSubscription.TopicArn, Raw: configSubscription.Raw}
	subArn, _ := common.NewUUID()
	subArn = configSubscription.TopicArn + ":" + subArn
	newSub.SubscriptionArn = subArn
	return newSub
}

func createSqsSubscription(configSubscription app.EnvSubsciption, topicArn string) *app.Subscription {
	if _, ok := app.SyncQueues.Queues[configSubscription.QueueName]; !ok {
		queueUrl := "http://" + app.CurrentEnvironment.Host + ":" + app.CurrentEnvironment.Port +
			"/" + app.CurrentEnvironment.AccountID + "/" + configSubscription.QueueName
		if app.CurrentEnvironment.Region != "" {
			queueUrl = "http://" + app.CurrentEnvironment.Region + "." + app.CurrentEnvironment.Host + ":" +
				app.CurrentEnvironment.Port + "/" + app.CurrentEnvironment.AccountID + "/" + configSubscription.QueueName
		}
		queueArn := "arn:aws:sqs:" + app.CurrentEnvironment.Region + ":" + app.CurrentEnvironment.AccountID + ":" + configSubscription.QueueName
		now := time.Now()
		app.SyncQueues.Queues[configSubscription.QueueName] = &app.Queue{
			Name:                configSubscription.QueueName,
			TimeoutSecs:         app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout,
			Arn:                 queueArn,
			URL:                 queueUrl,
			ReceiveWaitTimeSecs: app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds,
			IsFIFO:              app.HasFIFOQueueName(configSubscription.QueueName),
			Created:             now,
			LastModified:        now,
			Tags:                make(map[string]string),
		}
	}
	qArn := app.SyncQueues.Queues[configSubscription.QueueName].Arn
	newSub := &app.Subscription{EndPoint: qArn, Protocol: "sqs", TopicArn: topicArn, Raw: configSubscription.Raw}
	subArn, _ := common.NewUUID()
	subArn = topicArn + ":" + subArn
	newSub.SubscriptionArn = subArn
	return newSub
}
