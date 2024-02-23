package app

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

// CurrentEnvironment should get overwritten when the app starts up and loads the config.  For the
// sake of generating "partial" apps piece-meal during test automation we'll slap these placeholder
// values in here so the resource URLs aren't wonky like `http://://new-queue`.
var CurrentEnvironment = Environment{
	Host:      "host",
	Port:      "port",
	Region:    "region",
	AccountID: "accountID",
}

/*** Common ***/
type ResponseMetadata struct {
	RequestId string `xml:"RequestId"`
}

/*** Error Responses ***/
type ErrorResult struct {
	Type    string `xml:"Type,omitempty"`
	Code    string `xml:"Code,omitempty"`
	Message string `xml:"Message,omitempty"`
}

type ErrorResponse struct {
	Result    ErrorResult `xml:"Error"`
	RequestId string      `xml:"RequestId"`
}

type RandomLatency struct {
	Min int
	Max int
}
