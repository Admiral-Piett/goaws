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

func DeleteMessageBatchV1(req *http.Request) (int, interfaces.AbstractResponseBody) {

	requestBody := models.NewDeleteMessageBatchRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)

	if !ok {
		log.Error("Invalid Request - DeleteMessageBatchV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	queueUrl := requestBody.QueueUrl

	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	if _, ok := app.SyncQueues.Queues[queueName]; !ok {
		return utils.CreateErrorResponseV1("QueueNotFound", true)
	}

	if len(requestBody.Entries) == 0 {
		return utils.CreateErrorResponseV1("EmptyBatchRequest", true)
	}

	if len(requestBody.Entries) > 10 {
		return utils.CreateErrorResponseV1("TooManyEntriesInBatchRequest", true)
	}

	ids := map[string]struct{}{}
	for _, v := range requestBody.Entries {
		if _, ok := ids[v.Id]; ok {
			return utils.CreateErrorResponseV1("BatchEntryIdsNotDistinct", true)
		}
		ids[v.Id] = struct{}{}
	}

	app.SyncQueues.Lock()
	defer app.SyncQueues.Unlock()

	// create deleteMessageMap
	deleteMessageMap := make(map[string]*deleteEntry)
	for _, entry := range requestBody.Entries {
		deleteMessageMap[entry.ReceiptHandle] = &deleteEntry{
			Id:            entry.Id,
			ReceiptHandle: entry.ReceiptHandle,
			Deleted:       false,
		}
	}

	deletedEntries := make([]models.DeleteMessageBatchResultEntry, 0)
	// create a slice to hold messages that are not deleted
	remainingMessages := make([]app.Message, 0, len(app.SyncQueues.Queues[queueName].Messages))

	// delete message from queue
	for _, message := range app.SyncQueues.Queues[queueName].Messages {
		if deleteEntry, found := deleteMessageMap[message.ReceiptHandle]; found {
			// Unlock messages for the group
			log.Debugf("FIFO Queue %s unlocking group %s:", queueName, message.GroupID)
			app.SyncQueues.Queues[queueName].UnlockGroup(message.GroupID)
			delete(app.SyncQueues.Queues[queueName].Duplicates, message.DeduplicationID)
			deleteEntry.Deleted = true
			deletedEntries = append(deletedEntries, models.DeleteMessageBatchResultEntry{Id: deleteEntry.Id})
		} else {
			remainingMessages = append(remainingMessages, message)
		}
	}

	// Update the queue with the remaining mesages
	app.SyncQueues.Queues[queueName].Messages = remainingMessages

	// Process not found entries
	notFoundEntries := make([]models.BatchResultErrorEntry, 0)
	for _, deleteEntry := range deleteMessageMap {
		if !deleteEntry.Deleted {
			notFoundEntries = append(notFoundEntries, models.BatchResultErrorEntry{
				Code:        "1",
				Id:          deleteEntry.Id,
				Message:     "Message not found",
				SenderFault: true,
			})
		}
	}

	respStruct := models.DeleteMessageBatchResponse{
		Xmlns: models.BASE_XMLNS,
		Result: models.DeleteMessageBatchResult{
			Successful: deletedEntries,
			Failed:     notFoundEntries,
		},
		Metadata: models.BASE_RESPONSE_METADATA,
	}

	return http.StatusOK, respStruct

}

type deleteEntry struct {
	Id            string
	ReceiptHandle string
	Error         string
	Deleted       bool
}
