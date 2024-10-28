package gosqs

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app/models"

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
	models.SyncQueues.Lock()
	for _, queue := range models.SyncQueues.Queues {
		if strings.HasPrefix(queue.Name, requestBody.QueueNamePrefix) {
			queueUrls = append(queueUrls, queue.URL)
		}
	}
	models.SyncQueues.Unlock()

	respStruct := models.ListQueuesResponse{
		Xmlns:    models.BaseXmlns,
		Metadata: models.BaseResponseMetadata,
		Result: models.ListQueuesResult{
			QueueUrls: queueUrls,
		},
	}

	return http.StatusOK, respStruct
}
