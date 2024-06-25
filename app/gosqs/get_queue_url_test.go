package gosqs

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetQueueUrlV1_success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetQueueUrlRequest)
		*v = models.GetQueueUrlRequest{
			QueueName:              "unit-queue1",
			QueueOwnerAWSAccountId: "fugafuga",
		}
		return true
	}

	_, r := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)
	code, response := GetQueueUrlV1(r)

	get_queue_url_response := response.(models.GetQueueUrlResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, get_queue_url_response.Result.QueueUrl, fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"))

}

func TestGetQueueUrlV1_error_no_queue(t *testing.T) {

	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetQueueUrlRequest)
		*v = models.GetQueueUrlRequest{
			QueueName:              "not-exist-unit-queue1",
			QueueOwnerAWSAccountId: "fugafuga",
		}
		return true
	}

	_, r := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)
	code, response := GetQueueUrlV1(r)

	expected := models.ErrorResult{
		Type:    "Not Found",
		Code:    "AWS.SimpleQueueService.NonExistentQueue",
		Message: "The specified queue does not exist for this wsdl version.",
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, response.GetResult().(models.ErrorResult), expected)
}

func TestGetQueueUrlV1_error_request_transformer(t *testing.T) {

	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)
	code, _ := GetQueueUrlV1(r)
	assert.Equal(t, http.StatusBadRequest, code)
}
