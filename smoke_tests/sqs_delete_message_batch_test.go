package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_DeleteMessageBatchV1_json_error_queue_not_found(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	queueUrl := fmt.Sprintf("%s/%s", af.BASE_URL, "testing")

	testId1 := "test1"
	testId2 := "test2"
	testId3 := "test3"

	receiptHandle1 := "delete-test-1"
	receiptHandle2 := "delete-test-2"
	receiptHandle3 := "delete-test-3"

	_, error := sqsClient.DeleteMessageBatch(context.TODO(), &sqs.DeleteMessageBatchInput{
		Entries: []types.DeleteMessageBatchRequestEntry{
			{
				Id:            &testId1,
				ReceiptHandle: &receiptHandle1,
			},
			{
				Id:            &testId2,
				ReceiptHandle: &receiptHandle2,
			},
			{
				Id:            &testId3,
				ReceiptHandle: &receiptHandle3,
			},
		},
		QueueUrl: &queueUrl,
	})

	assert.Contains(t, error.Error(), "400")
	assert.Contains(t, error.Error(), "AWS.SimpleQueueService.NonExistentQueue")
	assert.Contains(t, error.Error(), "The specified queue does not exist for this wsdl version.")
}

func Test_DeleteMessageBatchV1_json_error_no_entry(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	crateQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	_, error := sqsClient.DeleteMessageBatch(context.TODO(), &sqs.DeleteMessageBatchInput{
		Entries:  make([]types.DeleteMessageBatchRequestEntry, 0),
		QueueUrl: crateQueueResponse.QueueUrl,
	})

	assert.Contains(t, error.Error(), "400")
	assert.Contains(t, error.Error(), "AWS.SimpleQueueService.EmptyBatchRequest")
	assert.Contains(t, error.Error(), "The batch request doesn't contain any entries.")
}

func Test_DeleteMessageBatchV1_json_error_too_many_entry(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	crateQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	testId1 := "test1"
	testId2 := "test2"
	testId3 := "test3"
	testId4 := "test4"
	testId5 := "test5"
	testId6 := "test6"
	testId7 := "test7"
	testId8 := "test8"
	testId9 := "test9"
	testId10 := "test10"
	testId11 := "test11"

	receiptHandle1 := "delete-test-1"
	receiptHandle2 := "delete-test-2"
	receiptHandle3 := "delete-test-3"
	receiptHandle4 := "delete-test-4"
	receiptHandle5 := "delete-test-5"
	receiptHandle6 := "delete-test-6"
	receiptHandle7 := "delete-test-7"
	receiptHandle8 := "delete-test-8"
	receiptHandle9 := "delete-test-9"
	receiptHandle10 := "delete-test-10"
	receiptHandle11 := "delete-test-11"

	_, error := sqsClient.DeleteMessageBatch(context.TODO(), &sqs.DeleteMessageBatchInput{
		Entries: []types.DeleteMessageBatchRequestEntry{
			{
				Id:            &testId1,
				ReceiptHandle: &receiptHandle1,
			},
			{
				Id:            &testId2,
				ReceiptHandle: &receiptHandle2,
			},
			{
				Id:            &testId3,
				ReceiptHandle: &receiptHandle3,
			},
			{
				Id:            &testId4,
				ReceiptHandle: &receiptHandle4,
			},
			{
				Id:            &testId5,
				ReceiptHandle: &receiptHandle5,
			},
			{
				Id:            &testId6,
				ReceiptHandle: &receiptHandle6,
			},
			{
				Id:            &testId7,
				ReceiptHandle: &receiptHandle7,
			},
			{
				Id:            &testId8,
				ReceiptHandle: &receiptHandle8,
			},
			{
				Id:            &testId9,
				ReceiptHandle: &receiptHandle9,
			},
			{
				Id:            &testId10,
				ReceiptHandle: &receiptHandle10,
			},
			{
				Id:            &testId11,
				ReceiptHandle: &receiptHandle11,
			},
		},
		QueueUrl: crateQueueResponse.QueueUrl,
	})

	assert.Contains(t, error.Error(), "400")
	assert.Contains(t, error.Error(), "AWS.SimpleQueueService.TooManyEntriesInBatchRequest")
	assert.Contains(t, error.Error(), "Maximum number of entries per request are 10.")
}

func Test_DeleteMessageBatchV1_json_error_batch_entry_ids_not_distinct(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	crateQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	testId1 := "test1"

	receiptHandle1 := "delete-test-1"
	receiptHandle2 := "delete-test-2"
	receiptHandle3 := "delete-test-3"

	_, error := sqsClient.DeleteMessageBatch(context.TODO(), &sqs.DeleteMessageBatchInput{
		Entries: []types.DeleteMessageBatchRequestEntry{
			{
				Id:            &testId1,
				ReceiptHandle: &receiptHandle1,
			},
			{
				Id:            &testId1,
				ReceiptHandle: &receiptHandle2,
			},
			{
				Id:            &testId1,
				ReceiptHandle: &receiptHandle3,
			},
		},
		QueueUrl: crateQueueResponse.QueueUrl,
	})

	assert.Contains(t, error.Error(), "400")
	assert.Contains(t, error.Error(), "AWS.SimpleQueueService.BatchEntryIdsNotDistinct")
	assert.Contains(t, error.Error(), "Two or more batch entries in the request have the same Id.")
}

func Test_DeleteMessageBatchV1_json_success_all_delete(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	// create queue
	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	testId1 := "test1"
	testId2 := "test2"
	testId3 := "test3"
	messageBody1 := "test%20message%20body%201"
	messageBody2 := "test%20message%20body%202"
	messageBody3 := "test%20message%20body%203"

	// send messages
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

	receiveMessageOutput, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            createQueueResponse.QueueUrl,
		MaxNumberOfMessages: 10,
	})

	// delete messages
	deleteMessageBatchOutput, error := sqsClient.DeleteMessageBatch(context.TODO(), &sqs.DeleteMessageBatchInput{
		Entries: []types.DeleteMessageBatchRequestEntry{
			{
				Id:            &testId1,
				ReceiptHandle: receiveMessageOutput.Messages[0].ReceiptHandle,
			},
			{
				Id:            &testId2,
				ReceiptHandle: receiveMessageOutput.Messages[1].ReceiptHandle,
			},
			{
				Id:            &testId3,
				ReceiptHandle: receiveMessageOutput.Messages[2].ReceiptHandle,
			},
		},
		QueueUrl: createQueueResponse.QueueUrl,
	})

	// received no message
	receiveMessageOutput2, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            createQueueResponse.QueueUrl,
		MaxNumberOfMessages: 10,
	})

	assert.Nil(t, error)
	assert.Empty(t, deleteMessageBatchOutput.Failed)
	assert.Equal(t, &testId1, deleteMessageBatchOutput.Successful[0].Id)
	assert.Equal(t, &testId2, deleteMessageBatchOutput.Successful[1].Id)
	assert.Equal(t, &testId3, deleteMessageBatchOutput.Successful[2].Id)
	assert.Empty(t, receiveMessageOutput2.Messages)
}

func Test_DeleteMessageBatchV1_json_success_not_found_message(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	// create queue
	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	dummyQueue := "dummy"
	// create dummy queue
	createQueueResponse2, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &dummyQueue,
	})

	testId1 := "test1"
	testId2 := "test2"
	testId3 := "test3"
	messageBody1 := "test%20message%20body%201"
	messageBody2 := "test%20message%20body%202"
	messageBody3 := "test%20message%20body%203"

	// send messages
	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: &messageBody1,
		QueueUrl:    createQueueResponse.QueueUrl,
	})
	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: &messageBody2,
		QueueUrl:    createQueueResponse2.QueueUrl,
	})
	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: &messageBody3,
		QueueUrl:    createQueueResponse.QueueUrl,
	})

	receiveMessageOutput, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            createQueueResponse.QueueUrl,
		MaxNumberOfMessages: 10,
	})

	receiveMessageOutput2, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            createQueueResponse2.QueueUrl,
		MaxNumberOfMessages: 10,
	})

	// delete messages
	deleteMessageBatchOutput, error := sqsClient.DeleteMessageBatch(context.TODO(), &sqs.DeleteMessageBatchInput{
		Entries: []types.DeleteMessageBatchRequestEntry{
			{
				Id:            &testId1,
				ReceiptHandle: receiveMessageOutput.Messages[0].ReceiptHandle,
			},
			{
				Id:            &testId2,
				ReceiptHandle: receiveMessageOutput2.Messages[0].ReceiptHandle,
			},
			{
				Id:            &testId3,
				ReceiptHandle: receiveMessageOutput.Messages[1].ReceiptHandle,
			},
		},
		QueueUrl: createQueueResponse.QueueUrl,
	})

	// received no message
	receiveMessageOutput3, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            createQueueResponse.QueueUrl,
		MaxNumberOfMessages: 10,
	})

	// not error
	assert.Nil(t, error)

	// deleted messages
	assert.Equal(t, &testId1, deleteMessageBatchOutput.Successful[0].Id)
	assert.Equal(t, &testId3, deleteMessageBatchOutput.Successful[1].Id)

	failedMessage := "Message not found"
	// not founded message
	assert.Equal(t, &testId2, deleteMessageBatchOutput.Failed[0].Id)
	assert.Equal(t, &failedMessage, deleteMessageBatchOutput.Failed[0].Message)

	// confirm no message
	assert.Empty(t, receiveMessageOutput3.Messages)

}

func Test_DeleteMessageBatchV1_xml_success_not_found_message(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	// create queue
	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	dummyQueue := "dummy"
	// create dummy queue
	createQueueResponse2, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &dummyQueue,
	})

	testId1 := "test1"
	testId2 := "test2"
	testId3 := "test3"

	messageBody1 := "test%20message%20body%201"
	messageBody2 := "test%20message%20body%202"
	messageBody3 := "test%20message%20body%203"

	type SendMessageRequestBodyXML struct {
		Action      string `xml:"Action"`
		Version     string `xml:"Version"`
		QueueUrl    string `xml:"QueueUrl"`
		MessageBody string `xml:"MessageBody"`
	}

	sendMessageRequest1 := SendMessageRequestBodyXML{
		Action:      "SendMessage",
		QueueUrl:    *createQueueResponse.QueueUrl,
		Version:     "2012-11-05",
		MessageBody: messageBody1,
	}
	sendMessageRequest2 := SendMessageRequestBodyXML{
		Action:      "SendMessage",
		QueueUrl:    *createQueueResponse2.QueueUrl,
		Version:     "2012-11-05",
		MessageBody: messageBody2,
	}
	sendMessageRequest3 := SendMessageRequestBodyXML{
		Action:      "SendMessage",
		QueueUrl:    *createQueueResponse.QueueUrl,
		Version:     "2012-11-05",
		MessageBody: messageBody3,
	}

	// send messages
	e := httpexpect.Default(t, server.URL)
	e.POST("/").
		WithForm(sendMessageRequest1).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	e.POST("/").
		WithForm(sendMessageRequest2).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	e.POST("/").
		WithForm(sendMessageRequest3).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	type ReceiveMessageRequestBodyXML struct {
		Action              string `xml:"Action"`
		QueueUrl            string `xml:"QueueUrl"`
		Version             string `xml:"Version"`
		MaxNumberOfMessages string `xml:"MaxNumberOfMessages"`
	}

	receiveMessageRequestBodyXML1 := ReceiveMessageRequestBodyXML{
		Action:              "ReceiveMessage",
		QueueUrl:            *createQueueResponse.QueueUrl,
		Version:             "2012-11-05",
		MaxNumberOfMessages: "10",
	}

	receiveMessageRequestBodyXML2 := ReceiveMessageRequestBodyXML{
		Action:              "ReceiveMessage",
		QueueUrl:            *createQueueResponse2.QueueUrl,
		Version:             "2012-11-05",
		MaxNumberOfMessages: "10",
	}

	// received messages
	receivedMessages1 := e.POST("/").
		WithForm(receiveMessageRequestBodyXML1).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	receivedMessageResponse1 := models.ReceiveMessageResponse{}
	xml.Unmarshal([]byte(receivedMessages1), &receivedMessageResponse1)

	// received messages
	receivedMessages2 := e.POST("/").
		WithForm(receiveMessageRequestBodyXML2).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	receivedMessageResponse2 := models.ReceiveMessageResponse{}
	xml.Unmarshal([]byte(receivedMessages2), &receivedMessageResponse2)

	deleteMessageBatchRequestBodyXML := struct {
		Action   string `xml:"Action"`
		QueueUrl string `xml:"QueueUrl"`
		Version  string `xml:"Version"`
	}{
		Action:   "DeleteMessageBatch",
		QueueUrl: *createQueueResponse.QueueUrl,
		Version:  "2012-11-05",
	}

	// delete messages
	deletedMessages := e.POST("/").
		WithForm(deleteMessageBatchRequestBodyXML).
		WithFormField("Entries.0.Id", testId1).
		WithFormField("Entries.0.ReceiptHandle", receivedMessageResponse1.Result.Messages[0].ReceiptHandle).
		WithFormField("Entries.1.Id", testId2).
		WithFormField("Entries.1.ReceiptHandle", receivedMessageResponse2.Result.Messages[0].ReceiptHandle).
		WithFormField("Entries.2.Id", testId3).
		WithFormField("Entries.2.ReceiptHandle", receivedMessageResponse1.Result.Messages[1].ReceiptHandle).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	deleteMessageBatchResponse := models.DeleteMessageBatchResponse{}
	xml.Unmarshal([]byte(deletedMessages), &deleteMessageBatchResponse)

	// confirm no message
	receivedMessages3 := e.POST("/").
		WithForm(receiveMessageRequestBodyXML1).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	receivedMessageResponse3 := models.ReceiveMessageResponse{}
	xml.Unmarshal([]byte(receivedMessages3), &receivedMessageResponse3)

	// success: delete messages
	assert.Contains(t, deleteMessageBatchResponse.Result.Successful[0].Id, testId1)
	assert.Contains(t, deleteMessageBatchResponse.Result.Successful[1].Id, testId3)

	failedMessage := "Message not found"
	// not founded message
	assert.NotEmpty(t, deleteMessageBatchResponse.Result.Failed)
	assert.Contains(t, deleteMessageBatchResponse.Result.Failed[0].Id, testId2)
	assert.Contains(t, deleteMessageBatchResponse.Result.Failed[0].Message, failedMessage)

	// confirm no message
	assert.Empty(t, receivedMessageResponse3.Result.Messages)
}

func Test_DeleteMessageBatchV1_xml_success_all_deletes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	// create queue
	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	testId1 := "test1"
	testId2 := "test2"
	testId3 := "test3"

	messageBody1 := "test%20message%20body%201"
	messageBody2 := "test%20message%20body%202"
	messageBody3 := "test%20message%20body%203"

	type SendMessageRequestBodyXML struct {
		Action      string `xml:"Action"`
		Version     string `xml:"Version"`
		QueueUrl    string `xml:"QueueUrl"`
		MessageBody string `xml:"MessageBody"`
	}

	sendMessageRequest1 := SendMessageRequestBodyXML{
		Action:      "SendMessage",
		QueueUrl:    *createQueueResponse.QueueUrl,
		Version:     "2012-11-05",
		MessageBody: messageBody1,
	}
	sendMessageRequest2 := SendMessageRequestBodyXML{
		Action:      "SendMessage",
		QueueUrl:    *createQueueResponse.QueueUrl,
		Version:     "2012-11-05",
		MessageBody: messageBody2,
	}
	sendMessageRequest3 := SendMessageRequestBodyXML{
		Action:      "SendMessage",
		QueueUrl:    *createQueueResponse.QueueUrl,
		Version:     "2012-11-05",
		MessageBody: messageBody3,
	}

	// send messages
	e := httpexpect.Default(t, server.URL)
	e.POST("/").
		WithForm(sendMessageRequest1).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	e.POST("/").
		WithForm(sendMessageRequest2).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	e.POST("/").
		WithForm(sendMessageRequest3).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	var ReceiveMessageRequestBodyXML = struct {
		Action              string `xml:"Action"`
		QueueUrl            string `xml:"QueueUrl"`
		Version             string `xml:"Version"`
		MaxNumberOfMessages string `xml:"MaxNumberOfMessages"`
	}{
		Action:              "ReceiveMessage",
		QueueUrl:            *createQueueResponse.QueueUrl,
		Version:             "2012-11-05",
		MaxNumberOfMessages: "10",
	}

	// received messages
	receivedMessages := e.POST("/").
		WithForm(ReceiveMessageRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	receivedMessageResponse := models.ReceiveMessageResponse{}
	xml.Unmarshal([]byte(receivedMessages), &receivedMessageResponse)

	deleteMessageBatchRequestBodyXML := struct {
		Action   string `xml:"Action"`
		QueueUrl string `xml:"QueueUrl"`
		Version  string `xml:"Version"`
	}{
		Action:   "DeleteMessageBatch",
		QueueUrl: *createQueueResponse.QueueUrl,
		Version:  "2012-11-05",
	}

	// delete messages
	deletedMessages := e.POST("/").
		WithForm(deleteMessageBatchRequestBodyXML).
		WithFormField("Entries.0.Id", testId1).
		WithFormField("Entries.0.ReceiptHandle", receivedMessageResponse.Result.Messages[0].ReceiptHandle).
		WithFormField("Entries.1.Id", testId2).
		WithFormField("Entries.1.ReceiptHandle", receivedMessageResponse.Result.Messages[1].ReceiptHandle).
		WithFormField("Entries.2.Id", testId3).
		WithFormField("Entries.2.ReceiptHandle", receivedMessageResponse.Result.Messages[2].ReceiptHandle).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	deleteMessageBatchResponse := models.DeleteMessageBatchResponse{}
	xml.Unmarshal([]byte(deletedMessages), &deleteMessageBatchResponse)

	// confirm no message
	receivedMessages2 := e.POST("/").
		WithForm(ReceiveMessageRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	receivedMessageResponse2 := models.ReceiveMessageResponse{}
	xml.Unmarshal([]byte(receivedMessages2), &receivedMessageResponse)

	// check no error
	assert.Empty(t, deleteMessageBatchResponse.Result.Failed)

	// checked delete message batch resoponse
	assert.Contains(t, deleteMessageBatchResponse.Result.Successful[0].Id, testId1)
	assert.Contains(t, deleteMessageBatchResponse.Result.Successful[1].Id, testId2)
	assert.Contains(t, deleteMessageBatchResponse.Result.Successful[2].Id, testId3)

	// confirm no message
	assert.Empty(t, receivedMessageResponse2)
}
