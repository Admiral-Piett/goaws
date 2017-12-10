package models

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
)

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

func CreateQueue(host, name string, timeout int) *Queue {
	log.WithFields(log.Fields{
		"host":    host,
		"queue":   name,
		"timeout": timeout,
	}).Debug("CreateQueue")

	queueUrl := "http://" + host + "/queue/" + name
	queue := &Queue{
		Name:        name,
		TimeoutSecs: timeout,
		Messages:    make([]Message, 0),
		URL:         queueUrl,
		Arn:         queueUrl,
	}
	SyncQueues.Lock()
	SyncQueues.Queues[name] = queue
	SyncQueues.Unlock()
	return queue
}

func AddMessageToQueue(name string, attributes map[string]MessageAttributeValue, messageBody string) (*Message, error) {
	var err error
	log.WithFields(log.Fields{
		"queue": name,
		"size":  len(SyncQueues.Queues[name].Messages),
		"msg":   messageBody,
	}).Debug("AddMessageToQueue")

	if _, ok := SyncQueues.Queues[name]; !ok {
		// Queue does not exists
		err = errors.New("QueueNotFound")
		return nil, err
	}

	msg := CreateMessage(messageBody, attributes)
	SyncQueues.Lock()
	SyncQueues.Queues[name].Messages = append(SyncQueues.Queues[name].Messages, *msg)
	SyncQueues.Unlock()

	return msg, nil
}

func RemoveMessageFromQueue(name string, receiptHandle string) error {
	log.WithFields(log.Fields{
		"queue":         name,
		"receiptHandle": receiptHandle,
		"size":          len(SyncQueues.Queues[name].Messages),
	}).Debug("RemoveMessageFromQueue")

	if receipt, ok := ReceiptInfos[receiptHandle]; ok {
		i := receipt.MessageIndex
		SyncQueues.Lock()
		SyncQueues.Queues[name].Messages = append(SyncQueues.Queues[name].Messages[:i], SyncQueues.Queues[name].Messages[i+1:]...)
		SyncQueues.Unlock()

		receipt.Lock()
		delete(ReceiptInfos, receiptHandle)
		receipt.Unlock()

		return nil
	} else {
		return errors.New("Not Found")
	}
}
