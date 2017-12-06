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
	Timeout time.Time
	Message *Message
}

var ReceiptInfos = struct {
	sync.RWMutex
	Receipts map[string]*ReceiptInfo
}{Receipts: make(map[string]*ReceiptInfo)}

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
		ReceiptInfos.Lock()
		t := ReceiptInfos.Receipts[msg.ReceiptHandle].Timeout
		ReceiptInfos.Receipts[msg.ReceiptHandle].Timeout = t.Add(time.Second * time.Duration(d))
		ReceiptInfos.Unlock()
		return nil
	}
	return errors.New("Message is not in flight")
}
