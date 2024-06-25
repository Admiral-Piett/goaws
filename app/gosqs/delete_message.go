package gosqs

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func DeleteMessageV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewDeleteMessageRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - DeleteMessageV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	// Retrieve FormValues required
	receiptHandle := requestBody.ReceiptHandle

	// Retrieve FormValues required
	queueUrl := requestBody.QueueUrl
	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	log.Info("Deleting Message, Queue:", queueName, ", ReceiptHandle:", receiptHandle)

	// Find queue/message with the receipt handle and delete
	app.SyncQueues.Lock()
	defer app.SyncQueues.Unlock()
	if _, ok := app.SyncQueues.Queues[queueName]; ok {
		for i, msg := range app.SyncQueues.Queues[queueName].Messages {
			if msg.ReceiptHandle == receiptHandle {
				// Unlock messages for the group
				log.Debugf("FIFO Queue %s unlocking group %s:", queueName, msg.GroupID)
				app.SyncQueues.Queues[queueName].UnlockGroup(msg.GroupID)
				//Delete message from Q
				app.SyncQueues.Queues[queueName].Messages = append(app.SyncQueues.Queues[queueName].Messages[:i], app.SyncQueues.Queues[queueName].Messages[i+1:]...)
				delete(app.SyncQueues.Queues[queueName].Duplicates, msg.DeduplicationID)

				// Create, encode/xml and send response
				respStruct := models.DeleteMessageResponse{
					Xmlns:    models.BASE_XMLNS,
					Metadata: models.BASE_RESPONSE_METADATA,
				}
				return 200, &respStruct
			}
		}
		log.Warning("Receipt Handle not found")
	} else {
		log.Warning("Queue not found")
	}

	return utils.CreateErrorResponseV1("MessageDoesNotExist", true)
}
