package conf

import (
	"path/filepath"
	"io/ioutil"
	"fmt"

	sqs "github.com/p4tin/goaws/gosqs"
	sns "github.com/p4tin/goaws/gosns"
	"github.com/p4tin/goaws/common"
	"github.com/ghodss/yaml"
)

type EnvSubsciption struct {
	QueueName string
	Raw bool
}

type EnvTopic struct {
	Name string
	Subscriptions []EnvSubsciption
}

type EnvQueue struct {
	Name string
}

type Environment struct {
	Host string
	Port string
	Topics []EnvTopic
	Queues []EnvQueue
}

var envs map[string]Environment


func LoadYamlConfig(env string, portNumber string) {
	filename, _ := filepath.Abs("./goaws.yaml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &envs)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	if env == "" {
		env = "Local"
	}
	sqs.SyncQueues.Lock()
	for _, queue := range envs[env].Queues {
		fmt.Println(queue.Name)
		queueUrl := "http://" + envs[env].Host + ":" + envs[env].Port +"/queue/" + queue.Name
		sqs.SyncQueues.Queues[queue.Name] = &sqs.Queue{Name: queue.Name, TimeoutSecs: 30, Arn: queueUrl, URL: queueUrl}
	}
	sqs.SyncQueues.Unlock()
	sns.SyncTopics.Lock()
	for _, topic := range envs["Dev"].Topics {
		fmt.Println(topic.Name)
		topicArn := "arn:aws:sns:local:000000000000:" + topic.Name

		newTopic := &sns.Topic{Name: topic.Name, Arn: topicArn}
		newTopic.Subscriptions = make([]*sns.Subscription, 0 ,0)

		for _, subs := range topic.Subscriptions {
			if _, ok := sqs.SyncQueues.Queues[subs.QueueName] ; !ok {
				//Queue does not exist yet, create it.
				sqs.SyncQueues.Lock()
				queueUrl := "http://" + envs[env].Host + ":" + envs[env].Port +"/queue/" + subs.QueueName
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
}
