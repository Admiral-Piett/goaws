package gosqs

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app"

	"github.com/stretchr/testify/assert"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
)

func TestDeleteQueueV1_success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteQueueRequest)
		*v = models.DeleteQueueRequest{
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}

	expectedResponse := models.DeleteQueueResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := DeleteQueueV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResponse, response)

	_, ok := app.SyncQueues.Queues["unit-queue1"]
	assert.False(t, ok)
}

func TestDeleteQueueV1_success_unknown_queue(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteQueueRequest)
		*v = models.DeleteQueueRequest{
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unknown-queue1"),
		}
		return true
	}

	expectedResponse := models.DeleteQueueResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := DeleteQueueV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResponse, response)
}

func TestDeleteQueueV1_error_invalid_request(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := DeleteQueueV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
