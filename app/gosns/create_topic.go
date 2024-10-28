package gosns

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	log "github.com/sirupsen/logrus"
)

func CreateTopicV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewCreateTopicRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - CreateTopicV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	topicName := requestBody.Name
	topicArn := ""
	if _, ok := models.SyncTopics.Topics[topicName]; ok {
		topicArn = models.SyncTopics.Topics[topicName].Arn
	} else {
		topicArn = fmt.Sprintf("arn:aws:sns:%s:%s:%s", models.CurrentEnvironment.Region, models.CurrentEnvironment.AccountID, topicName)

		log.Info("Creating Topic:", topicName)
		topic := &models.Topic{Name: topicName, Arn: topicArn}
		topic.Subscriptions = make([]*models.Subscription, 0)
		models.SyncTopics.Lock()
		models.SyncTopics.Topics[topicName] = topic
		models.SyncTopics.Unlock()
	}

	uuid := uuid.NewString()
	respStruct := models.CreateTopicResponse{
		Xmlns: models.BaseXmlns,
		Result: models.CreateTopicResult{
			TopicArn: topicArn,
		},
		Metadata: models.ResponseMetadata{RequestId: uuid},
	}

	return http.StatusOK, respStruct
}
