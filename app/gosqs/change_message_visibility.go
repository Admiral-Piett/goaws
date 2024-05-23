package gosqs

import (
	"net/http"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func ChangeMessageVisibilityV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewChangeMessageVisibilityRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req)
	if !ok {
		log.Error("Invalid Request - ChangeMessageVisibilityV1")
		return createErrorResponseV1(ErrInvalidParameterValue.Type)
	}

	vars := mux.Vars(req)

	queueUrl := requestBody.QueueUrl
	queueName := ""
	if queueUrl == "" {
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	receiptHandle := requestBody.ReceiptHandle

	visibilityTimeout := requestBody.VisibilityTimeout
	if visibilityTimeout > 43200 {
		return createErrorResponseV1("ValidationError")
	}

	if _, ok := app.SyncQueues.Queues[queueName]; !ok {
		return createErrorResponseV1("QueueNotFound")
	}

	app.SyncQueues.Lock()
	messageFound := false
	for i := 0; i < len(app.SyncQueues.Queues[queueName].Messages); i++ {
		queue := app.SyncQueues.Queues[queueName]
		msgs := queue.Messages
		if msgs[i].ReceiptHandle == receiptHandle {
			timeout := app.SyncQueues.Queues[queueName].VisibilityTimeout
			if visibilityTimeout == 0 {
				msgs[i].ReceiptTime = time.Now().UTC()
				msgs[i].ReceiptHandle = ""
				msgs[i].VisibilityTimeout = time.Now().Add(time.Duration(timeout) * time.Second)
				msgs[i].Retry++
				if queue.MaxReceiveCount > 0 &&
					queue.DeadLetterQueue != nil &&
					msgs[i].Retry > queue.MaxReceiveCount {
					queue.DeadLetterQueue.Messages = append(queue.DeadLetterQueue.Messages, msgs[i])
					queue.Messages = append(queue.Messages[:i], queue.Messages[i+1:]...)
					i++
				}
			} else {
				msgs[i].VisibilityTimeout = time.Now().Add(time.Duration(visibilityTimeout) * time.Second)
			}
			messageFound = true
			break
		}
	}
	app.SyncQueues.Unlock()
	if !messageFound {
		return createErrorResponseV1("MessageNotInFlight")
	}

	respStruct := models.ChangeMessageVisibilityResult{
		"http://queue.amazonaws.com/doc/2012-11-05/",
		app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000001"}}

	return http.StatusOK, &respStruct
}
