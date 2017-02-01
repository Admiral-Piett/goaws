package conf

import (
	"github.com/p4tin/goaws/app"

	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/p4tin/goaws/app/common"
	sns "github.com/p4tin/goaws/app/gosns"
	sqs "github.com/p4tin/goaws/app/gosqs"
)

func LoadYamlConfig(filename string, env string) []string {
	ports := []string{"4100"}
	if filename == "" {
		var err error
		filename, err = filepath.Abs("./conf/goaws.yaml")
		if err != nil {
			log.Println(err)
		}
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return ports
	}

	err = yaml.Unmarshal(yamlFile, &app.Envs)
	if err != nil {
		log.Printf("err: %v\n", err)
		return ports
	}

	if env == "" {
		env = "Local"
	}
	log.Println(env)
	region := "local"
	log.Printf("%#v\n", app.Envs[env])
	if app.Envs[env].Region != "" {
		region = app.Envs[env].Region
	}

	if app.Envs[env].Port != "" {
		ports = []string{app.Envs[env].Port}
	} else if app.Envs[env].SqsPort != "" && app.Envs[env].SnsPort != "" {
		ports = []string{app.Envs[env].SqsPort, app.Envs[env].SnsPort}
	}

	common.LogMessages = false
	common.LogFile = "./goaws_messages.log"

	if app.Envs[env].LogMessages == true {
		common.LogMessages = true
		if app.Envs[env].LogFile != "" {
			common.LogFile = app.Envs[env].LogFile
		}
	}

	sqs.SyncQueues.Lock()
	for _, queue := range app.Envs[env].Queues {
		queueUrl := "http://" + app.Envs[env].Host + ":" + ports[0] + "/queue/" + queue.Name
		sqs.SyncQueues.Queues[queue.Name] = &sqs.Queue{Name: queue.Name, TimeoutSecs: 30, Arn: queueUrl, URL: queueUrl}
	}
	sqs.SyncQueues.Unlock()
	sns.SyncTopics.Lock()
	for _, topic := range app.Envs[env].Topics {
		topicArn := "arn:aws:sns:" + region + ":000000000000:" + topic.Name

		newTopic := &sns.Topic{Name: topic.Name, Arn: topicArn}
		newTopic.Subscriptions = make([]*sns.Subscription, 0, 0)

		for _, subs := range topic.Subscriptions {
			if _, ok := sqs.SyncQueues.Queues[subs.QueueName]; !ok {
				//Queue does not exist yet, create it.
				sqs.SyncQueues.Lock()
				queueUrl := "http://" + app.Envs[env].Host + ":" + ports[0] + "/queue/" + subs.QueueName
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
