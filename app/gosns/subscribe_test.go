package gosns

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeV1_success_no_attributes(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SubscribeRequest)
		*v = models.SubscribeRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"),
			Endpoint: fmt.Sprintf("%s:%s", fixtures.BASE_URL, "unit-queue2"),
			Protocol: "sqs",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := SubscribeV1(r)

	response, _ := res.(models.SubscribeResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	// We are populating the request id with a new random value on this request
	assert.NotEqual(t, "", response.Metadata)
	assert.NotEqual(t, "", response.Result.SubscriptionArn)

	subscriptions := app.SyncTopics.Topics["unit-topic2"].Subscriptions
	assert.Len(t, subscriptions, 1)

	assert.Equal(t, fmt.Sprintf("%s:%s", fixtures.BASE_URL, "unit-queue2"), subscriptions[0].EndPoint)
	assert.Equal(t, "sqs", subscriptions[0].Protocol)
	assert.False(t, subscriptions[0].Raw)
	assert.Contains(t, subscriptions[0].SubscriptionArn, fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"))
	assert.Equal(t, subscriptions[0].TopicArn, fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"))
}

func TestSubscribeV1_success_with_attributes(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SubscribeRequest)
		*v = models.SubscribeRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"),
			Endpoint: fmt.Sprintf("%s:%s", fixtures.BASE_URL, "unit-queue2"),
			Protocol: "sqs",
			Attributes: models.SubscriptionAttributes{
				FilterPolicy:       app.FilterPolicy{"filter": []string{"policy"}},
				RawMessageDelivery: true,
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := SubscribeV1(r)

	response, _ := res.(models.SubscribeResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	// We are populating the request id with a new random value on this request
	assert.NotEqual(t, "", response.Metadata)
	assert.NotEqual(t, "", response.Result.SubscriptionArn)

	subscriptions := app.SyncTopics.Topics["unit-topic2"].Subscriptions
	assert.Len(t, subscriptions, 1)

	expectedFilterPolicy := app.FilterPolicy{"filter": []string{"policy"}}
	assert.Equal(t, fmt.Sprintf("%s:%s", fixtures.BASE_URL, "unit-queue2"), subscriptions[0].EndPoint)
	assert.Equal(t, &expectedFilterPolicy, subscriptions[0].FilterPolicy)
	assert.Equal(t, "sqs", subscriptions[0].Protocol)
	assert.True(t, subscriptions[0].Raw)
	assert.Contains(t, subscriptions[0].SubscriptionArn, fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"))
	assert.Equal(t, subscriptions[0].TopicArn, fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"))
}

func TestSubscribeV1_success_duplicate_subscription(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SubscribeRequest)
		*v = models.SubscribeRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic1"),
			Endpoint: fmt.Sprintf("%s:%s", fixtures.BASE_SQS_ARN, "subscribed-queue1"),
			Protocol: "sqs",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SubscribeV1(r)

	assert.Equal(t, http.StatusOK, code)

	subscriptions := app.SyncTopics.Topics["unit-topic1"].Subscriptions
	assert.Len(t, subscriptions, 1)

	assert.Equal(t, fmt.Sprintf("%s:%s", fixtures.BASE_SQS_ARN, "subscribed-queue1"), subscriptions[0].EndPoint)
	assert.Equal(t, "sqs", subscriptions[0].Protocol)
	assert.True(t, subscriptions[0].Raw)
	assert.Contains(t, subscriptions[0].SubscriptionArn, fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic1"))
	assert.Equal(t, subscriptions[0].TopicArn, fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic1"))
}

func TestSubscribeV1_error_invalid_request(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SubscribeV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSubscribeV1_error_missing_topic(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SubscribeRequest)
		*v = models.SubscribeRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "garbage"),
			Endpoint: fmt.Sprintf("%s:%s", fixtures.BASE_SQS_ARN, "subscribed-queue1"),
			Protocol: "sqs",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SubscribeV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
