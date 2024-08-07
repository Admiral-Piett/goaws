package gosns

import (
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/fixtures"

	"github.com/Admiral-Piett/goaws/app/conf"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestUnsubscribeV1_success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	subArn := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0].SubscriptionArn

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.UnsubscribeRequest)
		*v = models.UnsubscribeRequest{
			SubscriptionArn: subArn,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := UnsubscribeV1(r)

	assert.Equal(t, http.StatusOK, status)
	_, ok := response.(models.UnsubscribeResponse)

	subs := app.SyncTopics.Topics["unit-topic1"].Subscriptions
	assert.Len(t, subs, 0)
	assert.True(t, ok)
}

func TestUnsubscribeV1_invalid_request_body(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := UnsubscribeV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}

func TestUnsubscribeV1_invalid_subscription_arn(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.UnsubscribeRequest)
		*v = models.UnsubscribeRequest{
			SubscriptionArn: "garbage",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := UnsubscribeV1(r)

	assert.Equal(t, http.StatusNotFound, status)
}
