package conf

import (
	"path/filepath"

	"github.com/p4tin/goaws/common"
    "github.com/spf13/viper"
	sns "github.com/p4tin/goaws/gosns"
	sqs "github.com/p4tin/goaws/gosqs"
    "strings"
    "github.com/juju/errors"
)

type EnvSubsciption struct {
	QueueName string
	Raw       bool
}

type EnvTopic struct {
	Name          string
	Subscriptions []EnvSubsciption
}

type EnvQueue struct {
	Name string
}

type Environment struct {
	Host        string
	SQSPort     string
	SNSPort     string
	Region      string
	LogMessages bool
	LogFile     string
	Topics      []EnvTopic
	Queues      []EnvQueue
}

var envs map[string]*Environment
var requiredEnv *Environment

func GetLoadedEnv() *Environment {
    return requiredEnv
}

func (e *Environment) SetSQSPort(port string) {
    e.SQSPort = port
}

func (e *Environment) SetSNSPort(port string) {
    e.SNSPort = port
}

func LoadYamlConfig(filename, env, sqsPortNumber, snsPortNumber string) (*Environment, error) {
    // Reset when loading
    sqs.ResetSyncQueues()
    sns.ResetSyncTopics()

    if filename == "" {
		filename, _ = filepath.Abs("./conf/goaws.yaml")
	}

    viper.SetConfigFile(filename)
    err := viper.ReadInConfig()
    if err != nil {
        return nil, err
    }

    err = viper.Unmarshal(&envs)
    if err != nil {
        return nil, err
    }

	if env == "" {
		env = "Local"
	}

    env = strings.ToLower(env)
    requiredEnv, ok := envs[env];
    if !ok {
        return nil, errors.Errorf("Error: Env %s was not found in config!", env)
    }

    region := "local"
    if envs[env].Region != "" {
        region = envs[env].Region
    }

    if sqsPortNumber == "" {
		sqsPortNumber = requiredEnv.SQSPort
		if sqsPortNumber == "" {
            sqsPortNumber = "9324"
        }
	} else {
        requiredEnv.SetSQSPort(sqsPortNumber)
    }

	if snsPortNumber == "" {
		snsPortNumber = requiredEnv.SNSPort
		if snsPortNumber == "" {
			snsPortNumber = "9292"
		}
	} else {
        requiredEnv.SetSNSPort(snsPortNumber)
    }

	common.LogMessages = false
	common.LogFile = "./goaws_messages.log"

	if envs[env].LogMessages == true {
		common.LogMessages = true
		if requiredEnv.LogFile != "" {
			common.LogFile = requiredEnv.LogFile
		}
	}

	sqs.SyncQueues.Lock()
	for _, queue := range requiredEnv.Queues {
		queueUrl := "http://" + requiredEnv.Host + ":" + sqsPortNumber + "/queue/" + queue.Name
		sqs.SyncQueues.Queues[queue.Name] = &sqs.Queue{Name: queue.Name, TimeoutSecs: 30, Arn: queueUrl, URL: queueUrl}
	}
	sqs.SyncQueues.Unlock()
	sns.SyncTopics.Lock()
	for _, topic := range requiredEnv.Topics {
		topicArn := "arn:aws:sns:"+region+":000000000000:" + topic.Name

		newTopic := &sns.Topic{Name: topic.Name, Arn: topicArn}
		newTopic.Subscriptions = make([]*sns.Subscription, 0, 0)

		for _, subs := range topic.Subscriptions {
			if _, ok := sqs.SyncQueues.Queues[subs.QueueName]; !ok {
				//Queue does not exist yet, create it.
				sqs.SyncQueues.Lock()
				queueUrl := "http://" + requiredEnv.Host + ":" + sqsPortNumber + "/queue/" + subs.QueueName
				sqs.SyncQueues.Queues[subs.QueueName] = &sqs.Queue{Name: subs.QueueName, TimeoutSecs: 30, Arn: queueUrl, URL: queueUrl}
				sqs.SyncQueues.Unlock()
			}
			qUrl := sqs.SyncQueues.Queues[subs.QueueName].URL
			newSub := &sns.Subscription{EndPoint: qUrl, Protocol: "sqs", TopicArn: topicArn, Raw: subs.Raw}
			subArn, _ := common.NewUUID()
			subArn = topicArn + ":" + subArn
			newSub.SubscriptionArn = subArn
			newTopic.Subscriptions = append(newTopic.Subscriptions, newSub)
		}
		sns.SyncTopics.Topics[topic.Name] = newTopic
	}
	sns.SyncTopics.Unlock()

	return requiredEnv, nil
}
