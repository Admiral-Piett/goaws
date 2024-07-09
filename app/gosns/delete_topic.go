package gosns

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/common"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	log "github.com/sirupsen/logrus"
)

func DeleteTopicV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewDeleteTopicRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - DeleteTopicV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	topicArn := requestBody.TopicArn
	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]

	log.Info("Delete Topic - TopicName:", topicName)

	_, ok = app.SyncTopics.Topics[topicName]

	if !ok {
		return utils.CreateErrorResponseV1("TopicNotFound", false)
	}

	app.SyncTopics.Lock()
	delete(app.SyncTopics.Topics, topicName)
	app.SyncTopics.Unlock()
	uuid, _ := common.NewUUID()
	respStruct := models.DeleteTopicResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Metadata: app.ResponseMetadata{RequestId: uuid},
	}

	return http.StatusOK, respStruct

}
