package gosqs

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app/utils"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

func SendMessageV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewSendMessageRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - SendMessageV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}
	messageBody := requestBody.MessageBody
	messageGroupID := requestBody.MessageGroupId
	messageDeduplicationID := requestBody.MessageDeduplicationId

	queueUrl := getQueueFromPath(requestBody.QueueUrl, req.URL.String())

	queueName := ""
	if queueUrl == "" {
		// TODO: Remove this query param logic if it's not still valid or something
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	if _, ok := models.SyncQueues.Queues[queueName]; !ok {
		// Queue does not exist
		return utils.CreateErrorResponseV1("QueueNotFound", true)
	}

	if models.SyncQueues.Queues[queueName].MaximumMessageSize > 0 &&
		len(messageBody) > models.SyncQueues.Queues[queueName].MaximumMessageSize {
		// Message size is too big
		return utils.CreateErrorResponseV1("MessageTooBig", true)
	}

	delaySecs := models.SyncQueues.Queues[queueName].DelaySeconds
	if requestBody.DelaySeconds != 0 {
		delaySecs = requestBody.DelaySeconds
	}

	log.Debugf("Putting Message in Queue: [%s]", queueName)
	msg := models.SqsMessage{MessageBody: messageBody}
	if len(requestBody.MessageAttributes) > 0 {
		msg.MessageAttributes = requestBody.MessageAttributes
		msg.MD5OfMessageAttributes = utils.HashAttributes(requestBody.MessageAttributes)
	}
	msg.MD5OfMessageBody = utils.GetMD5Hash(messageBody)
	msg.Uuid = uuid.NewString()
	msg.GroupID = messageGroupID
	msg.DeduplicationID = messageDeduplicationID
	msg.SentTime = time.Now()
	msg.DelaySecs = delaySecs

	models.SyncQueues.Lock()
	fifoSeqNumber := ""
	if models.SyncQueues.Queues[queueName].IsFIFO {
		fifoSeqNumber = models.SyncQueues.Queues[queueName].NextSequenceNumber(messageGroupID)
	}

	if !models.SyncQueues.Queues[queueName].IsDuplicate(messageDeduplicationID) {
		models.SyncQueues.Queues[queueName].Messages = append(models.SyncQueues.Queues[queueName].Messages, msg)
	} else {
		log.Debugf("Message with deduplicationId [%s] in queue [%s] is duplicate ", messageDeduplicationID, queueName)
	}

	models.SyncQueues.Queues[queueName].InitDuplicatation(messageDeduplicationID)
	models.SyncQueues.Unlock()
	log.Infof("%s: Queue: %s, Message: %s\n", time.Now().Format("2006-01-02 15:04:05"), queueName, msg.MessageBody)

	respStruct := models.SendMessageResponse{
		Xmlns: models.BaseXmlns,
		Result: models.SendMessageResult{
			MD5OfMessageAttributes: msg.MD5OfMessageAttributes,
			MD5OfMessageBody:       msg.MD5OfMessageBody,
			MessageId:              msg.Uuid,
			SequenceNumber:         fifoSeqNumber,
		},
		Metadata: models.BaseResponseMetadata,
	}

	return http.StatusOK, respStruct
}
