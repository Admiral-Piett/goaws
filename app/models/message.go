package models

import (
	"errors"
	"sync"
	"time"

	"github.com/p4tin/goaws/app/common"
)

type Message struct {
	MessageBody            []byte
	Uuid                   string
	Attributes             map[string]MessageAttributeValue
	MD5OfMessageAttributes string
	MD5OfMessageBody       string
	ReceiptHandle          string
	ReceiptTime            time.Time
	VisibilityTime         time.Time
}

type ReceiptInfo struct {
	sync.RWMutex
	Timeout      time.Time
	MessageIndex int
	Message      *Message
}

var ReceiptInfos map[string]*ReceiptInfo

// CreateMessage creates a SQS message
func CreateMessage(body string, messageAttributes map[string]MessageAttributeValue) *Message {
	uuid, _ := common.NewUUID()
	return &Message{
		MessageBody:            []byte(body),
		Attributes:             messageAttributes,
		MD5OfMessageAttributes: HashAttributes(messageAttributes),
		MD5OfMessageBody:       common.GetMD5Hash(body),
		Uuid:                   uuid,
	}
}

// ChangeMessageVisiblity changes the visiblity timeout on the message
func (msg Message) ChangeMessageVisiblity(d int) error {
	if msg.ReceiptHandle != "" {
		ReceiptInfos[msg.ReceiptHandle].Lock()
		t := ReceiptInfos[msg.ReceiptHandle].Timeout
		ReceiptInfos[msg.ReceiptHandle].Timeout = t.Add(time.Second * time.Duration(d))
		ReceiptInfos[msg.ReceiptHandle].Unlock()
		return nil
	}
	return errors.New("Message is not in flight")
}
