package models

/*** config ***/
type EnvSubsciption struct {
	Protocol     string
	EndPoint     string
	TopicArn     string
	QueueName    string
	Raw          bool
	FilterPolicy string
}

type EnvTopic struct {
	Name          string
	Subscriptions []EnvSubsciption
}

type EnvQueue struct {
	Name                          string
	ReceiveMessageWaitTimeSeconds int
	RedrivePolicy                 string
	MaximumMessageSize            int
	VisibilityTimeout             int
	MessageRetentionPeriod        int
}

type EnvQueueAttributes struct {
	VisibilityTimeout             int
	ReceiveMessageWaitTimeSeconds int
	MaximumMessageSize            int
	MessageRetentionPeriod        int // seconds
}

type Environment struct {
	Host                   string
	Port                   string
	SqsPort                string
	SnsPort                string
	Region                 string
	AccountID              string
	LogToFile              bool
	LogFile                string
	EnableDuplicates       bool
	Topics                 []EnvTopic
	Queues                 []EnvQueue
	QueueAttributeDefaults EnvQueueAttributes
	RandomLatency          RandomLatency
}

type RandomLatency struct {
	Min int
	Max int
}
