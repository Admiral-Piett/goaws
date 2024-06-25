package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/gavv/httpexpect/v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/stretchr/testify/assert"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
)

func Test_PurgeQueueV1_json(t *testing.T) {
	defaultEnvironment := app.CurrentEnvironment
	app.CurrentEnvironment = app.Environment{
		EnableDuplicates: true,
	}
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetApp()
		app.CurrentEnvironment = defaultEnvironment
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	qName := fmt.Sprintf("%s.fifo", af.QueueName)
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &qName,
	})

	messageBody := "test-message"
	dedupeId := "dedupe-id"
	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:               &qName,
		MessageBody:            &messageBody,
		MessageDeduplicationId: &dedupeId,
	})

	sdkResponse, err := sqsClient.PurgeQueue(context.TODO(), &sqs.PurgeQueueInput{
		QueueUrl: &qName,
	})

	assert.Nil(t, err)
	assert.NotNil(t, sdkResponse)

	app.SyncQueues.Lock()
	defer app.SyncQueues.Unlock()
	targetQueue := app.SyncQueues.Queues[qName]
	assert.Nil(t, targetQueue.Messages)
	assert.Equal(t, map[string]time.Time{}, targetQueue.Duplicates)
}

func Test_PurgeQueueV1_xml(t *testing.T) {
	defaultEnvironment := app.CurrentEnvironment
	app.CurrentEnvironment = app.Environment{
		EnableDuplicates: true,
	}
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetApp()
		app.CurrentEnvironment = defaultEnvironment
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	qName := fmt.Sprintf("%s.fifo", af.QueueName)
	sdkResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &qName,
	})

	messageBody := "test-message"
	dedupeId := "dedupe-id"
	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:               &qName,
		MessageBody:            &messageBody,
		MessageDeduplicationId: &dedupeId,
	})

	r := e.POST("/").
		WithForm(struct {
			Action   string `xml:"Action"`
			QueueUrl string `xml:"QueueUrl"`
		}{
			Action:   "PurgeQueue",
			QueueUrl: *sdkResponse.QueueUrl,
		}).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	expected := models.PurgeQueueResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
	}
	response := models.PurgeQueueResponse{}
	xml.Unmarshal([]byte(r), &response)
	assert.Equal(t, expected, response)

	app.SyncQueues.Lock()
	defer app.SyncQueues.Unlock()
	targetQueue := app.SyncQueues.Queues[qName]
	assert.Nil(t, targetQueue.Messages)
	assert.Equal(t, map[string]time.Time{}, targetQueue.Duplicates)
}
