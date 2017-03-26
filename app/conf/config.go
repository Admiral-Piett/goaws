package conf

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/p4tin/goaws/app/common"
	sns "github.com/p4tin/goaws/app/gosns"
	sqs "github.com/p4tin/goaws/app/gosqs"
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
	Port        string
	SqsPort     string
	SnsPort     string
	Region      string
	LogMessages bool
	LogFile     string
	Topics      []EnvTopic
	Queues      []EnvQueue
}

var envs map[string]Environment

func LoadYamlConfig(filename string, env string) []string {
	ports := []string{"4100"}
	if filename == "" {
		filename, _ = filepath.Abs("./conf/goaws.yaml")
	}
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

	region := "local"
	if envs[env].Region != "" {
		region = envs[env].Region
	}

	if envs[env].Port != "" {
		ports = []string{envs[env].Port}
	} else if envs[env].SqsPort != "" && envs[env].SnsPort != "" {
		ports = []string{envs[env].SqsPort, envs[env].SnsPort}
	}

	common.LogMessages = false
	common.LogFile = "./goaws_messages.log"

	if envs[env].LogMessages == true {
		common.LogMessages = true
		if envs[env].LogFile != "" {
			common.LogFile = envs[env].LogFile
		}
	}

	sqs.SyncQueues.Lock()
	for _, queue := range envs[env].Queues {
		queueUrl := "http://" + envs[env].Host + ":" + ports[0] + "/queue/" + queue.Name
		sqs.SyncQueues.Queues[queue.Name] = &sqs.Queue{Name: queue.Name, TimeoutSecs: 30, Arn: queueUrl, URL: queueUrl}
	}
	sqs.SyncQueues.Unlock()
	sns.SyncTopics.Lock()
	for _, topic := range envs[env].Topics {
		topicArn := "arn:aws:sns:" + region + ":000000000000:" + topic.Name

		newTopic := &sns.Topic{Name: topic.Name, Arn: topicArn}
		newTopic.Subscriptions = make([]*sns.Subscription, 0, 0)

		for _, subs := range topic.Subscriptions {
			if _, ok := sqs.SyncQueues.Queues[subs.QueueName]; !ok {
				//Queue does not exist yet, create it.
				sqs.SyncQueues.Lock()
				queueUrl := "http://" + envs[env].Host + ":" + ports[0] + "/queue/" + subs.QueueName
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

	return ports
}
