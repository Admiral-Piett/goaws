package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

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
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *createQueueResponse.QueueUrl)

	e.POST("/queue/new-queue-1").
		WithForm(sf.SendMessageRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	receiveMessageResponse, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: createQueueResponse.QueueUrl,
	})

	assert.Nil(t, err)

	assert.Equal(t, 1, len(receiveMessageResponse.Messages))
	assert.Equal(t, sf.SendMessageRequestBodyXML.MessageBody, *receiveMessageResponse.Messages[0].Body)
}

func Test_ReceiveMessageV1_json_while_concurrent_delete(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
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
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *createQueueResponse.QueueUrl)

	e.POST("/queue/new-queue-1").
		WithForm(sf.SendMessageRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	r := e.POST("/queue/new-queue-1").
		WithForm(sf.ReceiveMessageRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	var r1 models.ReceiveMessageResponse
	xml.Unmarshal([]byte(r), &r1)

	assert.Equal(t, 1, len(r1.Result.Messages))
	assert.Equal(t, sf.SendMessageRequestBodyXML.MessageBody, string(r1.Result.Messages[0].Body))
}
