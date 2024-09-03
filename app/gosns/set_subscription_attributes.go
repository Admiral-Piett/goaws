package gosns

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func SetSubscriptionAttributesV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewSetSubscriptionAttributesRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - SetSubscriptionAttributesV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	subsArn := requestBody.SubscriptionArn
	attrName := requestBody.AttributeName
	attrValue := requestBody.AttributeValue

	sub := getSubscription(subsArn)
	if sub == nil {
		return utils.CreateErrorResponseV1("SubscriptionNotFound", false)
	}

	switch attrName {
	case "RawMessageDelivery":
		app.SyncTopics.Lock()
		if attrValue == "true" {
			sub.Raw = true
		} else {
			sub.Raw = false
		}
		app.SyncTopics.Unlock()

	case "FilterPolicy":
		filterPolicy := &app.FilterPolicy{}
		err := json.Unmarshal([]byte(attrValue), filterPolicy)
		if err != nil {
			return utils.CreateErrorResponseV1("InvalidParameterValue", false)
		}
		app.SyncTopics.Lock()
		sub.FilterPolicy = filterPolicy
		app.SyncTopics.Unlock()

	case "DeliveryPolicy", "FilterPolicyScope", "RedrivePolicy", "SubscriptionRoleArn":
		log.Info(fmt.Sprintf("AttributeName [%s] is valid on AWS but it is not implemented.", attrName))

	default:
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	uuid := uuid.NewString()
	respStruct := models.SetSubscriptionAttributesResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: app.ResponseMetadata{RequestId: uuid}}

	return http.StatusOK, respStruct
}
