package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"sync"
	"testing"

	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	sf "github.com/Admiral-Piett/goaws/smoke_tests/fixtures"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ReceiveMessageV1_json(t *testing.T) {
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

	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *createQueueResponse.QueueUrl)

	_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    createQueueResponse.QueueUrl,
		MessageBody: aws.String("MyTestMessage"),
	})
	assert.Nil(t, err)

	receiveMessageResponse, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: createQueueResponse.QueueUrl,
	})

	assert.Nil(t, err)

	assert.Equal(t, 1, len(receiveMessageResponse.Messages))
	assert.Equal(t, "MyTestMessage", *receiveMessageResponse.Messages[0].Body)
	assert.Equal(t, "ad4883a84ad41c79aa3a373698c0d4e9", *receiveMessageResponse.Messages[0].MD5OfBody)
	assert.Equal(t, "", *receiveMessageResponse.Messages[0].MD5OfMessageAttributes)

	assert.NotEmpty(t, receiveMessageResponse.Messages[0].Attributes["ApproximateFirstReceiveTimestamp"])
	assert.NotEmpty(t, receiveMessageResponse.Messages[0].Attributes["SenderId"])
	assert.NotEmpty(t, receiveMessageResponse.Messages[0].Attributes["ApproximateReceiveCount"])
	assert.NotEmpty(t, receiveMessageResponse.Messages[0].Attributes["SentTimestamp"])

	assert.Len(t, receiveMessageResponse.Messages[0].MessageAttributes, 0)
}

func Test_ReceiveMessageV1_json_with_attributes(t *testing.T) {
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

	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *createQueueResponse.QueueUrl)

	_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    createQueueResponse.QueueUrl,
		MessageBody: aws.String("MyTestMessage"),
		MessageAttributes: map[string]sqstypes.MessageAttributeValue{
			"attr1": {
				DataType:    aws.String("String"),
				StringValue: aws.String("string-value"),
			},
			"attr2": {
				DataType:    aws.String("Number"),
				StringValue: aws.String("number-value"),
			},
			"attr3": {
				DataType:    aws.String("Binary"),
				BinaryValue: []byte("binary-value"),
			},
		},
	})
	assert.Nil(t, err)

	receiveMessageResponse, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: createQueueResponse.QueueUrl,
	})

	assert.Nil(t, err)

	assert.Equal(t, 1, len(receiveMessageResponse.Messages))
	assert.Equal(t, "MyTestMessage", *receiveMessageResponse.Messages[0].Body)
	assert.Equal(t, "ad4883a84ad41c79aa3a373698c0d4e9", *receiveMessageResponse.Messages[0].MD5OfBody)
	assert.Equal(t, "ae8770938aee44bc548cf65ac377e3bf", *receiveMessageResponse.Messages[0].MD5OfMessageAttributes)

	assert.NotEmpty(t, receiveMessageResponse.Messages[0].Attributes["ApproximateFirstReceiveTimestamp"])
	assert.NotEmpty(t, receiveMessageResponse.Messages[0].Attributes["SenderId"])
	assert.NotEmpty(t, receiveMessageResponse.Messages[0].Attributes["ApproximateReceiveCount"])
	assert.NotEmpty(t, receiveMessageResponse.Messages[0].Attributes["SentTimestamp"])

	assert.Len(t, receiveMessageResponse.Messages[0].MessageAttributes, 3)
	assert.Equal(t, "String", *receiveMessageResponse.Messages[0].MessageAttributes["attr1"].DataType)
	assert.Equal(t, "string-value", *receiveMessageResponse.Messages[0].MessageAttributes["attr1"].StringValue)
	assert.Equal(t, "Number", *receiveMessageResponse.Messages[0].MessageAttributes["attr2"].DataType)
	assert.Equal(t, "number-value", *receiveMessageResponse.Messages[0].MessageAttributes["attr2"].StringValue)
	assert.Equal(t, "Binary", *receiveMessageResponse.Messages[0].MessageAttributes["attr3"].DataType)
	assert.Equal(t, []uint8("binary-value"), receiveMessageResponse.Messages[0].MessageAttributes["attr3"].BinaryValue)
}

func Test_ReceiveMessageV1_json_while_concurrent_delete(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName:  &af.QueueName,
		Attributes: map[string]string{"ReceiveMessageWaitTimeSeconds": "1"},
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl: createQueueResponse.QueueUrl,
		})
		assert.Contains(t, err.Error(), "AWS.SimpleQueueService.NonExistentQueue")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := sqsClient.DeleteQueue(context.TODO(), &sqs.DeleteQueueInput{
			QueueUrl: createQueueResponse.QueueUrl,
		})
		assert.Nil(t, err)
	}()
	wg.Wait()
}

func Test_ReceiveMessageV1_xml(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *createQueueResponse.QueueUrl)

	_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    createQueueResponse.QueueUrl,
		MessageBody: aws.String("MyTestMessage"),
	})
	assert.Nil(t, err)

	response := e.POST("/queue/new-queue-1").
		WithForm(sf.ReceiveMessageRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	var receiveMessageResponse models.ReceiveMessageResponse
	xml.Unmarshal([]byte(response), &receiveMessageResponse)

	assert.Equal(t, 1, len(receiveMessageResponse.Result.Messages))
	assert.Equal(t, 1, len(receiveMessageResponse.Result.Messages))
	assert.Equal(t, "MyTestMessage", receiveMessageResponse.Result.Messages[0].Body)
	assert.Equal(t, "ad4883a84ad41c79aa3a373698c0d4e9", receiveMessageResponse.Result.Messages[0].MD5OfBody)
	assert.Equal(t, "", receiveMessageResponse.Result.Messages[0].MD5OfMessageAttributes)

	entry := "<Attribute><Name>ApproximateFirstReceiveTimestamp</Name><Value>"
	assert.Contains(t, response, entry)

	entry = "<Attribute><Name>SenderId</Name><Value>"
	assert.Contains(t, response, entry)

	entry = "<Attribute><Name>ApproximateReceiveCount</Name><Value>"
	assert.Contains(t, response, entry)

	entry = "<Attribute><Name>SentTimestamp</Name><Value>"
	assert.Contains(t, response, entry)

	assert.Len(t, receiveMessageResponse.Result.Messages[0].MessageAttributes, 0)
}

func Test_ReceiveMessageV1_xml_with_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *createQueueResponse.QueueUrl)

	_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    createQueueResponse.QueueUrl,
		MessageBody: aws.String("MyTestMessage"),
		MessageAttributes: map[string]sqstypes.MessageAttributeValue{
			"attr1": {
				DataType:    aws.String("String"),
				StringValue: aws.String("string-value"),
			},
			"attr2": {
				DataType:    aws.String("Number"),
				StringValue: aws.String("number-value"),
			},
			"attr3": {
				DataType:    aws.String("Binary"),
				BinaryValue: []byte("binary-value"),
			},
		},
	})
	assert.Nil(t, err)

	response := e.POST("/queue/new-queue-1").
		WithForm(sf.ReceiveMessageRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	var receiveMessageResponse models.ReceiveMessageResponse
	xml.Unmarshal([]byte(response), &receiveMessageResponse)

	assert.Equal(t, 1, len(receiveMessageResponse.Result.Messages))
	assert.Equal(t, 1, len(receiveMessageResponse.Result.Messages))
	assert.Equal(t, "MyTestMessage", receiveMessageResponse.Result.Messages[0].Body)
	assert.Equal(t, "ad4883a84ad41c79aa3a373698c0d4e9", receiveMessageResponse.Result.Messages[0].MD5OfBody)
	assert.Equal(t, "ae8770938aee44bc548cf65ac377e3bf", receiveMessageResponse.Result.Messages[0].MD5OfMessageAttributes)

	entry := "<Attribute><Name>ApproximateFirstReceiveTimestamp</Name><Value>"
	assert.Contains(t, response, entry)

	entry = "<Attribute><Name>SenderId</Name><Value>"
	assert.Contains(t, response, entry)

	entry = "<Attribute><Name>ApproximateReceiveCount</Name><Value>"
	assert.Contains(t, response, entry)

	entry = "<Attribute><Name>SentTimestamp</Name><Value>"
	assert.Contains(t, response, entry)

	entry = "<MessageAttribute><Name>attr1</Name><Value><DataType>String</DataType><StringValue>string-value</StringValue></Value></MessageAttribute>"
	assert.Contains(t, response, entry)

	entry = "<MessageAttribute><Name>attr2</Name><Value><DataType>Number</DataType><StringValue>number-value</StringValue></Value></MessageAttribute>"
	assert.Contains(t, response, entry)

	entry = "<MessageAttribute><Name>attr3</Name><Value><BinaryValue>YmluYXJ5LXZhbHVl</BinaryValue><DataType>Binary</DataType></Value></MessageAttribute>"
	assert.Contains(t, response, entry)
}
