package app

import (
	"sync"
	"time"

	"github.com/p4tin/goaws/app/common"
)

type SqsErrorType struct {
	HttpError int
	Type      string
	Code      string
	Message   string
}

var SqsErrors map[string]SqsErrorType

type Message struct {
	MessageBody            []byte
	Uuid                   string
	MD5OfMessageAttributes string
	MD5OfMessageBody       string
	ReceiptHandle          string
	ReceiptTime            time.Time
	MessageAttributes      map[string]MessageAttributeValue
}

func (m *Message) GetResult() *ResultMessage {
	attrs := []*ResultMessageAttribute{}
	for _, attr := range m.MessageAttributes {
		attrs = append(attrs, attr.GetResult())
	}

	return &ResultMessage{
		MessageId:              m.Uuid,
		Body:                   m.MessageBody,
		ReceiptHandle:          m.ReceiptHandle,
		MD5OfBody:              common.GetMD5Hash(string(m.MessageBody)),
		MD5OfMessageAttributes: m.MD5OfMessageAttributes,
		MessageAttributes:      attrs,
	}
}

type MessageAttributeValue struct {
	Name     string
	DataType string
	Value    string
	ValueKey string
}

func (a *MessageAttributeValue) GetResult() *ResultMessageAttribute {
	v := &ResultMessageAttributeValue{
		DataType: a.DataType,
	}

	switch a.DataType {
	case "Binary":
		v.BinaryValue = a.Value
	case "String":
		v.StringValue = a.Value
	}

	return &ResultMessageAttribute{
		Name:  a.Name,
		Value: v,
	}
}

type Queue struct {
	Name        string
	URL         string
	Arn         string
	TimeoutSecs int
	Messages    []Message
}

var SyncQueues = struct {
	sync.RWMutex
	Queues map[string]*Queue
}{Queues: make(map[string]*Queue)}
