package gosqs

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	log "github.com/sirupsen/logrus"
)

func SetQueueAttributesV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewSetQueueAttributesRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - GetQueueAttributesV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}
	if requestBody.QueueUrl == "" {
		log.Error("Missing QueueUrl - GetQueueAttributesV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	// NOTE: I tore out the handling for devining the url from a param.  I can't find documentation that
	//  that is valid any longer.
	uriSegments := strings.Split(requestBody.QueueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	log.Infof("Set Queue QueueAttributes: %s", queueName)
	app.SyncQueues.Lock()
	defer app.SyncQueues.Unlock()
	queue, ok := app.SyncQueues.Queues[queueName]
	if !ok {
		log.Warningf("Get Queue URL: %s, queue does not exist!!!", queueName)
		return utils.CreateErrorResponseV1("QueueNotFound", true)
	}
	if err := setQueueAttributesV1(queue, requestBody.Attributes); err != nil {
		return utils.CreateErrorResponseV1(err.Error(), true)
	}

	respStruct := models.SetQueueAttributesResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
	}
	return http.StatusOK, respStruct
}
