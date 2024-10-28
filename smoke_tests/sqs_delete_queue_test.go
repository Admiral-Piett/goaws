package smoke_tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/gavv/httpexpect/v2"

	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/stretchr/testify/assert"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
)

func Test_DeleteQueueV1_json(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	qName := "unit-queue1"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &qName,
	})

	qUrl := fmt.Sprintf("%s/%s", af.BASE_URL, qName)
	_, err := sqsClient.DeleteQueue(context.TODO(), &sqs.DeleteQueueInput{
		QueueUrl: &qUrl,
	})

	assert.Nil(t, err)

	models.SyncQueues.Lock()
	defer models.SyncQueues.Unlock()

	targetQueue, ok := models.SyncQueues.Queues["unit-queue1"]
	assert.False(t, ok)
	assert.Nil(t, targetQueue)
}

func Test_DeleteQueueV1_xml(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		models.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	qName := "unit-queue1"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &qName,
	})

	qUrl := fmt.Sprintf("%s/%s", af.BASE_URL, qName)

	e.POST("/").
		WithForm(struct {
			Action   string `xml:"Action"`
			QueueUrl string `xml:"QueueUrl"`
		}{
			Action:   "DeleteQueue",
			QueueUrl: qUrl,
		}).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	models.SyncQueues.Lock()
	defer models.SyncQueues.Unlock()

	targetQueue, ok := models.SyncQueues.Queues["unit-queue1"]
	assert.False(t, ok)
	assert.Nil(t, targetQueue)
}
