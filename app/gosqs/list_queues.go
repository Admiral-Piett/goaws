package gosqs

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	log "github.com/sirupsen/logrus"
)

// TODO - set up MaxResults, NextToken request params
//
//	https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ListQueues.html
func ListQueuesV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewListQueuesRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, true)
	if !ok {
		log.Error("Invalid Request - ListQueuesV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	log.Info("Listing Queues")
	queueUrls := make([]string, 0)
	app.SyncQueues.Lock()
	for _, queue := range app.SyncQueues.Queues {
		if strings.HasPrefix(queue.Name, requestBody.QueueNamePrefix) {
			queueUrls = append(queueUrls, queue.URL)
		}
	}
	app.SyncQueues.Unlock()

	respStruct := models.ListQueuesResponse{
		Xmlns:    models.BASE_XMLNS,
		Metadata: models.BASE_RESPONSE_METADATA,
		Result: models.ListQueuesResult{
			QueueUrls: queueUrls,
		},
	}

	return http.StatusOK, respStruct
}
