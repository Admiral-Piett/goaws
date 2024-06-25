package gosqs

import (
	"net/http"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app"
	log "github.com/sirupsen/logrus"
)

func PurgeQueueV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewPurgeQueueRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - PurgeQueueV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	uriSegments := strings.Split(requestBody.QueueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	app.SyncQueues.Lock()
	defer app.SyncQueues.Unlock()
	if _, ok := app.SyncQueues.Queues[queueName]; !ok {
		log.Errorf("Purge Queue: %s, queue does not exist!!!", queueName)
		return utils.CreateErrorResponseV1("QueueNotFound", true)
	}

	log.Infof("Purging Queue: %s", queueName)
	app.SyncQueues.Queues[queueName].Messages = nil
	app.SyncQueues.Queues[queueName].Duplicates = make(map[string]time.Time)

	respStruct := models.PurgeQueueResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
	}
	return http.StatusOK, respStruct
}
