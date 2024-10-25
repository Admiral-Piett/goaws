package gosns

import (
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestPublishV1_success_sqs(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	topic := app.SyncTopics.Topics["unit-topic1"]
	topicArn := topic.Arn

	message := "{\"IAm\": \"aMessage\"}"
	expectedPublishRequest := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = expectedPublishRequest
		return true
	}

	publishCalledWith := [][]interface{}{}
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		publishCalledWith = append(publishCalledWith, []interface{}{topic, message})
		return "", nil
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := PublishV1(r)

	assert.Equal(t, http.StatusOK, status)
	response, ok := response.(models.PublishResponse)
	assert.True(t, ok)

	assert.Equal(t, []interface{}{topic, &expectedPublishRequest}, publishCalledWith[0])
}

func TestPublishV1_request_transformer_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := PublishV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}

func TestPublishV1_request_missing_topic_arn(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			Message: message,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := PublishV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}

func TestPublishV1_request_missing_message(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			TopicArn: topicArn,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := PublishV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}

func TestPublishV1_request_invalid_topic(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			TopicArn: "garbage",
			Message:  message,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := PublishV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}
