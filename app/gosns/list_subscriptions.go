package gosns

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	log "github.com/sirupsen/logrus"
)

func ListSubscriptionsV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewListSubscriptionsRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - ListSubscriptionsV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	log.Debug("Listing Subscriptions")
	requestId := uuid.NewString()
	respStruct := models.ListSubscriptionsResponse{}
	respStruct.Xmlns = models.BaseXmlns
	respStruct.Metadata.RequestId = requestId
	respStruct.Result.Subscriptions.Member = make([]models.TopicMemberResult, 0)

	for _, topic := range models.SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			tar := models.TopicMemberResult{TopicArn: topic.Arn, Protocol: sub.Protocol,
				SubscriptionArn: sub.SubscriptionArn, Endpoint: sub.EndPoint, Owner: models.CurrentEnvironment.AccountID}
			respStruct.Result.Subscriptions.Member = append(respStruct.Result.Subscriptions.Member, tar)
		}
	}

	return http.StatusOK, respStruct
}
