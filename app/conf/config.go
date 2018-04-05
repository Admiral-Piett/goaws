package conf

import (
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
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
		log.Printf("err: %v\n", err)
		return ports
	}
	if env == "" {
		env = "Local"
	}

	if envs[env].Region == "" {
		app.CurrentEnvironment.Region = "local"
	}

	if envs[env].Port != "" {
		ports = []string{envs[env].Port}
	} else if envs[env].SqsPort != "" && envs[env].SnsPort != "" {
		ports = []string{envs[env].SqsPort, envs[env].SnsPort}
		app.CurrentEnvironment.Port = envs[env].SqsPort
	}

	app.CurrentEnvironment = envs[env]

	common.LogMessages = false
	common.LogFile = "./goaws_messages.log"

	if envs[env].LogMessages == true {
		common.LogMessages = true
		if envs[env].LogFile != "" {
			common.LogFile = envs[env].LogFile
		}
	}

	if app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout == 0 {
		app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout = 30
	}

	app.SyncQueues.Lock()
	app.SyncTopics.Lock()
	for _, queue := range envs[env].Queues {
		queueUrl := "http://" + app.CurrentEnvironment.Host + ":" + app.CurrentEnvironment.Port + "/queue/" + queue.Name
		queueArn := "arn:aws:sqs:" + app.CurrentEnvironment.Region + ":000000000000:" + queue.Name

		if queue.ReceiveMessageWaitTimeSeconds == 0 {
			queue.ReceiveMessageWaitTimeSeconds = app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds
		}

		app.SyncQueues.Queues[queue.Name] = &app.Queue{
			Name:                queue.Name,
			TimeoutSecs:         app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout,
			Arn:                 queueArn,
			URL:                 queueUrl,
			ReceiveWaitTimeSecs: queue.ReceiveMessageWaitTimeSeconds,
			IsFIFO:              app.HasFIFOQueueName(queue.Name),
		}
	}

	for _, topic := range envs[env].Topics {
		topicArn := "arn:aws:sns:" + app.CurrentEnvironment.Region + ":000000000000:" + topic.Name

		newTopic := &app.Topic{Name: topic.Name, Arn: topicArn}
		newTopic.Subscriptions = make([]*app.Subscription, 0, 0)

		for _, subs := range topic.Subscriptions {
			if _, ok := app.SyncQueues.Queues[subs.QueueName]; !ok {
				//Queue does not exist yet, create it.
				queueUrl := "http://" + app.CurrentEnvironment.Host + ":" + app.CurrentEnvironment.Port + "/queue/" + subs.QueueName
				queueArn := "arn:aws:sqs:" + app.CurrentEnvironment.Region + ":000000000000:" + subs.QueueName
				app.SyncQueues.Queues[subs.QueueName] = &app.Queue{
					Name:                subs.QueueName,
					TimeoutSecs:         app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout,
					Arn:                 queueArn,
					URL:                 queueUrl,
					ReceiveWaitTimeSecs: app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds,
					IsFIFO:              app.HasFIFOQueueName(subs.QueueName),
				}
			}
			qArn := app.SyncQueues.Queues[subs.QueueName].Arn
			newSub := &app.Subscription{EndPoint: qArn, Protocol: "sqs", TopicArn: topicArn, Raw: subs.Raw}
			subArn, _ := common.NewUUID()
			subArn = topicArn + ":" + subArn
			newSub.SubscriptionArn = subArn
			newTopic.Subscriptions = append(newTopic.Subscriptions, newSub)
		}
		app.SyncTopics.Topics[topic.Name] = newTopic
	}

	app.SyncQueues.Unlock()
	app.SyncTopics.Unlock()

	return ports
}
