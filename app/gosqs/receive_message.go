package gosqs

import (
	"fmt"
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

// TODO - Admiral-Piett - could we refactor the way we hide messages?  Change data structure to a queue
// organized by "reveal time" or a map with the key being a timestamp of when it could be shown?
// Ordered Map - https://github.com/elliotchance/orderedmap
func ReceiveMessageV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewReceiveMessageRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - ReceiveMessageV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	maxNumberOfMessages := requestBody.MaxNumberOfMessages
	if maxNumberOfMessages == 0 {
		maxNumberOfMessages = 1
	}

	queueName := ""
	if requestBody.QueueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(requestBody.QueueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	if _, ok := models.SyncQueues.Queues[queueName]; !ok {
		return utils.CreateErrorResponseV1("QueueNotFound", true)
	}

	var messages []*models.ResultMessage
	respStruct := models.ReceiveMessageResponse{}

	waitTimeSeconds := requestBody.WaitTimeSeconds
	if waitTimeSeconds == 0 {
		models.SyncQueues.RLock()
		waitTimeSeconds = models.SyncQueues.Queues[queueName].ReceiveMessageWaitTimeSeconds
		models.SyncQueues.RUnlock()
	}

	loops := waitTimeSeconds * 10
	for loops > 0 {
		models.SyncQueues.RLock()
		_, queueFound := models.SyncQueues.Queues[queueName]
		if !queueFound {
			models.SyncQueues.RUnlock()
			return utils.CreateErrorResponseV1("QueueNotFound", true)
		}
		messageFound := len(models.SyncQueues.Queues[queueName].Messages)-numberOfHiddenMessagesInQueue(*models.SyncQueues.Queues[queueName]) != 0
		models.SyncQueues.RUnlock()
		if !messageFound {
			continueTimer := time.NewTimer(100 * time.Millisecond)
			select {
			case <-req.Context().Done():
				continueTimer.Stop()
				return http.StatusOK, models.ReceiveMessageResponse{
					Xmlns:    models.BaseXmlns,
					Result:   models.ReceiveMessageResult{},
					Metadata: models.BaseResponseMetadata,
				}
			case <-continueTimer.C:
				continueTimer.Stop()
			}
			loops--
		} else {
			break
		}

	}
	log.Debugf("Getting Message from Queue:%s", queueName)

	models.SyncQueues.Lock()         // Lock the Queues
	defer models.SyncQueues.Unlock() // Unlock the Queues

	if len(models.SyncQueues.Queues[queueName].Messages) > 0 {
		numMsg := 0
		messages = make([]*models.ResultMessage, 0)
		for i := range models.SyncQueues.Queues[queueName].Messages {
			if numMsg >= maxNumberOfMessages {
				break
			}

			if models.SyncQueues.Queues[queueName].Messages[i].ReceiptHandle != "" {
				continue
			}

			randomId := uuid.NewString()

			msg := &models.SyncQueues.Queues[queueName].Messages[i]
			if !msg.IsReadyForReceipt() {
				continue
			}
			msg.ReceiptHandle = msg.Uuid + "#" + randomId
			msg.ReceiptTime = time.Now().UTC()
			msg.VisibilityTimeout = time.Now().Add(time.Duration(models.SyncQueues.Queues[queueName].VisibilityTimeout) * time.Second)

			if models.SyncQueues.Queues[queueName].IsFIFO {
				// If we got messages here it means we have not processed it yet, so get next
				if models.SyncQueues.Queues[queueName].IsLocked(msg.GroupID) {
					continue
				}
				// Otherwise lock messages for group ID
				models.SyncQueues.Queues[queueName].LockGroup(msg.GroupID)
			}

			messages = append(messages, buildResultMessage(msg))

			numMsg++
		}

		respStruct = models.ReceiveMessageResponse{
			"http://queue.amazonaws.com/doc/2012-11-05/",
			models.ReceiveMessageResult{
				Messages: messages,
			},
			models.ResponseMetadata{
				RequestId: "00000000-0000-0000-0000-000000000000",
			},
		}
	} else {
		log.Warning("No messages in Queue:", queueName)
		respStruct = models.ReceiveMessageResponse{Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/", Result: models.ReceiveMessageResult{}, Metadata: models.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	}

	return http.StatusOK, respStruct
}

func buildResultMessage(m *models.SqsMessage) *models.ResultMessage {
	return &models.ResultMessage{
		MessageId:              m.Uuid,
		Body:                   m.MessageBody,
		ReceiptHandle:          m.ReceiptHandle,
		MD5OfBody:              utils.GetMD5Hash(m.MessageBody),
		MD5OfMessageAttributes: m.MD5OfMessageAttributes,
		MessageAttributes:      m.MessageAttributes,
		Attributes: map[string]string{
			"ApproximateFirstReceiveTimestamp": fmt.Sprintf("%d", m.ReceiptTime.UnixNano()/int64(time.Millisecond)),
			"SenderId":                         models.CurrentEnvironment.AccountID,
			"ApproximateReceiveCount":          fmt.Sprintf("%d", m.NumberOfReceives+1),
			"SentTimestamp":                    fmt.Sprintf("%d", time.Now().UTC().UnixNano()/int64(time.Millisecond)),
		},
	}
}
