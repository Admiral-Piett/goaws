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

func TestListSubscriptionsByTopicV1_Not_Found(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "NoQueuesOrTopics")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ListSubscriptionsByTopicRequest)
		*v = models.ListSubscriptionsByTopicRequest{
			NextToken: "",
			TopicArn:  "not exist arn",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := ListSubscriptionsByTopicV1(r)
	response, _ := res.(models.ListSubscriptionsByTopicResponse)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Empty(t, response.Result.Subscriptions.Member)
}

func TestListSubscriptionsByTopicV1_Transform_Error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "NoQueuesOrTopics")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := ListSubscriptionsByTopicV1(r)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestListSubscriptionsByTopicV1_Success_Multiple_Subscription(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["local-topic1"].Arn
	subscriptions := app.SyncTopics.Topics["local-topic1"].Subscriptions
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ListSubscriptionsByTopicRequest)
		*v = models.ListSubscriptionsByTopicRequest{
			NextToken: "",
			TopicArn:  topicArn,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := ListSubscriptionsByTopicV1(r)
	response, _ := res.(models.ListSubscriptionsByTopicResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, response.Result.Subscriptions.Member, 2)

	expectedMember := []models.TopicMemberResult{
		{
			TopicArn:        subscriptions[0].TopicArn,
			SubscriptionArn: subscriptions[0].SubscriptionArn,
			Protocol:        subscriptions[0].Protocol,
			Owner:           app.CurrentEnvironment.AccountID,
			Endpoint:        subscriptions[0].EndPoint,
		},
		{
			TopicArn:        subscriptions[1].TopicArn,
			SubscriptionArn: subscriptions[1].SubscriptionArn,
			Protocol:        subscriptions[1].Protocol,
			Owner:           app.CurrentEnvironment.AccountID,
			Endpoint:        subscriptions[1].EndPoint,
		},
	}

	assert.ElementsMatch(t, expectedMember, response.Result.Subscriptions.Member)
}
