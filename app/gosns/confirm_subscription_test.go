package gosns

import (
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestConfirmSubscriptionV1_Success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		TOPIC_DATA = make(map[string]*pendingConfirm)
	}()

	topicArn := "test-topic-arn"
	confirmToken := "test-token"
	subscriptionArn := "test-sub-arn"

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ConfirmSubscriptionRequest)
		*v = models.ConfirmSubscriptionRequest{
			TopicArn: topicArn,
			Token:    confirmToken,
		}
		return true
	}
	// set pending subscription
	TOPIC_DATA[topicArn] = &pendingConfirm{
		subArn: subscriptionArn,
		token:  confirmToken,
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := ConfirmSubscriptionV1(r)

	result := response.GetResult().(models.ConfirmSubscriptionResult)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, subscriptionArn, result.SubscriptionArn)
}

func TestConfirmSubscriptionV1_NotFoundSubscription(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "NoQueuesOrTopics")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		TOPIC_DATA = make(map[string]*pendingConfirm)
	}()

	topicArn := "test-topic-arn"
	confirmToken := "test-token"

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ConfirmSubscriptionRequest)
		*v = models.ConfirmSubscriptionRequest{
			TopicArn: topicArn,
			Token:    confirmToken,
		}
		return true
	}
	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := ConfirmSubscriptionV1(r)
	result := response.GetResult().(models.ErrorResult)
	assert.Equal(t, http.StatusNotFound, code)
	assert.Contains(t, result.Message, "The specified subscription does not exist for this wsdl version.")
}

func TestConfirmSubscriptionV1_MismatchToken(t *testing.T) {

	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		TOPIC_DATA = make(map[string]*pendingConfirm)
	}()

	topicArn := "test-topic-arn"
	confirmToken := "test-token"
	subscriptionArn := "test-sub-arn"

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ConfirmSubscriptionRequest)
		*v = models.ConfirmSubscriptionRequest{
			TopicArn: topicArn,
			Token:    "dummy",
		}
		return true
	}

	// set dummy subscription
	TOPIC_DATA[topicArn] = &pendingConfirm{
		subArn: subscriptionArn,
		token:  confirmToken,
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := ConfirmSubscriptionV1(r)
	result := response.GetResult().(models.ErrorResult)
	assert.Equal(t, http.StatusNotFound, code)
	assert.Contains(t, result.Message, "The specified subscription does not exist for this wsdl version.")
}

func TestConfirmSubscriptionV1_TransformerError(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		TOPIC_DATA = make(map[string]*pendingConfirm)
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := ConfirmSubscriptionV1(r)
	assert.Equal(t, http.StatusBadRequest, code)
}
