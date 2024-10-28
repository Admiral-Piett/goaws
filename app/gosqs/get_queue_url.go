package gosqs

import (
	"net/http"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	log "github.com/sirupsen/logrus"
)

func GetQueueUrlV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewGetQueueUrlRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - GetQueueUrlV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	queueName := requestBody.QueueName
	if _, ok := models.SyncQueues.Queues[queueName]; !ok {
		log.Error("Get Queue URL:", queueName, ", queue does not exist!!!")
		return utils.CreateErrorResponseV1("QueueNotFound", true)
	}

	queue := models.SyncQueues.Queues[queueName]
	log.Debug("Get Queue URL:", queue.Name)

	result := models.GetQueueUrlResult{QueueUrl: queue.URL}
	respStruct := models.GetQueueUrlResponse{
		Xmlns:    models.BaseXmlns,
		Result:   result,
		Metadata: models.BaseResponseMetadata,
	}
	return http.StatusOK, respStruct
}
