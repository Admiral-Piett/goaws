package gosns

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestListTopicsV1_NoTopics(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "NoQueuesOrTopics")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ListTopicsRequest)
		*v = models.ListTopicsRequest{
			NextToken: "",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := ListTopicsV1(r)

	response, _ := res.(models.ListTopicsResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	assert.NotEqual(t, "", response.Metadata)

	assert.Len(t, response.Result.Topics.Member, 0)
}

func TestListTopicsV1_BaseTopics(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ListTopicsRequest)
		*v = models.ListTopicsRequest{
			NextToken: "",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := ListTopicsV1(r)

	response, _ := res.(models.ListTopicsResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	assert.NotEqual(t, "", response.Metadata)

	assert.Len(t, response.Result.Topics.Member, 4)

	topicArnVisited := map[string]bool{}

	for _, member := range response.Result.Topics.Member {
		_, ok := topicArnVisited[member.TopicArn]
		assert.False(t, ok, fmt.Sprintf("Found duplicated listed arn entry: %s", member.TopicArn))
		topicArnVisited[member.TopicArn] = true
	}
}

func TestListTopicsV1_request_transformer_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := ListTopicsV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
