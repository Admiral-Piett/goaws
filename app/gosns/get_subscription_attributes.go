package gosns

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetSubscriptionAttributesV1(req *http.Request) (int, interfaces.AbstractResponseBody) {

	requestBody := models.NewGetSubscriptionAttributesRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)

	if !ok {
		log.Error("Invalid Request - GetSubscriptionAttributesV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	sub := getSubscription(requestBody.SubscriptionArn)
	if sub == nil {
		return utils.CreateErrorResponseV1("SubscriptionNotFound", false)
	}

	entries := make([]models.SubscriptionAttributeEntry, 0, 0)
	entry := models.SubscriptionAttributeEntry{Key: "Owner", Value: app.CurrentEnvironment.AccountID}
	entries = append(entries, entry)
	entry = models.SubscriptionAttributeEntry{Key: "RawMessageDelivery", Value: strconv.FormatBool(sub.Raw)}
	entries = append(entries, entry)
	entry = models.SubscriptionAttributeEntry{Key: "TopicArn", Value: sub.TopicArn}
	entries = append(entries, entry)
	entry = models.SubscriptionAttributeEntry{Key: "Endpoint", Value: sub.EndPoint}
	entries = append(entries, entry)
	entry = models.SubscriptionAttributeEntry{Key: "PendingConfirmation", Value: "false"}
	entries = append(entries, entry)
	entry = models.SubscriptionAttributeEntry{Key: "ConfirmationWasAuthenticated", Value: "true"}
	entries = append(entries, entry)
	entry = models.SubscriptionAttributeEntry{Key: "SubscriptionArn", Value: sub.SubscriptionArn}
	entries = append(entries, entry)
	entry = models.SubscriptionAttributeEntry{Key: "Protocol", Value: sub.Protocol}
	entries = append(entries, entry)

	if sub.FilterPolicy != nil {
		filterPolicyBytes, _ := json.Marshal(sub.FilterPolicy)
		entry = models.SubscriptionAttributeEntry{Key: "FilterPolicy", Value: string(filterPolicyBytes)}
		entries = append(entries, entry)
	}

	result := models.GetSubscriptionAttributesResult{Attributes: models.GetSubscriptionAttributes{Entries: entries}}
	uuid := uuid.NewString()
	respStruct := models.GetSubscriptionAttributesResponse{
		Xmlns:    models.BASE_XMLNS,
		Result:   result,
		Metadata: app.ResponseMetadata{RequestId: uuid}}

	return http.StatusOK, respStruct
}
