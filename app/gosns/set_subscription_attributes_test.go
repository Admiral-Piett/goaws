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

func TestSetSubscriptionAttributesV1_success_SetRawMessageDelivery_true(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	sub := localTopic1.Subscriptions[0]
	assert.False(t, sub.Raw)

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: sub.SubscriptionArn,
			AttributeName:   "RawMessageDelivery",
			AttributeValue:  "true",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)

	// Assert SubscriptionAttribute has been updated
	assert.True(t, sub.Raw)
}

func TestSetSubscriptionAttributesV1_success_SetRawMessageDelivery_false(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	sub := localTopic1.Subscriptions[1]
	assert.True(t, sub.Raw)

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: sub.SubscriptionArn,
			AttributeName:   "RawMessageDelivery",
			AttributeValue:  "false",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)

	// Assert SubscriptionAttribute has been updated
	assert.False(t, sub.Raw)
}

func TestSetSubscriptionAttributesV1_success_SetFilterPolicy(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	sub := localTopic1.Subscriptions[0]
	assert.Empty(t, sub.FilterPolicy)

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: sub.SubscriptionArn,
			AttributeName:   "FilterPolicy",
			AttributeValue:  "{\"foo\":[\"bar\"]}",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)

	// Assert SubscriptionAttribute has been updated
	expectedFilterPolicy := make(app.FilterPolicy)
	expectedFilterPolicy["foo"] = []string{"bar"}
	assert.Equal(t, &expectedFilterPolicy, sub.FilterPolicy)
}

func TestSetSubscriptionAttributesV1_error_SetFilterPolicy_invalid(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	sub := localTopic1.Subscriptions[0]
	assert.Empty(t, sub.FilterPolicy)

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: sub.SubscriptionArn,
			AttributeName:   "FilterPolicy",
			AttributeValue:  "Not a json string", // Invalid value
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSetSubscriptionAttributesV1_success_SetDeliveryPolicy(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	sub := localTopic1.Subscriptions[0]

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: sub.SubscriptionArn,
			AttributeName:   "DeliveryPolicy",
			AttributeValue:  "foo",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)
}

func TestSetSubscriptionAttributesV1_success_SetFilterPolicyScope(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	sub := localTopic1.Subscriptions[0]

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: sub.SubscriptionArn,
			AttributeName:   "FilterPolicyScope",
			AttributeValue:  "foo",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)
}

func TestSetSubscriptionAttributesV1_success_SetRedrivePolicy(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	sub := localTopic1.Subscriptions[0]

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: sub.SubscriptionArn,
			AttributeName:   "SubscriptionRoleArn",
			AttributeValue:  "foo",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)
}

func TestSetSubscriptionAttributesV1_success_SetSubscriptionRoleArn(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	sub := localTopic1.Subscriptions[0]

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: sub.SubscriptionArn,
			AttributeName:   "SubscriptionRoleArn",
			AttributeValue:  "foo",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusOK, code)
}

func TestSetSubscriptionAttributesV1_error_InvalidAttribute(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	localTopic1 := app.SyncTopics.Topics["local-topic1"]
	subscriptions := localTopic1.Subscriptions
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: subscriptions[1].SubscriptionArn,
			AttributeName:   "InvalidAttribute",
			AttributeValue:  "foo",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := SetSubscriptionAttributesV1(r)
	errorResult := response.GetResult().(models.ErrorResult)

	expected := models.ErrorResult{
		Type:    "InvalidParameterValue",
		Code:    "AWS.SimpleNotificationService.InvalidParameterValue",
		Message: "An invalid or out-of-range value was supplied for the input parameter.",
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, expected, errorResult)
}

func TestSetSubscriptionAttributesV1_error_NonExistentSubscription(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "NoQueuesOrTopics")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SetSubscriptionAttributesRequest)
		*v = models.SetSubscriptionAttributesRequest{
			SubscriptionArn: "foo",
			AttributeName:   "RawMessageDelivery",
			AttributeValue:  "true",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := SetSubscriptionAttributesV1(r)
	errorResult := response.GetResult().(models.ErrorResult)

	expected := models.ErrorResult{
		Type:    "Not Found",
		Code:    "AWS.SimpleNotificationService.NonExistentSubscription",
		Message: "The specified subscription does not exist for this wsdl version.",
	}
	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, expected, errorResult)
}

func TestSetSubscriptionAttributesV1_error_invalid_request(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SetSubscriptionAttributesV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
