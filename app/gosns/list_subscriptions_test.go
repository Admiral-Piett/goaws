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

func TestListSubcriptionsV1_NoSubscriptions(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "NoQueuesOrTopics")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ListSubscriptionsRequest)
		*v = models.ListSubscriptionsRequest{
			NextToken: "",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := ListSubscriptionsV1(r)

	response, _ := res.(models.ListSubscriptionsResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	assert.NotEqual(t, "", response.Metadata)

	assert.Len(t, response.Result.Subscriptions.Member, 0)
}

func TestListSubcriptionsV1_MultipleSubscriptions(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ListSubscriptionsRequest)
		*v = models.ListSubscriptionsRequest{
			NextToken: "",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := ListSubscriptionsV1(r)

	response, _ := res.(models.ListSubscriptionsResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	assert.NotEqual(t, "", response.Metadata)

	assert.Len(t, response.Result.Subscriptions.Member, 2)
	assert.NotEqual(t, response.Result.Subscriptions.Member[0].SubscriptionArn, response.Result.Subscriptions.Member[1].SubscriptionArn)
}

func TestListSubscriptionsV1_request_transformer_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := ListSubscriptionsV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}
