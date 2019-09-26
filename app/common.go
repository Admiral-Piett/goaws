package app

/*** config ***/
type EnvSubsciption struct {
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
}

type EnvQueueAttributes struct {
	VisibilityTimeout             int
	ReceiveMessageWaitTimeSeconds int
}

type Environment struct {
	Host                   string
	Port                   string
	SqsPort                string
	SnsPort                string
	Region                 string
	AccountID              string
	LogMessages            bool
	LogFile                string
	Topics                 []EnvTopic
	Queues                 []EnvQueue
	QueueAttributeDefaults EnvQueueAttributes
}

var CurrentEnvironment Environment

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
