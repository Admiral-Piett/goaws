package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/stretchr/testify/assert"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	sf "github.com/Admiral-Piett/goaws/smoke_tests/fixtures"
	"github.com/gavv/httpexpect/v2"
)

func Test_ListQueues_json_no_queues(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	sdkResponse, err := sqsClient.ListQueues(context.TODO(), &sqs.ListQueuesInput{})

	assert.Nil(t, err)
	assert.Equal(t, []string{}, sdkResponse.QueueUrls)
}

func Test_ListQueues_json_multiple_queues(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	queueName2 := "new-queue-2"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName2,
	})
	queueName3 := "new-queue-3"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName3,
	})

	sdkResponse, err := sqsClient.ListQueues(context.TODO(), &sqs.ListQueuesInput{})

	assert.Nil(t, err)

	assert.Contains(t, sdkResponse.QueueUrls, af.QueueUrl)
	assert.Contains(t, sdkResponse.QueueUrls, fmt.Sprintf("%s/new-queue-2", af.BASE_URL))
	assert.Contains(t, sdkResponse.QueueUrls, fmt.Sprintf("%s/new-queue-3", af.BASE_URL))
}

func Test_ListQueues_json_multiple_queues_with_queue_name_prefix(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	queueName1 := "old-queue-1"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName1,
	})
	queueName2 := "new-queue-2"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName2,
	})
	queueName3 := "new-queue-3"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName3,
	})

	prefix := "old"
	sdkResponse, err := sqsClient.ListQueues(context.TODO(), &sqs.ListQueuesInput{QueueNamePrefix: &prefix})

	assert.Nil(t, err)

	assert.Equal(t, sdkResponse.QueueUrls, []string{fmt.Sprintf("%s/old-queue-1", af.BASE_URL)})
}

func Test_ListQueues_xml_no_queues(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	r := e.POST("/").
		WithForm(sf.ListQueuesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	expected := models.ListQueuesResponse{
		Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/",
		Result: models.ListQueuesResult{
			QueueUrls: []string(nil),
		},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}
	response := models.ListQueuesResponse{}
	xml.Unmarshal([]byte(r), &response)
	assert.Equal(t, expected, response)
}

func Test_ListQueues_xml_multiple_queues(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	queueName2 := "new-queue-2"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName2,
	})
	queueName3 := "new-queue-3"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName3,
	})

	r := e.POST("/").
		WithForm(sf.ListQueuesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	response := models.ListQueuesResponse{}
	xml.Unmarshal([]byte(r), &response)
	assert.Contains(t, response.Result.QueueUrls, af.QueueUrl)
	assert.Contains(t, response.Result.QueueUrls, fmt.Sprintf("%s/new-queue-2", af.BASE_URL))
	assert.Contains(t, response.Result.QueueUrls, fmt.Sprintf("%s/new-queue-3", af.BASE_URL))
}

func Test_ListQueues_xml_multiple_queues_with_queue_name_prefix(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	queueName1 := "old-queue-1"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName1,
	})
	queueName2 := "new-queue-2"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName2,
	})
	queueName3 := "new-queue-3"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName3,
	})

	body := struct {
		Action          string `xml:"Action"`
		Version         string `xml:"Version"`
		QueueNamePrefix string `xml:"QueueNamePrefix"`
	}{
		Action:          "ListQueues",
		Version:         "2012-11-05",
		QueueNamePrefix: "old",
	}
	r := e.POST("/").
		WithForm(body).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	expected := models.ListQueuesResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Result:   models.ListQueuesResult{QueueUrls: []string{fmt.Sprintf("%s/%s", af.BASE_URL, queueName1)}},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}
	response := models.ListQueuesResponse{}
	xml.Unmarshal([]byte(r), &response)
	assert.Equal(t, expected, response)
}
