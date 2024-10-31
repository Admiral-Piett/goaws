package gosqs

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func SendMessageBatchV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewSendMessageBatchRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - SendMessageBatchV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	queueUrl := requestBody.QueueUrl

	// TODO: Remove this query param logic if it's not still valid or something
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

	sendEntries := requestBody.Entries

	if len(sendEntries) == 0 {
		return utils.CreateErrorResponseV1("EmptyBatchRequest", true)
	}

	if len(sendEntries) > 10 {
		return utils.CreateErrorResponseV1("TooManyEntriesInBatchRequest", true)
	}
	ids := map[string]struct{}{}
	for _, v := range sendEntries {
		if _, ok := ids[v.Id]; ok {
			return utils.CreateErrorResponseV1("BatchEntryIdsNotDistinct", true)
		}
		ids[v.Id] = struct{}{}
	}

	sentEntries := make([]models.SendMessageBatchResultEntry, 0)
	log.Debug("Putting Message in Queue:", queueName)
	for _, sendEntry := range sendEntries {
		msg := models.SqsMessage{MessageBody: sendEntry.MessageBody}
		if len(sendEntry.MessageAttributes) > 0 {
			msg.MessageAttributes = sendEntry.MessageAttributes
			msg.MD5OfMessageAttributes = utils.HashAttributes(sendEntry.MessageAttributes)
		}
		msg.MD5OfMessageBody = utils.GetMD5Hash(sendEntry.MessageBody)
		msg.GroupID = sendEntry.MessageGroupId
		msg.DeduplicationID = sendEntry.MessageDeduplicationId
		msg.Uuid = uuid.NewString()
		msg.SentTime = time.Now()
		models.SyncQueues.Lock()
		fifoSeqNumber := ""
		if models.SyncQueues.Queues[queueName].IsFIFO {
			fifoSeqNumber = models.SyncQueues.Queues[queueName].NextSequenceNumber(sendEntry.MessageGroupId)
		}

		if !models.SyncQueues.Queues[queueName].IsDuplicate(sendEntry.MessageDeduplicationId) {
			models.SyncQueues.Queues[queueName].Messages = append(models.SyncQueues.Queues[queueName].Messages, msg)
		} else {
			log.Debugf("Message with deduplicationId [%s] in queue [%s] is duplicate ", sendEntry.MessageDeduplicationId, queueName)
		}

		models.SyncQueues.Queues[queueName].InitDuplicatation(sendEntry.MessageDeduplicationId)

		models.SyncQueues.Unlock()
		se := models.SendMessageBatchResultEntry{
			Id:                     sendEntry.Id,
			MessageId:              msg.Uuid,
			MD5OfMessageBody:       msg.MD5OfMessageBody,
			MD5OfMessageAttributes: msg.MD5OfMessageAttributes,
			SequenceNumber:         fifoSeqNumber,
		}
		sentEntries = append(sentEntries, se)
		log.Infof("%s: Queue: %s, Message: %s\n", time.Now().Format("2006-01-02 15:04:05"), queueName, msg.MessageBody)
	}

	respStruct := models.SendMessageBatchResponse{
		Xmlns:    models.BaseXmlns,
		Result:   models.SendMessageBatchResult{Entry: sentEntries},
		Metadata: models.BaseResponseMetadata,
	}

	return http.StatusOK, respStruct

}
