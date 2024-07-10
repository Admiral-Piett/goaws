package gosns

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	log "github.com/sirupsen/logrus"
)

func UnsubscribeV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewUnsubscribeRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - UnsubscribeV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	log.Infof("Unsubscribe: %s", requestBody.SubscriptionArn)
	for _, topic := range app.SyncTopics.Topics {
		for i, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == requestBody.SubscriptionArn {
				app.SyncTopics.Lock()

				copy(topic.Subscriptions[i:], topic.Subscriptions[i+1:])
				topic.Subscriptions[len(topic.Subscriptions)-1] = nil
				topic.Subscriptions = topic.Subscriptions[:len(topic.Subscriptions)-1]

				app.SyncTopics.Unlock()

				respStruct := models.UnsubscribeResponse{
					Xmlns:    models.BASE_XMLNS,
					Metadata: app.ResponseMetadata{RequestId: uuid.NewString()},
				}
				return http.StatusOK, respStruct
			}
		}
	}
	return utils.CreateErrorResponseV1("SubscriptionNotFound", false)
}
