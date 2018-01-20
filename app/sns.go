package app

import "sync"

type SnsErrorType struct {
	HttpError int
	Type      string
	Code      string
	Message   string
}

var SnsErrors map[string]SnsErrorType

type Subscription struct {
	TopicArn        string
	Protocol        string
	SubscriptionArn string
	EndPoint        string
	Raw             bool
}

type Topic struct {
	Name          string
	Arn           string
	Subscriptions []*Subscription
}

type (
	Protocol         string
	MessageStructure string
)

const (
	ProtocolSQS     Protocol = "sqs"
	ProtocolDefault Protocol = "default"
)

const (
	MessageStructureJSON MessageStructure = "json"
)

// Predefined errors
const (
	ErrNoDefaultElementInJSON = "Invalid parameter: Message Structure - No default entry in JSON message body"
)

var SyncTopics = struct {
	sync.RWMutex
	Topics map[string]*Topic
}{Topics: make(map[string]*Topic)}
