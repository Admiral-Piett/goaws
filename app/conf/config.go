package conf

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
)

var envs map[string]models.Environment

func LoadYamlConfig(filename string, env string) []string {
	ports := []string{"4100"}

	if filename == "" {
		root, _ := filepath.Abs(".")
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if "goaws.yaml" == d.Name() {
				filename = path
			}
			return nil
		})
		if err != nil || filename == "" {
			log.Warn("Failure to find default config file")
			return ports
		}
	}

	filename, _ = filepath.Abs(filename)
	if _, err := os.Stat(filename); err != nil {
		log.Warnf("Failure to find config file: %s", filename)
		return ports
	}

	log.Infof("Loading config file: %s", filename)
	yamlFile, err := os.ReadFile(filename)
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
		models.CurrentEnvironment.Region = "local"
	}

	models.CurrentEnvironment = envs[env]

	if envs[env].Port != "" {
		ports = []string{envs[env].Port}
	} else if envs[env].SqsPort != "" && envs[env].SnsPort != "" {
		ports = []string{envs[env].SqsPort, envs[env].SnsPort}
		models.CurrentEnvironment.Port = envs[env].SqsPort
	}

	models.LogMessages = false
	models.LogFile = "./goaws_messages.log"
	if envs[env].LogToFile == true {
		models.LogMessages = true
		if envs[env].LogFile != "" {
			models.LogFile = envs[env].LogFile
		}
	}

	if models.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout <= 0 {
		models.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout = 30
	}

	if models.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize <= 0 {
		models.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize = 262144 // 256K
	}

	if models.CurrentEnvironment.QueueAttributeDefaults.MessageRetentionPeriod <= 0 {
		models.CurrentEnvironment.QueueAttributeDefaults.MessageRetentionPeriod = 345600 // 4 days
	}

	if models.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds <= 0 {
		models.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds = 0
	}

	if models.CurrentEnvironment.AccountID == "" {
		models.CurrentEnvironment.AccountID = "queue"
	}

	if models.CurrentEnvironment.Host == "" {
		models.CurrentEnvironment.Host = "localhost"
		models.CurrentEnvironment.Port = "4100"
	}

	models.SyncQueues.Lock()
	models.SyncTopics.Lock()
	for _, queue := range envs[env].Queues {
		queueUrl := "http://" + models.CurrentEnvironment.Host + ":" + models.CurrentEnvironment.Port +
			"/" + models.CurrentEnvironment.AccountID + "/" + queue.Name
		if models.CurrentEnvironment.Region != "" {
			queueUrl = "http://" + models.CurrentEnvironment.Region + "." + models.CurrentEnvironment.Host + ":" +
				models.CurrentEnvironment.Port + "/" + models.CurrentEnvironment.AccountID + "/" + queue.Name
		}
		queueArn := "arn:aws:sqs:" + models.CurrentEnvironment.Region + ":" + models.CurrentEnvironment.AccountID + ":" + queue.Name

		if queue.ReceiveMessageWaitTimeSeconds == 0 {
			queue.ReceiveMessageWaitTimeSeconds = models.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds
		}

		if queue.MaximumMessageSize == 0 {
			queue.MaximumMessageSize = models.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize
		}

		if queue.VisibilityTimeout == 0 {
			queue.VisibilityTimeout = models.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout
		}

		if queue.MessageRetentionPeriod == 0 {
			queue.MessageRetentionPeriod = models.CurrentEnvironment.QueueAttributeDefaults.MessageRetentionPeriod
		}

		models.SyncQueues.Queues[queue.Name] = &models.Queue{
			Name:                          queue.Name,
			VisibilityTimeout:             queue.VisibilityTimeout,
			Arn:                           queueArn,
			URL:                           queueUrl,
			ReceiveMessageWaitTimeSeconds: queue.ReceiveMessageWaitTimeSeconds,
			MaximumMessageSize:            queue.MaximumMessageSize,
			MessageRetentionPeriod:        queue.MessageRetentionPeriod,
			IsFIFO:                        utils.HasFIFOQueueName(queue.Name),
			EnableDuplicates:              models.CurrentEnvironment.EnableDuplicates,
			Duplicates:                    make(map[string]time.Time),
		}
	}

	// loop one more time to create queue's RedrivePolicy and assign deadletter queues in case dead letter queue is defined first in the config
	for _, queue := range envs[env].Queues {
		q := models.SyncQueues.Queues[queue.Name]
		if queue.RedrivePolicy != "" {
			err := setQueueRedrivePolicy(models.SyncQueues.Queues, q, queue.RedrivePolicy)
			if err != nil {
				log.Errorf("err: %s", err)
				return ports
			}
		}

	}

	for _, topic := range envs[env].Topics {
		topicArn := "arn:aws:sns:" + models.CurrentEnvironment.Region + ":" + models.CurrentEnvironment.AccountID + ":" + topic.Name

		newTopic := &models.Topic{Name: topic.Name, Arn: topicArn}
		newTopic.Subscriptions = make([]*models.Subscription, 0, 0)

		for _, subs := range topic.Subscriptions {
			var newSub *models.Subscription
			if strings.Contains(subs.Protocol, "http") {
				newSub = createHttpSubscription(subs)
			} else {
				//Queue does not exist yet, create it.
				newSub = createSqsSubscription(subs, topicArn)
			}
			if subs.FilterPolicy != "" {
				filterPolicy := &models.FilterPolicy{}
				err = json.Unmarshal([]byte(subs.FilterPolicy), filterPolicy)
				if err != nil {
					log.Errorf("err: %s", err)
					return ports
				}
				newSub.FilterPolicy = filterPolicy
			}

			newTopic.Subscriptions = append(newTopic.Subscriptions, newSub)
		}
		models.SyncTopics.Topics[topic.Name] = newTopic
	}

	models.SyncQueues.Unlock()
	models.SyncTopics.Unlock()

	return ports
}

func createHttpSubscription(configSubscription models.EnvSubsciption) *models.Subscription {
	newSub := &models.Subscription{EndPoint: configSubscription.EndPoint, Protocol: configSubscription.Protocol, TopicArn: configSubscription.TopicArn, Raw: configSubscription.Raw}
	subArn := uuid.NewString()
	subArn = configSubscription.TopicArn + ":" + subArn
	newSub.SubscriptionArn = subArn
	return newSub
}

func createSqsSubscription(configSubscription models.EnvSubsciption, topicArn string) *models.Subscription {
	if _, ok := models.SyncQueues.Queues[configSubscription.QueueName]; !ok {
		queueUrl := "http://" + models.CurrentEnvironment.Host + ":" + models.CurrentEnvironment.Port +
			"/" + models.CurrentEnvironment.AccountID + "/" + configSubscription.QueueName
		if models.CurrentEnvironment.Region != "" {
			queueUrl = "http://" + models.CurrentEnvironment.Region + "." + models.CurrentEnvironment.Host + ":" +
				models.CurrentEnvironment.Port + "/" + models.CurrentEnvironment.AccountID + "/" + configSubscription.QueueName
		}
		queueArn := "arn:aws:sqs:" + models.CurrentEnvironment.Region + ":" + models.CurrentEnvironment.AccountID + ":" + configSubscription.QueueName
		models.SyncQueues.Queues[configSubscription.QueueName] = &models.Queue{
			Name:                          configSubscription.QueueName,
			VisibilityTimeout:             models.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout,
			Arn:                           queueArn,
			URL:                           queueUrl,
			ReceiveMessageWaitTimeSeconds: models.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds,
			MaximumMessageSize:            models.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize,
			IsFIFO:                        utils.HasFIFOQueueName(configSubscription.QueueName),
			EnableDuplicates:              models.CurrentEnvironment.EnableDuplicates,
			Duplicates:                    make(map[string]time.Time),
		}
	}
	qArn := models.SyncQueues.Queues[configSubscription.QueueName].Arn
	newSub := &models.Subscription{EndPoint: qArn, Protocol: "sqs", TopicArn: topicArn, Raw: configSubscription.Raw}
	subArn := uuid.NewString()
	subArn = topicArn + ":" + subArn
	newSub.SubscriptionArn = subArn
	return newSub
}

func setQueueRedrivePolicy(queues map[string]*models.Queue, q *models.Queue, strRedrivePolicy string) error {
	// support both int and string maxReceiveCount (Amazon clients use string)
	redrivePolicy1 := struct {
		MaxReceiveCount     int    `json:"maxReceiveCount"`
		DeadLetterTargetArn string `json:"deadLetterTargetArn"`
	}{}
	redrivePolicy2 := struct {
		MaxReceiveCount     string `json:"maxReceiveCount"`
		DeadLetterTargetArn string `json:"deadLetterTargetArn"`
	}{}
	err1 := json.Unmarshal([]byte(strRedrivePolicy), &redrivePolicy1)
	err2 := json.Unmarshal([]byte(strRedrivePolicy), &redrivePolicy2)
	maxReceiveCount := redrivePolicy1.MaxReceiveCount
	deadLetterQueueArn := redrivePolicy1.DeadLetterTargetArn
	if err1 != nil && err2 != nil {
		return fmt.Errorf("invalid json for queue redrive policy ")
	} else if err1 != nil {
		maxReceiveCount, _ = strconv.Atoi(redrivePolicy2.MaxReceiveCount)
		deadLetterQueueArn = redrivePolicy2.DeadLetterTargetArn
	}

	if (deadLetterQueueArn != "" && maxReceiveCount == 0) ||
		(deadLetterQueueArn == "" && maxReceiveCount != 0) {
		return fmt.Errorf("invalid redrive policy values")
	}
	dlt := strings.Split(deadLetterQueueArn, ":")
	deadLetterQueueName := dlt[len(dlt)-1]
	deadLetterQueue, ok := queues[deadLetterQueueName]
	if !ok {
		return fmt.Errorf("deadletter queue not found")
	}
	q.DeadLetterQueue = deadLetterQueue
	q.MaxReceiveCount = maxReceiveCount

	return nil
}
