package gosqs

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app/utils"

	log "github.com/sirupsen/logrus"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/common"
	"github.com/gorilla/mux"
)

func SendMessageV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewSendMessageRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req)
	if !ok {
		log.Error("Invalid Request - CreateQueueV1")
		return createErrorResponseV1(ErrInvalidParameterValue.Type)
	}
	messageBody := requestBody.MessageBody
	messageGroupID := requestBody.MessageGroupId
	messageDeduplicationID := requestBody.MessageDeduplicationId
	messageAttributes := extractMessageAttributes(req, "")

	queueUrl := getQueueFromPath(requestBody.QueueUrl, req.URL.String())

	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	if _, ok := app.SyncQueues.Queues[queueName]; !ok {
		// Queue does not exist
		return createErrorResponseV1("QueueNotFound")
	}

	if app.SyncQueues.Queues[queueName].MaximumMessageSize > 0 &&
		len(messageBody) > app.SyncQueues.Queues[queueName].MaximumMessageSize {
		// Message size is too big
		return createErrorResponseV1("MessageTooBig")
	}

	delaySecs := app.SyncQueues.Queues[queueName].DelaySeconds
	if mv := req.FormValue("DelaySeconds"); mv != "" {
		delaySecs, _ = strconv.Atoi(mv)
	}

	log.Println("Putting Message in Queue:", queueName)
	msg := app.Message{MessageBody: []byte(messageBody)}
	if len(messageAttributes) > 0 {
		msg.MessageAttributes = messageAttributes
		msg.MD5OfMessageAttributes = common.HashAttributes(messageAttributes)
	}
	msg.MD5OfMessageBody = common.GetMD5Hash(messageBody)
	msg.Uuid, _ = common.NewUUID()
	msg.GroupID = messageGroupID
	msg.DeduplicationID = messageDeduplicationID
	msg.SentTime = time.Now()
	msg.DelaySecs = delaySecs

	app.SyncQueues.Lock()
	fifoSeqNumber := ""
	if app.SyncQueues.Queues[queueName].IsFIFO {
		fifoSeqNumber = app.SyncQueues.Queues[queueName].NextSequenceNumber(messageGroupID)
	}

	if !app.SyncQueues.Queues[queueName].IsDuplicate(messageDeduplicationID) {
		app.SyncQueues.Queues[queueName].Messages = append(app.SyncQueues.Queues[queueName].Messages, msg)
	} else {
		log.Debugf("Message with deduplicationId [%s] in queue [%s] is duplicate ", messageDeduplicationID, queueName)
	}

	app.SyncQueues.Queues[queueName].InitDuplicatation(messageDeduplicationID)
	app.SyncQueues.Unlock()
	log.Infof("%s: Queue: %s, Message: %s\n", time.Now().Format("2006-01-02 15:04:05"), queueName, msg.MessageBody)

	respStruct := models.SendMessageResponse{
		Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/",
		Result: models.SendMessageResult{
			MD5OfMessageAttributes: msg.MD5OfMessageAttributes,
			MD5OfMessageBody:       msg.MD5OfMessageBody,
			MessageId:              msg.Uuid,
			SequenceNumber:         fifoSeqNumber,
		},
		Metadata: app.ResponseMetadata{
			RequestId: "00000000-0000-0000-0000-000000000000",
		},
	}

	return http.StatusOK, respStruct
}
