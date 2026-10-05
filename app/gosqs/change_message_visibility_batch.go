package gosqs

import (
	"net/http"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func ChangeMessageVisibilityBatchV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewChangeMessageVisibilityBatchRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - ChangeMessageVisibilityBatchV1")
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

	if _, ok := models.SyncQueues.Queues[queueName]; !ok {
		return utils.CreateErrorResponseV1("QueueNotFound", true)
	}

	if len(requestBody.Entries) == 0 {
		return utils.CreateErrorResponseV1("EmptyBatchRequest", true)
	}

	if len(requestBody.Entries) > 10 {
		return utils.CreateErrorResponseV1("TooManyEntriesInBatchRequest", true)
	}

	ids := map[string]bool{}
	for _, v := range requestBody.Entries {
		if _, found := ids[v.Id]; found {
			return utils.CreateErrorResponseV1("BatchEntryIdsNotDistinct", true)
		}
		ids[v.Id] = true
	}

	for _, entry := range requestBody.Entries {
		if entry.VisibilityTimeout > 43200 {
			return utils.CreateErrorResponseV1("InvalidVisibilityTimeout", true)
		}
	}

	// Build a map of receipt handle -> entry for efficient lookup
	type changeEntry struct {
		Id                string
		ReceiptHandle     string
		VisibilityTimeout int
		Found             bool
	}
	changeMap := make(map[string]*changeEntry)
	for i := range requestBody.Entries {
		e := &requestBody.Entries[i]
		changeMap[e.ReceiptHandle] = &changeEntry{
			Id:                e.Id,
			ReceiptHandle:     e.ReceiptHandle,
			VisibilityTimeout: e.VisibilityTimeout,
			Found:             false,
		}
	}

	models.SyncQueues.Lock()
	queue := models.SyncQueues.Queues[queueName]
	defaultTimeout := queue.VisibilityTimeout
	for i := 0; i < len(queue.Messages); i++ {
		msg := &queue.Messages[i]
		if entry, found := changeMap[msg.ReceiptHandle]; found {
			if entry.VisibilityTimeout == 0 {
				msg.ReceiptTime = time.Now().UTC()
				msg.ReceiptHandle = ""
				msg.VisibilityTimeout = time.Now().Add(time.Duration(defaultTimeout) * time.Second)
				msg.Retry++
				if queue.MaxReceiveCount > 0 &&
					queue.DeadLetterQueue != nil &&
					msg.Retry >= queue.MaxReceiveCount {
					queue.DeadLetterQueue.Messages = append(queue.DeadLetterQueue.Messages, *msg)
					queue.Messages = append(queue.Messages[:i], queue.Messages[i+1:]...)
					i--
				}
			} else {
				msg.VisibilityTimeout = time.Now().Add(time.Duration(entry.VisibilityTimeout) * time.Second)
			}
			entry.Found = true
		}
	}
	models.SyncQueues.Unlock()

	successful := make([]models.ChangeMessageVisibilityBatchResultEntry, 0)
	failed := make([]models.BatchResultErrorEntry, 0)
	for _, entry := range changeMap {
		if entry.Found {
			successful = append(successful, models.ChangeMessageVisibilityBatchResultEntry{Id: entry.Id})
		} else {
			failed = append(failed, models.BatchResultErrorEntry{
				Code:        "1",
				Id:          entry.Id,
				Message:     "Message not in flight",
				SenderFault: true,
			})
		}
	}

	respStruct := models.ChangeMessageVisibilityBatchResponse{
		Xmlns: models.BaseXmlns,
		Result: models.ChangeMessageVisibilityBatchResult{
			Successful: successful,
			Failed:     failed,
		},
		Metadata: models.BaseResponseMetadata,
	}

	return http.StatusOK, respStruct
}
