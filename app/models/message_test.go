package models

import (
	"testing"
	"time"
)

func TestCreateMessage(t *testing.T) {
	attrs := make(map[string]MessageAttributeValue)
	attrs["test"] = MessageAttributeValue{
		dataType: "string",
		value:    "test",
		valueKey: "test",
	}
	expectedHash := HashAttributes(attrs)

	msg := CreateMessage("This is the body", attrs)
	if string(msg.MessageBody) != "This is the body" {
		t.Errorf("Bad Message Body on create test, wanted 'This is the body' but got '%s'", string(msg.MessageBody))
	}

	if expectedHash != msg.MD5OfMessageAttributes {
		t.Errorf("Bad Message Attrs Hash on create test, wanted '%s' but got '%s'", expectedHash, msg.MD5OfMessageAttributes)
	}
}

func TestChangeVisibilityTimeout(t *testing.T) {
	msg := CreateMessage("This is the body", nil)
	msg.ReceiptHandle = "abc"
	ReceiptInfos := &ReceiptInfos
	ReceiptInfos.Receipts = make(map[string]*ReceiptInfo)

	firstTimeout := 10
	secondTimeout := 10

	now := time.Now()
	expectedTimeout := now.Add(time.Second * time.Duration(20))

	ReceiptInfos.Receipts["abc"] = &ReceiptInfo{
		Timeout: now.Add(time.Second * time.Duration(firstTimeout)),
		Message: msg,
	}

	msg.ChangeMessageVisiblity(secondTimeout)

	if ReceiptInfos.Receipts["abc"].Timeout != expectedTimeout {
		t.Errorf("Bad Message timeout on change visibility test, wanted '%s' but got '%s'", expectedTimeout, msg.VisibilityTime)
	}
}
