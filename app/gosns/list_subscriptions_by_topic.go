package gosns

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func ListSubscriptionsByTopicV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewListSubscriptionsByTopicRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - ListSubscriptionsByTopicV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	topicArn := requestBody.TopicArn
	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]
	var topic app.Topic

	if value, ok := app.SyncTopics.Topics[topicName]; ok {
		topic = *value
	} else {
		return utils.CreateErrorResponseV1("TopicNotFound", false)
	}

	resultMember := make([]models.TopicMemberResult, 0)

	for _, sub := range topic.Subscriptions {
		tar := models.TopicMemberResult{TopicArn: topic.Arn, Protocol: sub.Protocol,
			SubscriptionArn: sub.SubscriptionArn, Endpoint: sub.EndPoint, Owner: app.CurrentEnvironment.AccountID}
		resultMember = append(resultMember, tar)
	}

	respStruct := models.ListSubscriptionsByTopicResponse{
		Xmlns: models.BASE_XMLNS,
		Result: models.ListSubscriptionsByTopicResult{
			Subscriptions: models.TopicSubscriptions{
				Member: resultMember,
			},
		},
		Metadata: app.ResponseMetadata{RequestId: uuid.NewString()},
	}
	return http.StatusOK, respStruct

}
