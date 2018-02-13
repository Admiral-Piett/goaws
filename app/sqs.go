package app

import (
	"sync"
	"time"
)

type SqsErrorType struct {
	HttpError int
	Type      string
	Code      string
	Message   string
}

func (s *SqsErrorType) Error() string {
	return s.Type
}

var SqsErrors map[string]SqsErrorType

type Message struct {
	MessageBody            []byte
	Uuid                   string
	MD5OfMessageAttributes string
	MD5OfMessageBody       string
	ReceiptHandle          string
	ReceiptTime            time.Time
	VisibilityTimeout      time.Time
	Retry                  int
	MessageAttributes      map[string]MessageAttributeValue
}

type MessageAttributeValue struct {
	Name     string
	DataType string
	Value    string
	ValueKey string
}

type Queue struct {
	Name            string
	URL             string
	Arn             string
	TimeoutSecs     int
	Messages        []Message
	DeadLetterQueue *Queue
	MaxReceiveCount int
}

var SyncQueues = struct {
	sync.RWMutex
	Queues map[string]*Queue
}{Queues: make(map[string]*Queue)}
