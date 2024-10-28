package gosqs

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app/conf"

	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestPurgeQueueV1_success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PurgeQueueRequest)
		*v = models.PurgeQueueRequest{
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}

	// Put a message on the queue
	targetQueue := models.SyncQueues.Queues["unit-queue1"]
	models.SyncQueues.Lock()
	targetQueue.Messages = []models.SqsMessage{models.SqsMessage{}}
	targetQueue.Duplicates = map[string]time.Time{
		"dedupe-id": time.Now(),
	}
	models.SyncQueues.Unlock()

	expectedResponse := models.PurgeQueueResponse{
		Xmlns:    models.BaseXmlns,
		Metadata: models.BaseResponseMetadata,
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := PurgeQueueV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResponse, response)

	assert.Nil(t, targetQueue.Messages)
	assert.Equal(t, map[string]time.Time{}, targetQueue.Duplicates)
}

func TestPurgeQueueV1_success_no_messages_on_queue(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PurgeQueueRequest)
		*v = models.PurgeQueueRequest{
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}

	expectedResponse := models.PurgeQueueResponse{
		Xmlns:    models.BaseXmlns,
		Metadata: models.BaseResponseMetadata,
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := PurgeQueueV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResponse, response)

	targetQueue := models.SyncQueues.Queues["unit-queue1"]
	assert.Nil(t, targetQueue.Messages)
	assert.Equal(t, map[string]time.Time{}, targetQueue.Duplicates)
}

func TestPurgeQueueV1_request_transformer_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PurgeQueueV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestPurgeQueueV1_requested_queue_does_not_exist(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PurgeQueueRequest)
		*v = models.PurgeQueueRequest{
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "garbage"),
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PurgeQueueV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
