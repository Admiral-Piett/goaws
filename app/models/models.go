package models

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type MessageStructure string
type Protocol string

type MessageAttribute struct {
	BinaryListValues []string `json:"BinaryListValues,omitempty" xml:"BinaryListValues,omitempty"` // currently unsupported by AWS
	BinaryValue      string   `json:"BinaryValue,omitempty" xml:"BinaryValue,omitempty"`
	DataType         string   `json:"DataType,omitempty" xml:"DataType,omitempty"`
	StringListValues []string `json:"StringListValues,omitempty" xml:"StringListValues,omitempty"` // currently unsupported by AWS
	StringValue      string   `json:"StringValue,omitempty" xml:"StringValue,omitempty"`
}

type SNSMessage struct {
	Type              string                      `json:"Type"`
	Token             string                      `json:"Token,omitempty"`
	MessageId         string                      `json:"MessageId"`
	TopicArn          string                      `json:"TopicArn"`
	Subject           string                      `json:"Subject"`
	Message           string                      `json:"Message"`
	Timestamp         string                      `json:"Timestamp"`
	SignatureVersion  string                      `json:"SignatureVersion"`
	Signature         string                      `json:"Signature,omitempty"`
	SigningCertURL    string                      `json:"SigningCertURL"`
	UnsubscribeURL    string                      `json:"UnsubscribeURL"`
	SubscribeURL      string                      `json:"SubscribeURL,omitempty"`
	MessageAttributes map[string]MessageAttribute `json:"MessageAttributes,omitempty"`
}

type Subscription struct {
	TopicArn        string
	Protocol        string
	SubscriptionArn string
	EndPoint        string
	Raw             bool
	FilterPolicy    *FilterPolicy
}

type Topic struct {
	Name          string
	Arn           string
	Subscriptions []*Subscription
}

// only simple "ExactMatch" string policy is supported at the moment
type FilterPolicy map[string][]string

// Function checks if MessageAttributes passed to Topic satisfy FilterPolicy set by subscription
func (fp *FilterPolicy) IsSatisfiedBy(msgAttrs map[string]MessageAttribute) bool {
	for policyAttrName, policyAttrValues := range *fp {
		attrValue, ok := msgAttrs[policyAttrName]
		if !ok {
			return false // the attribute has to be present in the message
		}

		// String, String.Array, Number data-types are allowed by SNS filter policies
		// however go-AWS currently only supports String filter policies. That feature can be added here
		// ref: https://docs.aws.amazon.com/sns/latest/dg/message-filtering.html
		if attrValue.DataType != "String" {
			return false
		}

		if !stringInSlice(attrValue.StringValue, policyAttrValues) {
			return false // the attribute value has to be among filtered ones
		}
	}

	return true
}

type SqsMessage struct {
	MessageBody            string
	Uuid                   string
	MD5OfMessageAttributes string
	MD5OfMessageBody       string
	ReceiptHandle          string
	ReceiptTime            time.Time
	VisibilityTimeout      time.Time
	NumberOfReceives       int
	Retry                  int
	MessageAttributes      map[string]MessageAttribute
	GroupID                string
	DeduplicationID        string
	SentTime               time.Time
	DelaySecs              int
}

func (m *SqsMessage) IsReadyForReceipt() bool {
	randomLatency, err := generateRandomLatency()
	if err != nil {
		log.Error(err)
		return true
	}
	showAt := m.SentTime.Add(randomLatency).Add(time.Duration(m.DelaySecs) * time.Second)
	return showAt.Before(time.Now())
}

type Queue struct {
	Name                          string
	URL                           string
	Arn                           string
	VisibilityTimeout             int // seconds
	ReceiveMessageWaitTimeSeconds int
	DelaySeconds                  int
	MaximumMessageSize            int
	MessageRetentionPeriod        int // seconds  // TODO - not used in the code yet
	Messages                      []SqsMessage
	DeadLetterQueue               *Queue
	MaxReceiveCount               int
	IsFIFO                        bool
	FIFOMessages                  map[string]int
	FIFOSequenceNumbers           map[string]int
	EnableDuplicates              bool
	Duplicates                    map[string]time.Time
}

func (q *Queue) NextSequenceNumber(groupId string) string {
	if _, ok := q.FIFOSequenceNumbers[groupId]; !ok {
		q.FIFOSequenceNumbers = map[string]int{
			groupId: 0,
		}
	}

	q.FIFOSequenceNumbers[groupId]++
	return strconv.Itoa(q.FIFOSequenceNumbers[groupId])
}

func (q *Queue) IsLocked(groupId string) bool {
	_, ok := q.FIFOMessages[groupId]
	return ok
}

func (q *Queue) LockGroup(groupId string) {
	if _, ok := q.FIFOMessages[groupId]; !ok {
		q.FIFOMessages = map[string]int{
			groupId: 0,
		}
	}
}

func (q *Queue) UnlockGroup(groupId string) {
	if _, ok := q.FIFOMessages[groupId]; ok {
		delete(q.FIFOMessages, groupId)
	}
}

func (q *Queue) IsDuplicate(deduplicationId string) bool {
	if !q.EnableDuplicates || !q.IsFIFO || deduplicationId == "" {
		return false
	}

	_, ok := q.Duplicates[deduplicationId]

	return ok
}

func (q *Queue) InitDuplicatation(deduplicationId string) {
	if !q.EnableDuplicates || !q.IsFIFO || deduplicationId == "" {
		return
	}

	if _, ok := q.Duplicates[deduplicationId]; !ok {
		q.Duplicates[deduplicationId] = time.Now()
	}
}
