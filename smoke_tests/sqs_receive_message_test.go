package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
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
		utils.ResetResources()
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

	if err != nil {
		t.Fatalf("Error receiving message: %v", err)
	}

	assert.Equal(t, 1, len(receiveMessageResponse.Messages))
	assert.Equal(t, sf.SendMessageRequestBodyXML.MessageBody, *receiveMessageResponse.Messages[0].Body)
}

func Test_ReceiveMessageV1_xml(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		utils.ResetResources()
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
