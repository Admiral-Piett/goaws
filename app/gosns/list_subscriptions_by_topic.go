package gosns

import (
	"net/http"
	"strings"

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
	var topic models.Topic

	if value, ok := models.SyncTopics.Topics[topicName]; ok {
		topic = *value
	} else {
		return utils.CreateErrorResponseV1("TopicNotFound", false)
	}

	resultMember := make([]models.TopicMemberResult, 0)

	for _, sub := range topic.Subscriptions {
		tar := models.TopicMemberResult{TopicArn: topic.Arn, Protocol: sub.Protocol,
			SubscriptionArn: sub.SubscriptionArn, Endpoint: sub.EndPoint, Owner: models.CurrentEnvironment.AccountID}
		resultMember = append(resultMember, tar)
	}

	respStruct := models.ListSubscriptionsByTopicResponse{
		Xmlns: models.BaseXmlns,
		Result: models.ListSubscriptionsByTopicResult{
			Subscriptions: models.TopicSubscriptions{
				Member: resultMember,
			},
		},
		Metadata: models.ResponseMetadata{RequestId: uuid.NewString()},
	}
	return http.StatusOK, respStruct

}
