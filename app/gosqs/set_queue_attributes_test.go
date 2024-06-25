package gosqs

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestSetQueueAttributesV1_success_multiple_attributes(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetQueueAttributesRequest)
		*v = fixtures.SetQueueAttributesRequest
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := SetQueueAttributesV1(r)

	expectedResponse := models.SetQueueAttributesResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
	}
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResponse, response)

	actualQueue := app.SyncQueues.Queues["unit-queue1"]
	assert.Equal(t, 5, actualQueue.VisibilityTimeout)
	assert.Equal(t, 4, actualQueue.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 1, actualQueue.DelaySeconds)
	assert.Equal(t, 2, actualQueue.MaximumMessageSize)
	assert.Equal(t, 3, actualQueue.MessageRetentionPeriod)
}

func TestSetQueueAttributesV1_success_single_attribute(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetQueueAttributesRequest)
		*v = models.SetQueueAttributesRequest{
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
			Attributes: models.QueueAttributes{
				VisibilityTimeout: 5,
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := SetQueueAttributesV1(r)

	expectedResponse := models.SetQueueAttributesResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
	}
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResponse, response)

	actualQueue := app.SyncQueues.Queues["unit-queue1"]
	assert.Equal(t, 5, actualQueue.VisibilityTimeout)
	assert.Equal(t, 0, actualQueue.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 0, actualQueue.DelaySeconds)
	assert.Equal(t, 0, actualQueue.MaximumMessageSize)
	assert.Equal(t, 345600, actualQueue.MessageRetentionPeriod)
}

func TestSetQueueAttributesV1_invalid_request_body(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetQueueAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSetQueueAttributesV1_missing_queue_url(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetQueueAttributesRequest)
		*v = models.SetQueueAttributesRequest{
			Attributes: models.QueueAttributes{},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetQueueAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSetQueueAttributesV1_missing_expected_queue(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetQueueAttributesRequest)
		*v = models.SetQueueAttributesRequest{
			QueueUrl: "garbage",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetQueueAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSetQueueAttributesV1_invalid_redrive_queue(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetQueueAttributesRequest)
		*v = models.SetQueueAttributesRequest{
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
			Attributes: models.QueueAttributes{
				RedrivePolicy: models.RedrivePolicy{
					MaxReceiveCount:     100,
					DeadLetterTargetArn: fmt.Sprintf("arn:aws:sqs:us-east-1:100010001000:%s", "garbage"),
				},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetQueueAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
