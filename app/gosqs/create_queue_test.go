package gosqs

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
)

func TestCreateQueueV1_success(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		v := resultingStruct.(*models.CreateQueueRequest)
		*v = fixtures.CreateQueueRequest
		return true
	}

	expectedQueue := &app.Queue{
		Name: fixtures.QueueName,
		URL: fmt.Sprintf("http://%s.%s:%s/%s/%s",
			fixtures.LOCAL_ENVIRONMENT.Region,
			fixtures.LOCAL_ENVIRONMENT.Host,
			fixtures.LOCAL_ENVIRONMENT.Port,
			fixtures.LOCAL_ENVIRONMENT.AccountID,
			fixtures.QueueName,
		),
		Arn: fmt.Sprintf("arn:aws:sqs:%s:%s:%s",
			fixtures.LOCAL_ENVIRONMENT.Region,
			fixtures.LOCAL_ENVIRONMENT.AccountID,
			fixtures.QueueName,
		),
		VisibilityTimeout:             5,
		ReceiveMessageWaitTimeSeconds: 4,
		DelaySeconds:                  1,
		MaximumMessageSize:            2,
		MessageRetentionPeriod:        3,
		Duplicates:                    make(map[string]time.Time),
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, response := CreateQueueV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, fixtures.CreateQueueResponse, response)

	actualQueue := app.SyncQueues.Queues[fixtures.QueueName]
	assert.Equal(t, expectedQueue, actualQueue)
}

func TestCreateQueueV1_success_with_redrive_policy(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		dupe, _ := copystructure.Copy(fixtures.CreateQueueRequest)
		c, _ := dupe.(models.CreateQueueRequest)
		c.Attributes.RedrivePolicy = models.RedrivePolicy{
			MaxReceiveCount:     100,
			DeadLetterTargetArn: fmt.Sprintf("arn:aws:sqs:us-east-1:100010001000:%s", fixtures.DeadLetterQueueName),
		}

		v := resultingStruct.(*models.CreateQueueRequest)
		*v = c
		return true
	}

	dlq := &app.Queue{
		Name: fixtures.DeadLetterQueueName,
	}
	app.SyncQueues.Queues[fixtures.DeadLetterQueueName] = dlq

	expectedQueue := &app.Queue{
		Name: fixtures.QueueName,
		URL: fmt.Sprintf("http://%s.%s:%s/%s/%s",
			fixtures.LOCAL_ENVIRONMENT.Region,
			fixtures.LOCAL_ENVIRONMENT.Host,
			fixtures.LOCAL_ENVIRONMENT.Port,
			fixtures.LOCAL_ENVIRONMENT.AccountID,
			fixtures.QueueName,
		),
		Arn: fmt.Sprintf("arn:aws:sqs:%s:%s:%s",
			fixtures.LOCAL_ENVIRONMENT.Region,
			fixtures.LOCAL_ENVIRONMENT.AccountID,
			fixtures.QueueName,
		),
		VisibilityTimeout:             5,
		ReceiveMessageWaitTimeSeconds: 4,
		DelaySeconds:                  1,
		MaximumMessageSize:            2,
		MessageRetentionPeriod:        3,
		DeadLetterQueue:               dlq,
		MaxReceiveCount:               100,
		Duplicates:                    make(map[string]time.Time),
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, response := CreateQueueV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, fixtures.CreateQueueResponse, response)

	actualQueue := app.SyncQueues.Queues[fixtures.QueueName]
	assert.Equal(t, expectedQueue, actualQueue)
}

func TestCreateQueueV1_success_with_existing_queue(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		v := resultingStruct.(*models.CreateQueueRequest)
		*v = fixtures.CreateQueueRequest
		return true
	}

	q := &app.Queue{
		Name: fixtures.QueueName,
	}
	app.SyncQueues.Queues[fixtures.QueueName] = q

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, response := CreateQueueV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, fixtures.CreateQueueResponse, response)

	actualQueue := app.SyncQueues.Queues[fixtures.QueueName]
	assert.Equal(t, q, actualQueue)
}

func TestCreateQueueV1_success_with_no_request_attributes_falls_back_to_default(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		dupe, _ := copystructure.Copy(fixtures.CreateQueueRequest)
		c, _ := dupe.(models.CreateQueueRequest)
		c.Attributes = models.Attributes{}

		v := resultingStruct.(*models.CreateQueueRequest)
		*v = c
		return true
	}

	expectedQueue := &app.Queue{
		Name: fixtures.QueueName,
		URL: fmt.Sprintf("http://%s.%s:%s/%s/%s",
			fixtures.LOCAL_ENVIRONMENT.Region,
			fixtures.LOCAL_ENVIRONMENT.Host,
			fixtures.LOCAL_ENVIRONMENT.Port,
			fixtures.LOCAL_ENVIRONMENT.AccountID,
			fixtures.QueueName,
		),
		Arn: fmt.Sprintf("arn:aws:sqs:%s:%s:%s",
			fixtures.LOCAL_ENVIRONMENT.Region,
			fixtures.LOCAL_ENVIRONMENT.AccountID,
			fixtures.QueueName,
		),
		VisibilityTimeout:             0,
		ReceiveMessageWaitTimeSeconds: 0,
		DelaySeconds:                  0,
		MaximumMessageSize:            0,
		MessageRetentionPeriod:        0,
		Duplicates:                    make(map[string]time.Time),
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, response := CreateQueueV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, fixtures.CreateQueueResponse, response)

	actualQueue := app.SyncQueues.Queues[fixtures.QueueName]
	assert.Equal(t, expectedQueue, actualQueue)
}

func TestCreateQueueV1_success_no_configured_region_for_queue_url(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	app.CurrentEnvironment.Region = ""
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		dupe, _ := copystructure.Copy(fixtures.CreateQueueRequest)
		c, _ := dupe.(models.CreateQueueRequest)
		c.Attributes = models.Attributes{}

		v := resultingStruct.(*models.CreateQueueRequest)
		*v = c
		return true
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := CreateQueueV1(r)

	assert.Equal(t, http.StatusOK, code)

	actualQueue := app.SyncQueues.Queues[fixtures.QueueName]
	assert.Equal(t,
		fmt.Sprintf("http://%s:%s/%s/%s",
			fixtures.LOCAL_ENVIRONMENT.Host,
			fixtures.LOCAL_ENVIRONMENT.Port,
			fixtures.LOCAL_ENVIRONMENT.AccountID,
			fixtures.QueueName,
		),
		actualQueue.URL,
	)
}

func TestCreateQueueV1_request_transformer_error(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		return false
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := CreateQueueV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestCreateQueueV1_invalid_dead_letter_queue_error(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		dupe, _ := copystructure.Copy(fixtures.CreateQueueRequest)
		c, _ := dupe.(models.CreateQueueRequest)
		c.Attributes.RedrivePolicy = models.RedrivePolicy{
			MaxReceiveCount:     100,
			DeadLetterTargetArn: fmt.Sprintf("arn:aws:sqs:us-east-1:100010001000:%s", "garbage"),
		}

		v := resultingStruct.(*models.CreateQueueRequest)
		*v = c
		return true
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := CreateQueueV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
