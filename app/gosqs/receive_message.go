package gosqs

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/common"
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

	if _, ok := app.SyncQueues.Queues[queueName]; !ok {
		return utils.CreateErrorResponseV1("QueueNotFound", true)
	}

	var messages []*models.ResultMessage
	respStruct := models.ReceiveMessageResponse{}

	waitTimeSeconds := requestBody.WaitTimeSeconds
	if waitTimeSeconds == 0 {
		app.SyncQueues.RLock()
		waitTimeSeconds = app.SyncQueues.Queues[queueName].ReceiveMessageWaitTimeSeconds
		app.SyncQueues.RUnlock()
	}

	loops := waitTimeSeconds * 10
	for loops > 0 {
		app.SyncQueues.RLock()
		_, queueFound := app.SyncQueues.Queues[queueName]
		if !queueFound {
			app.SyncQueues.RUnlock()
			return utils.CreateErrorResponseV1("QueueNotFound", true)
		}
		messageFound := len(app.SyncQueues.Queues[queueName].Messages)-numberOfHiddenMessagesInQueue(*app.SyncQueues.Queues[queueName]) != 0
		app.SyncQueues.RUnlock()
		if !messageFound {
			continueTimer := time.NewTimer(100 * time.Millisecond)
			select {
			case <-req.Context().Done():
				continueTimer.Stop()
				return http.StatusOK, models.ReceiveMessageResponse{
					Xmlns:    models.BASE_XMLNS,
					Result:   models.ReceiveMessageResult{},
					Metadata: models.BASE_RESPONSE_METADATA,
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

	app.SyncQueues.Lock()         // Lock the Queues
	defer app.SyncQueues.Unlock() // Unlock the Queues

	if len(app.SyncQueues.Queues[queueName].Messages) > 0 {
		numMsg := 0
		messages = make([]*models.ResultMessage, 0)
		for i := range app.SyncQueues.Queues[queueName].Messages {
			if numMsg >= maxNumberOfMessages {
				break
			}

			if app.SyncQueues.Queues[queueName].Messages[i].ReceiptHandle != "" {
				continue
			}

			uuid, _ := common.NewUUID()

			msg := &app.SyncQueues.Queues[queueName].Messages[i]
			if !msg.IsReadyForReceipt() {
				continue
			}
			msg.ReceiptHandle = msg.Uuid + "#" + uuid
			msg.ReceiptTime = time.Now().UTC()
			msg.VisibilityTimeout = time.Now().Add(time.Duration(app.SyncQueues.Queues[queueName].VisibilityTimeout) * time.Second)

			if app.SyncQueues.Queues[queueName].IsFIFO {
				// If we got messages here it means we have not processed it yet, so get next
				if app.SyncQueues.Queues[queueName].IsLocked(msg.GroupID) {
					continue
				}
				// Otherwise lock messages for group ID
				app.SyncQueues.Queues[queueName].LockGroup(msg.GroupID)
			}

			messages = append(messages, getMessageResult(msg))

			numMsg++
		}

		respStruct = models.ReceiveMessageResponse{
			"http://queue.amazonaws.com/doc/2012-11-05/",
			models.ReceiveMessageResult{
				Messages: messages,
			},
			app.ResponseMetadata{
				RequestId: "00000000-0000-0000-0000-000000000000",
			},
		}
	} else {
		log.Warning("No messages in Queue:", queueName)
		respStruct = models.ReceiveMessageResponse{Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/", Result: models.ReceiveMessageResult{}, Metadata: app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	}

	return http.StatusOK, respStruct
}

func getMessageResult(m *app.Message) *models.ResultMessage {
	msgMttrs := []*models.ResultMessageAttribute{}
	for _, attr := range m.MessageAttributes {
		msgMttrs = append(msgMttrs, getMessageAttributeResult(&attr))
	}

	attrsMap := map[string]string{
		"ApproximateFirstReceiveTimestamp": fmt.Sprintf("%d", m.ReceiptTime.UnixNano()/int64(time.Millisecond)),
		"SenderId":                         app.CurrentEnvironment.AccountID,
		"ApproximateReceiveCount":          fmt.Sprintf("%d", m.NumberOfReceives+1),
		"SentTimestamp":                    fmt.Sprintf("%d", time.Now().UTC().UnixNano()/int64(time.Millisecond)),
	}

	var attrs []*models.ResultAttribute
	for k, v := range attrsMap {
		attrs = append(attrs, &models.ResultAttribute{
			Name:  k,
			Value: v,
		})
	}

	return &models.ResultMessage{
		MessageId:              m.Uuid,
		Body:                   m.MessageBody,
		ReceiptHandle:          m.ReceiptHandle,
		MD5OfBody:              common.GetMD5Hash(string(m.MessageBody)),
		MD5OfMessageAttributes: m.MD5OfMessageAttributes,
		MessageAttributes:      msgMttrs,
		Attributes:             attrs,
	}
}
