package gosqs

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/mitchellh/copystructure"

	"github.com/Admiral-Piett/goaws/app/conf"

	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetQueueAttributesV1_success_all(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetQueueAttributesRequest)
		*v = fixtures.GetQueueAttributesRequest
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := GetQueueAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, fixtures.GetQueueAttributesResponse, response)
}

func TestGetQueueAttributesV1_success_no_request_attrs_returns_all(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetQueueAttributesRequest)
		*v = models.GetQueueAttributesRequest{
			QueueUrl: "unit-queue1",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := GetQueueAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, fixtures.GetQueueAttributesResponse, response)
}

func TestGetQueueAttributesV1_success_all_with_redrive_queue(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetQueueAttributesRequest)
		*v = models.GetQueueAttributesRequest{
			QueueUrl:       "unit-queue2",
			AttributeNames: []string{"All"},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := GetQueueAttributesV1(r)

	dupe, _ := copystructure.Copy(fixtures.GetQueueAttributesResponse)
	expectedResponse, _ := dupe.(models.GetQueueAttributesResponse)
	expectedResponse.Result.Attrs[9].Value = fmt.Sprintf("%s:%s", fixtures.BASE_SQS_ARN, "unit-queue2")
	expectedResponse.Result.Attrs = append(expectedResponse.Result.Attrs,
		models.Attribute{
			Name:  "RedrivePolicy",
			Value: fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, fixtures.BASE_SQS_ARN, "other-queue1"),
		},
	)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResponse, response)
}

func TestGetQueueAttributesV1_success_specific_fields(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetQueueAttributesRequest)
		*v = models.GetQueueAttributesRequest{
			QueueUrl:       fmt.Sprintf("%s/unit-queue1", fixtures.BASE_URL),
			AttributeNames: []string{"DelaySeconds"},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := GetQueueAttributesV1(r)

	expectedResponse := models.GetQueueAttributesResponse{
		Xmlns: models.BASE_XMLNS,
		Result: models.GetQueueAttributesResult{Attrs: []models.Attribute{
			models.Attribute{
				Name:  "DelaySeconds",
				Value: "0",
			},
		}},
		Metadata: models.BASE_RESPONSE_METADATA,
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResponse, response)
}

func TestGetQueueAttributesV1_request_transformer_error(t *testing.T) {
	defer func() {
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := GetQueueAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestGetQueueAttributesV1_missing_queue_url_in_request_returns_error(t *testing.T) {
	defer func() {
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetQueueAttributesRequest)
		*v = models.GetQueueAttributesRequest{
			QueueUrl:       "",
			AttributeNames: []string{},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := GetQueueAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestGetQueueAttributesV1_missing_queue_returns_error(t *testing.T) {
	defer func() {
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetQueueAttributesRequest)
		*v = fixtures.GetQueueAttributesRequest
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := GetQueueAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
