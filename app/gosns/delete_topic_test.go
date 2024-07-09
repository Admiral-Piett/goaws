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

func TestDeleteTopicV1_Success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	initial_num_topics := len(app.SyncTopics.Topics)

	topicName1 := "unit-topic1"

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteTopicRequest)
		*v = models.DeleteTopicRequest{
			TopicArn: "arn:aws:sns:region:accountID:" + topicName1,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := DeleteTopicV1(r)

	response, _ := res.(models.DeleteTopicResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	assert.NotEqual(t, "", response.Metadata)

	topics := app.SyncTopics.Topics
	assert.Equal(t, initial_num_topics-1, len(topics))
	_, ok := topics[topicName1]
	assert.False(t, ok)
}

func TestDeleteTopicV1_NotFound(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "NoQueuesOrTopics")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteTopicRequest)
		*v = models.DeleteTopicRequest{
			TopicArn: "asdf",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := DeleteTopicV1(r)
	resp := res.(models.ErrorResponse)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, resp.Result.Type, "Not Found")
}

func TestDeleteTopicV1_request_transformer_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := DeleteTopicV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
