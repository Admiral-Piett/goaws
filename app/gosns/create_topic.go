package gosns

import (
	"fmt"
	"net/http"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/common"
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
	if _, ok := app.SyncTopics.Topics[topicName]; ok {
		topicArn = app.SyncTopics.Topics[topicName].Arn
	} else {
		topicArn = fmt.Sprintf("arn:aws:sns:%s:%s:%s", app.CurrentEnvironment.Region, app.CurrentEnvironment.AccountID, topicName)

		log.Info("Creating Topic:", topicName)
		topic := &app.Topic{Name: topicName, Arn: topicArn}
		topic.Subscriptions = make([]*app.Subscription, 0)
		app.SyncTopics.Lock()
		app.SyncTopics.Topics[topicName] = topic
		app.SyncTopics.Unlock()
	}

	uuid, _ := common.NewUUID()
	respStruct := models.CreateTopicResponse{
		Xmlns: models.BASE_XMLNS,
		Result: models.CreateTopicResult{
			TopicArn: topicArn,
		},
		Metadata: app.ResponseMetadata{RequestId: uuid},
	}

	return http.StatusOK, respStruct
}
