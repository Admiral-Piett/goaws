package app

import (
	"sync"
)

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
	FilterPolicy    *FilterPolicy
}

// only simple "ExactMatch" string policy is supported at the moment
type FilterPolicy map[string][]string

func (fp *FilterPolicy) IsSatisfiedBy(msgAttrs *TopicMessageAttributes) bool {
	for policyAttrName, policyAttrValues := range *fp {
		attrValue, ok := (*msgAttrs)[policyAttrName]
		if !ok {
			return false // the attribute has to be present in the message
		}

		if !stringInSlice(attrValue, policyAttrValues) {
			return false // the attribute value has to be among filtered ones
		}
	}

	return true
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type Topic struct {
	Name          string
	Arn           string
	Subscriptions []*Subscription
}

type TopicMessageAttributes map[string]string

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
