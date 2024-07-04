package gosns

import (
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateTopicV1_success(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	targetTopicName := "new-topic-1"
	request_success := models.CreateTopicRequest{
		Name: targetTopicName,
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.CreateTopicRequest)
		*v = request_success
		return true
	}

	// No topic yet
	assert.Equal(t, 0, len(app.SyncTopics.Topics))

	// Request
	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := CreateTopicV1(r)

	// Result
	assert.Equal(t, http.StatusOK, status)
	createTopicResponse, ok := response.(models.CreateTopicResponse)
	assert.True(t, ok)
	assert.Contains(t, createTopicResponse.Result.TopicArn, "arn:aws:sns:")
	assert.Contains(t, createTopicResponse.Result.TopicArn, targetTopicName)
	// 1 topic there
	assert.Equal(t, 1, len(app.SyncTopics.Topics))
}

func TestCreateTopicV1_existant_topic(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	targetTopicName := "new-topic-1"

	// Same topic name with existant topic
	request_success := models.CreateTopicRequest{
		Name: targetTopicName,
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.CreateTopicRequest)
		*v = request_success
		return true
	}

	// Prepare existant topic
	targetTopicArn := "arn:aws:sns:us-east-1:123456789012:" + targetTopicName
	topic := &app.Topic{
		Name: targetTopicName,
		Arn:  targetTopicArn,
	}
	app.SyncTopics.Topics[targetTopicName] = topic
	assert.Equal(t, 1, len(app.SyncTopics.Topics))

	// Reques
	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := CreateTopicV1(r)

	// Result
	assert.Equal(t, http.StatusOK, status)
	createTopicResponse, ok := response.(models.CreateTopicResponse)
	assert.True(t, ok)
	assert.Equal(t, targetTopicArn, createTopicResponse.Result.TopicArn) // Same with existant topic
	// No additional topic
	assert.Equal(t, 1, len(app.SyncTopics.Topics))
}

func TestCreateTopicV1_request_transformer_error(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := CreateTopicV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
