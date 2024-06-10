package gosqs

import (
	"net/http"
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
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - SendMessageV1")
		return createErrorResponseV1(ErrInvalidParameterValue.Type)
	}
	messageBody := requestBody.MessageBody
	messageGroupID := requestBody.MessageGroupId
	messageDeduplicationID := requestBody.MessageDeduplicationId
	messageAttributes := requestBody.MessageAttributes

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
	if requestBody.DelaySeconds != 0 {
		delaySecs = requestBody.DelaySeconds
	}

	log.Debugf("Putting Message in Queue: [%s]", queueName)
	msg := app.Message{MessageBody: []byte(messageBody)}
	if len(messageAttributes) > 0 {
		oldStyleMessageAttributes := convertToOldMessageAttributeValueStructure(messageAttributes)
		msg.MessageAttributes = oldStyleMessageAttributes
		msg.MD5OfMessageAttributes = common.HashAttributes(oldStyleMessageAttributes)
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
		Xmlns: models.BASE_XMLNS,
		Result: models.SendMessageResult{
			MD5OfMessageAttributes: msg.MD5OfMessageAttributes,
			MD5OfMessageBody:       msg.MD5OfMessageBody,
			MessageId:              msg.Uuid,
			SequenceNumber:         fifoSeqNumber,
		},
		Metadata: models.BASE_RESPONSE_METADATA,
	}

	return http.StatusOK, respStruct
}

// TODO:
// Refactor internal model for MessageAttribute between SendMessage and ReceiveMessage
// from app.MessageAttributeValue(old) to models.MessageAttributeValue(new) and remove this temporary function.
func convertToOldMessageAttributeValueStructure(newValues map[string]models.MessageAttributeValue) map[string]app.MessageAttributeValue {
	attributes := make(map[string]app.MessageAttributeValue)

	for name, entry := range newValues {
		// StringListValue and BinaryListValue is currently not implemented
		// Please refer app/gosqs/message_attributes.go
		value := ""
		valueKey := ""
		if entry.StringValue != "" {
			value = entry.StringValue
			valueKey = "StringValue"
		} else if entry.BinaryValue != "" {
			value = entry.BinaryValue
			valueKey = "BinaryValue"
		}
		attributes[name] = app.MessageAttributeValue{
			Name:     name,
			DataType: entry.DataType,
			Value:    value,
			ValueKey: valueKey,
		}
	}

	return attributes
}
