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

func TestGetSubscriptionAttributesV1_NonExistentSubscription(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "NoQueuesOrTopics")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetSubscriptionAttributesRequest)
		*v = models.GetSubscriptionAttributesRequest{
			SubscriptionArn: "hogehoge",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := GetSubscriptionAttributesV1(r)
	errorResult := response.GetResult().(models.ErrorResult)

	expected := models.ErrorResult{
		Type:    "Not Found",
		Code:    "AWS.SimpleNotificationService.NonExistentSubscription",
		Message: "The specified subscription does not exist for this wsdl version.",
	}
	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, expected, errorResult)
}

func TestGetSubscriptionAttributesV1_TransformError(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := GetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestGetSubscriptionAttributesV1_success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	subscriptions := localTopic1.Subscriptions
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.GetSubscriptionAttributesRequest)
		*v = models.GetSubscriptionAttributesRequest{
			// local-queue5
			SubscriptionArn: subscriptions[1].SubscriptionArn,
		}
		return true
	}
	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := GetSubscriptionAttributesV1(r)

	result := response.GetResult().(models.GetSubscriptionAttributesResult)
	assert.Equal(t, http.StatusOK, code)
	expectedAttributes := []models.SubscriptionAttributeEntry{
		{
			Key:   "Owner",
			Value: app.CurrentEnvironment.AccountID,
		},
		{
			Key:   "RawMessageDelivery",
			Value: "true",
		},
		{
			Key:   "TopicArn",
			Value: localTopic1.Arn,
		},
		{
			Key:   "Endpoint",
			Value: subscriptions[1].EndPoint,
		},
		{
			Key:   "PendingConfirmation",
			Value: "false",
		},
		{
			Key:   "ConfirmationWasAuthenticated",
			Value: "true",
		}, {
			Key:   "SubscriptionArn",
			Value: subscriptions[1].SubscriptionArn,
		}, {
			Key:   "Protocol",
			Value: "sqs",
		},
		{
			Key:   "FilterPolicy",
			Value: "{\"foo\":[\"bar\"]}",
		},
	}

	assert.ElementsMatch(t, expectedAttributes, result.Attributes.Entries)
}
