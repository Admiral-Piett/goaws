package smoke_tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/utils"
	sf "github.com/Admiral-Piett/goaws/smoke_tests/fixtures"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ChangeMessageVisibilityV1_json(t *testing.T) {
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

	receiveMessageResponse, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: createQueueResponse.QueueUrl,
	})

	_, err := sqsClient.ChangeMessageVisibility(context.TODO(), &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          createQueueResponse.QueueUrl,
		ReceiptHandle:     receiveMessageResponse.Messages[0].ReceiptHandle,
		VisibilityTimeout: 2,
	})

	if err != nil {
		t.Fatalf("Error changing message visibility: %v", err)
	}
}

func Test_ChangeMessageVisibilityV1_xml(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		utils.ResetResources()
	}()
	ctx := context.Background()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(ctx)
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse, _ := sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *createQueueResponse.QueueUrl)

	e.POST("/queue/new-queue-1").
		WithForm(sf.SendMessageRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	receiveMessageResponse, _ := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl: createQueueResponse.QueueUrl,
	})

	requestBody := sf.ChangeMessageVisibilityRequestBodyXML
	requestBody.ReceiptHandle = *receiveMessageResponse.Messages[0].ReceiptHandle
	e.POST("/queue/new-queue-1").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
}
