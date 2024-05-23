package gosqs

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestListQueuesV1_success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		v := resultingStruct.(*models.ListQueueRequest)
		*v = models.ListQueueRequest{}
		return true
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, response := ListQueuesV1(r)
	r1 := response.(models.ListQueuesResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Contains(t, r1.Result.QueueUrls, fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"))
	assert.Contains(t, r1.Result.QueueUrls, fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue2"))
	assert.Contains(t, r1.Result.QueueUrls, fmt.Sprintf("%s/%s", fixtures.BASE_URL, "other-queue1"))
}

func TestListQueuesV1_success_no_queues(t *testing.T) {
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		v := resultingStruct.(*models.ListQueueRequest)
		*v = models.ListQueueRequest{}
		return true
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, response := ListQueuesV1(r)
	r1 := response.(models.ListQueuesResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, r1.Result.QueueUrls, []string{})
}

func TestListQueuesV1_success_with_queue_name_prefix(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		v := resultingStruct.(*models.ListQueueRequest)
		*v = models.ListQueueRequest{QueueNamePrefix: "other"}
		return true
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, response := ListQueuesV1(r)
	r1 := response.(models.ListQueuesResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, []string{fmt.Sprintf("%s/%s", fixtures.BASE_URL, "other-queue1")}, r1.Result.QueueUrls)
}

func TestListQueuesV1_success_with_queue_name_prefix_no_matching_queues(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		v := resultingStruct.(*models.ListQueueRequest)
		*v = models.ListQueueRequest{QueueNamePrefix: "garbage"}
		return true
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, response := ListQueuesV1(r)
	r1 := response.(models.ListQueuesResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, []string{}, r1.Result.QueueUrls)
}

func TestListQueuesV1_request_transformer_error(t *testing.T) {
	//conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		utils.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
		return false
	}

	_, r := utils.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := ListQueuesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
