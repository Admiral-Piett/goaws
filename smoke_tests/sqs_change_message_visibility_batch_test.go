package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ChangeMessageVisibilityBatchV1_json_success(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	messageBody1 := "test message body 1"
	messageBody2 := "test message body 2"
	messageBody3 := "test message body 3"

	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: &messageBody1,
		QueueUrl:    createQueueResponse.QueueUrl,
	})
	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: &messageBody2,
		QueueUrl:    createQueueResponse.QueueUrl,
	})
	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: &messageBody3,
		QueueUrl:    createQueueResponse.QueueUrl,
	})

	receiveMessageOutput, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            createQueueResponse.QueueUrl,
		MaxNumberOfMessages: 10,
	})
	assert.Nil(t, err)
	assert.Len(t, receiveMessageOutput.Messages, 3)

	testId1 := "test1"
	testId2 := "test2"
	testId3 := "test3"

	changeVisibilityBatchOutput, err := sqsClient.ChangeMessageVisibilityBatch(context.TODO(), &sqs.ChangeMessageVisibilityBatchInput{
		Entries: []types.ChangeMessageVisibilityBatchRequestEntry{
			{
				Id:                &testId1,
				ReceiptHandle:     receiveMessageOutput.Messages[0].ReceiptHandle,
				VisibilityTimeout: 30,
			},
			{
				Id:                &testId2,
				ReceiptHandle:     receiveMessageOutput.Messages[1].ReceiptHandle,
				VisibilityTimeout: 60,
			},
			{
				Id:                &testId3,
				ReceiptHandle:     receiveMessageOutput.Messages[2].ReceiptHandle,
				VisibilityTimeout: 90,
			},
		},
		QueueUrl: createQueueResponse.QueueUrl,
	})

	assert.Nil(t, err)
	assert.Len(t, changeVisibilityBatchOutput.Successful, 3)
	assert.Empty(t, changeVisibilityBatchOutput.Failed)
}

func Test_ChangeMessageVisibilityBatchV1_json_error_queue_not_found(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	queueUrl := fmt.Sprintf("%s/%s", af.BASE_URL, "not-exist-queue")
	testId1 := "test1"
	receiptHandle1 := "handle1"

	_, err := sqsClient.ChangeMessageVisibilityBatch(context.TODO(), &sqs.ChangeMessageVisibilityBatchInput{
		Entries: []types.ChangeMessageVisibilityBatchRequestEntry{
			{
				Id:                &testId1,
				ReceiptHandle:     &receiptHandle1,
				VisibilityTimeout: 30,
			},
		},
		QueueUrl: &queueUrl,
	})

	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "AWS.SimpleQueueService.NonExistentQueue")
}

func Test_ChangeMessageVisibilityBatchV1_json_error_empty_batch(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	_, err := sqsClient.ChangeMessageVisibilityBatch(context.TODO(), &sqs.ChangeMessageVisibilityBatchInput{
		Entries:  make([]types.ChangeMessageVisibilityBatchRequestEntry, 0),
		QueueUrl: createQueueResponse.QueueUrl,
	})

	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "AWS.SimpleQueueService.EmptyBatchRequest")
}

func Test_ChangeMessageVisibilityBatchV1_json_error_too_many_entries(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	entries := make([]types.ChangeMessageVisibilityBatchRequestEntry, 11)
	for i := range entries {
		id := fmt.Sprintf("test%d", i+1)
		handle := fmt.Sprintf("handle%d", i+1)
		entries[i] = types.ChangeMessageVisibilityBatchRequestEntry{
			Id:                &id,
			ReceiptHandle:     &handle,
			VisibilityTimeout: 30,
		}
	}

	_, err := sqsClient.ChangeMessageVisibilityBatch(context.TODO(), &sqs.ChangeMessageVisibilityBatchInput{
		Entries:  entries,
		QueueUrl: createQueueResponse.QueueUrl,
	})

	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "AWS.SimpleQueueService.TooManyEntriesInBatchRequest")
}

func Test_ChangeMessageVisibilityBatchV1_json_error_ids_not_distinct(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	dupId := "duplicate-id"
	handle1 := "handle1"
	handle2 := "handle2"

	_, err := sqsClient.ChangeMessageVisibilityBatch(context.TODO(), &sqs.ChangeMessageVisibilityBatchInput{
		Entries: []types.ChangeMessageVisibilityBatchRequestEntry{
			{Id: &dupId, ReceiptHandle: &handle1, VisibilityTimeout: 30},
			{Id: &dupId, ReceiptHandle: &handle2, VisibilityTimeout: 60},
		},
		QueueUrl: createQueueResponse.QueueUrl,
	})

	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "AWS.SimpleQueueService.BatchEntryIdsNotDistinct")
}

func Test_ChangeMessageVisibilityBatchV1_xml_success(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	e := httpexpect.Default(t, server.URL)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	messageBody1 := "test message body 1"
	messageBody2 := "test message body 2"

	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: &messageBody1,
		QueueUrl:    createQueueResponse.QueueUrl,
	})
	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: &messageBody2,
		QueueUrl:    createQueueResponse.QueueUrl,
	})

	receiveMessageOutput, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            createQueueResponse.QueueUrl,
		MaxNumberOfMessages: 10,
	})
	assert.Nil(t, err)

	testId1 := "test1"
	testId2 := "test2"

	requestBodyXML := struct {
		Action   string `xml:"Action"`
		QueueUrl string `xml:"QueueUrl"`
		Version  string `xml:"Version"`
	}{
		Action:   "ChangeMessageVisibilityBatch",
		QueueUrl: *createQueueResponse.QueueUrl,
		Version:  "2012-11-05",
	}

	body := e.POST("/").
		WithForm(requestBodyXML).
		WithFormField("ChangeMessageVisibilityBatchRequestEntry.1.Id", testId1).
		WithFormField("ChangeMessageVisibilityBatchRequestEntry.1.ReceiptHandle", *receiveMessageOutput.Messages[0].ReceiptHandle).
		WithFormField("ChangeMessageVisibilityBatchRequestEntry.1.VisibilityTimeout", 30).
		WithFormField("ChangeMessageVisibilityBatchRequestEntry.2.Id", testId2).
		WithFormField("ChangeMessageVisibilityBatchRequestEntry.2.ReceiptHandle", *receiveMessageOutput.Messages[1].ReceiptHandle).
		WithFormField("ChangeMessageVisibilityBatchRequestEntry.2.VisibilityTimeout", 60).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	changeVisibilityBatchResponse := models.ChangeMessageVisibilityBatchResponse{}
	xml.Unmarshal([]byte(body), &changeVisibilityBatchResponse)

	assert.Len(t, changeVisibilityBatchResponse.Result.Successful, 2)
	assert.Empty(t, changeVisibilityBatchResponse.Result.Failed)
}
