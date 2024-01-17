package fixtures

import "github.com/Admiral-Piett/goaws/app"

var ENV_SUBSCRIPTION_QUEUE_4 = app.EnvSubsciption{
	Protocol:     "",
	EndPoint:     "",
	TopicArn:     "",
	QueueName:    "local-queue4",
	Raw:          false,
	FilterPolicy: "",
}

var ENV_SUBSCRIPTION_QUEUE_5 = app.EnvSubsciption{
	Protocol:     "",
	EndPoint:     "",
	TopicArn:     "",
	QueueName:    "local-queue5",
	Raw:          true,
	FilterPolicy: "{\"foo\":[\"bar\"]}",
}

var LOCAL_ENV_TOPIC_1 = app.EnvTopic{
	Name: "local-topic1",
	Subscriptions: []app.EnvSubsciption{
		ENV_SUBSCRIPTION_QUEUE_4,
		ENV_SUBSCRIPTION_QUEUE_5,
	},
}

var LOCAL_ENV_TOPIC_2 = app.EnvTopic{
	Name:          "local-topic2",
	Subscriptions: []app.EnvSubsciption(nil),
}

var LOCAL_ENV_QUEUE_1 = app.EnvQueue{
	Name:                          "local-queue1",
	ReceiveMessageWaitTimeSeconds: 0,
	RedrivePolicy:                 "",
	MaximumMessageSize:            0,
}

var LOCAL_ENV_QUEUE_2 = app.EnvQueue{
	Name:                          "local-queue2",
	ReceiveMessageWaitTimeSeconds: 20,
	RedrivePolicy:                 "",
	MaximumMessageSize:            128,
}

var LOCAL_ENV_QUEUE_3 = app.EnvQueue{
	Name:                          "local-queue3",
	ReceiveMessageWaitTimeSeconds: 0,
	RedrivePolicy:                 "{\"maxReceiveCount\": 100, \"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:000000000000:local-queue3-dlq\"}",
	MaximumMessageSize:            0,
}

var LOCAL_ENV_QUEUE_3_DLQ = app.EnvQueue{
	Name:                          "local-queue3-dlq",
	ReceiveMessageWaitTimeSeconds: 0,
	RedrivePolicy:                 "",
	MaximumMessageSize:            0,
}

var DEFAULT_ENVIRONMENT = app.Environment{
	Host:      "localhost",
	Port:      "4100",
	Region:    "local",
	AccountID: "queue",
	QueueAttributeDefaults: app.EnvQueueAttributes{
		VisibilityTimeout:             30,
		ReceiveMessageWaitTimeSeconds: 0,
		MaximumMessageSize:            262144,
	},
	RandomLatency: app.RandomLatency{
		Min: 0,
		Max: 0,
	},
}

var NO_QUEUES_NO_TOPICS_ENVIRONEMENT = app.Environment{
	Host:      "localhost",
	Port:      "4100",
	Region:    "eu-west-1",
	LogFile:   "./goaws_messages.log",
	AccountID: "queue",
	QueueAttributeDefaults: app.EnvQueueAttributes{
		VisibilityTimeout:             30,
		ReceiveMessageWaitTimeSeconds: 0,
		MaximumMessageSize:            262144,
	},
	RandomLatency: app.RandomLatency{
		Min: 0,
		Max: 0,
	},
}

var LOCAL_ENVIRONMENT = app.Environment{
	Host:             "localhost",
	Port:             "4200",
	SqsPort:          "",
	SnsPort:          "",
	Region:           "us-east-1",
	AccountID:        "100010001000",
	LogToFile:        false,
	LogFile:          "./goaws_messages.log",
	EnableDuplicates: false,
	Topics: []app.EnvTopic{
		LOCAL_ENV_TOPIC_1,
		LOCAL_ENV_TOPIC_2,
	},
	Queues: []app.EnvQueue{
		LOCAL_ENV_QUEUE_1,
		LOCAL_ENV_QUEUE_2,
		LOCAL_ENV_QUEUE_3,
		LOCAL_ENV_QUEUE_3_DLQ,
	},
	QueueAttributeDefaults: app.EnvQueueAttributes{
		VisibilityTimeout:             10,
		ReceiveMessageWaitTimeSeconds: 11,
		MaximumMessageSize:            1024,
		MessageRetentionPeriod:        1000,
	},
	RandomLatency: app.RandomLatency{
		Min: 0,
		Max: 0,
	},
}
