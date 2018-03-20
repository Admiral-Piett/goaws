package app

import (
	"strconv"
	"strings"
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
	GroupID                string
}

type MessageAttributeValue struct {
	Name     string
	DataType string
	Value    string
	ValueKey string
}

type Queue struct {
	Name                string
	URL                 string
	Arn                 string
	TimeoutSecs         int
	ReceiveWaitTimeSecs int
	Messages            []Message
	DeadLetterQueue     *Queue
	MaxReceiveCount     int
	IsFIFO              bool
	FIFOMessages        map[string]int
	FIFOSequenceNumbers map[string]int
}

var SyncQueues = struct {
	sync.RWMutex
	Queues map[string]*Queue
}{Queues: make(map[string]*Queue)}

func HasFIFOQueueName(queueName string) bool {
	return strings.HasSuffix(queueName, ".fifo")
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
