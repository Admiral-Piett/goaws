package gosqs

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app"
	log "github.com/sirupsen/logrus"
)

func DeleteQueueV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewDeleteQueueRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - DeleteQueueV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	uriSegments := strings.Split(requestBody.QueueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	log.Infof("Deleting Queue: %s", queueName)

	app.SyncQueues.Lock()
	delete(app.SyncQueues.Queues, queueName)
	app.SyncQueues.Unlock()

	respStruct := models.DeleteQueueResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
	}
	return http.StatusOK, respStruct
}
