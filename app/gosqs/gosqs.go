package gosqs

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	"github.com/Admiral-Piett/goaws/app/models"

	log "github.com/sirupsen/logrus"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/common"
	"github.com/gorilla/mux"
)

func init() {
	app.SyncQueues.Queues = make(map[string]*app.Queue)

	app.SqsErrors = make(map[string]app.SqsErrorType)
	err1 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleQueueService.NonExistentQueue", Message: "The specified queue does not exist for this wsdl version."}
	app.SqsErrors["QueueNotFound"] = err1
	err2 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "Duplicate", Code: "AWS.SimpleQueueService.QueueExists", Message: "The specified queue already exists."}
	app.SqsErrors["QueueExists"] = err2
	err3 := app.SqsErrorType{HttpError: http.StatusNotFound, Type: "Not Found", Code: "AWS.SimpleQueueService.QueueExists", Message: "The specified queue does not contain the message specified."}
	app.SqsErrors["MessageDoesNotExist"] = err3
	err4 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "GeneralError", Code: "AWS.SimpleQueueService.GeneralError", Message: "General Error."}
	app.SqsErrors["GeneralError"] = err4
	err5 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "TooManyEntriesInBatchRequest", Code: "AWS.SimpleQueueService.TooManyEntriesInBatchRequest", Message: "Maximum number of entries per request are 10."}
	app.SqsErrors["TooManyEntriesInBatchRequest"] = err5
	err6 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "BatchEntryIdsNotDistinct", Code: "AWS.SimpleQueueService.BatchEntryIdsNotDistinct", Message: "Two or more batch entries in the request have the same Id."}
	app.SqsErrors["BatchEntryIdsNotDistinct"] = err6
	err7 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "EmptyBatchRequest", Code: "AWS.SimpleQueueService.EmptyBatchRequest", Message: "The batch request doesn't contain any entries."}
	app.SqsErrors["EmptyBatchRequest"] = err7
	err8 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "ValidationError", Code: "AWS.SimpleQueueService.ValidationError", Message: "The visibility timeout is incorrect"}
	app.SqsErrors["InvalidVisibilityTimeout"] = err8
	err9 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "MessageNotInFlight", Code: "AWS.SimpleQueueService.MessageNotInFlight", Message: "The message referred to isn't in flight."}
	app.SqsErrors["MessageNotInFlight"] = err9
	err10 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "MessageTooBig", Code: "InvalidParameterValue", Message: "The message size exceeds the limit."}
	app.SqsErrors["MessageTooBig"] = err10
	app.SqsErrors[ErrInvalidParameterValue.Type] = *ErrInvalidParameterValue
	app.SqsErrors[ErrInvalidAttributeValue.Type] = *ErrInvalidAttributeValue
}

func PeriodicTasks(d time.Duration, quit <-chan struct{}) {
	ticker := time.NewTicker(d)
	for {
		select {
		case <-ticker.C:
			app.SyncQueues.Lock()
			for j := range app.SyncQueues.Queues {
				queue := app.SyncQueues.Queues[j]

				log.Debugf("Queue [%s] length [%d]", queue.Name, len(queue.Messages))
				for i := 0; i < len(queue.Messages); i++ {
					msg := &queue.Messages[i]

					// Reset deduplication period
					for dedupId, startTime := range queue.Duplicates {
						if time.Now().After(startTime.Add(app.DeduplicationPeriod)) {
							log.Debugf("deduplication period for message with deduplicationId [%s] expired", dedupId)
							delete(queue.Duplicates, dedupId)
						}
					}

					if msg.ReceiptHandle != "" {
						if msg.VisibilityTimeout.Before(time.Now()) {
							log.Debugf("Making message visible again %s", msg.ReceiptHandle)
							queue.UnlockGroup(msg.GroupID)
							msg.ReceiptHandle = ""
							msg.ReceiptTime = time.Now().UTC()
							msg.Retry++
							if queue.MaxReceiveCount > 0 &&
								queue.DeadLetterQueue != nil &&
								msg.Retry > queue.MaxReceiveCount {
								queue.DeadLetterQueue.Messages = append(queue.DeadLetterQueue.Messages, *msg)
								queue.Messages = append(queue.Messages[:i], queue.Messages[i+1:]...)
								i++
							}
						}
					}
				}
			}
			app.SyncQueues.Unlock()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

type SendEntry struct {
	Id                     string
	MessageBody            string
	MessageAttributes      map[string]app.MessageAttributeValue
	MessageGroupId         string
	MessageDeduplicationId string
}

func SendMessageBatch(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	req.ParseForm()

	queueUrl := getQueueFromPath(req.FormValue("QueueUrl"), req.URL.String())
	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	if _, ok := app.SyncQueues.Queues[queueName]; !ok {
		createErrorResponse(w, req, "QueueNotFound")
		return
	}

	sendEntries := []SendEntry{}

	for k, v := range req.Form {
		keySegments := strings.Split(k, ".")
		if keySegments[0] == "SendMessageBatchRequestEntry" {
			if len(keySegments) < 3 {
				createErrorResponse(w, req, "EmptyBatchRequest")
				return
			}
			keyIndex, err := strconv.Atoi(keySegments[1])
			if err != nil {
				createErrorResponse(w, req, "Error")
				return
			}

			if len(sendEntries) < keyIndex {
				newSendEntries := make([]SendEntry, keyIndex)
				copy(newSendEntries, sendEntries)
				sendEntries = newSendEntries
			}

			if keySegments[2] == "Id" {
				sendEntries[keyIndex-1].Id = v[0]
			}

			if keySegments[2] == "MessageBody" {
				sendEntries[keyIndex-1].MessageBody = v[0]
			}

			if keySegments[2] == "MessageGroupId" {
				sendEntries[keyIndex-1].MessageGroupId = v[0]
			}

			if keySegments[2] == "MessageDeduplicationId" {
				sendEntries[keyIndex-1].MessageDeduplicationId = v[0]
			}
		}
	}

	if len(sendEntries) == 0 {
		createErrorResponse(w, req, "EmptyBatchRequest")
		return
	}

	if len(sendEntries) > 10 {
		createErrorResponse(w, req, "TooManyEntriesInBatchRequest")
		return
	}
	ids := map[string]struct{}{}
	for _, v := range sendEntries {
		if _, ok := ids[v.Id]; ok {
			createErrorResponse(w, req, "BatchEntryIdsNotDistinct")
			return
		}
		ids[v.Id] = struct{}{}
	}

	sentEntries := make([]app.SendMessageBatchResultEntry, 0)
	log.Println("Putting Message in Queue:", queueName)
	for _, sendEntry := range sendEntries {
		msg := app.Message{MessageBody: []byte(sendEntry.MessageBody)}
		if len(sendEntry.MessageAttributes) > 0 {
			msg.MessageAttributes = sendEntry.MessageAttributes
			msg.MD5OfMessageAttributes = common.HashAttributes(sendEntry.MessageAttributes)
		}
		msg.MD5OfMessageBody = common.GetMD5Hash(sendEntry.MessageBody)
		msg.GroupID = sendEntry.MessageGroupId
		msg.DeduplicationID = sendEntry.MessageDeduplicationId
		msg.Uuid, _ = common.NewUUID()
		msg.SentTime = time.Now()
		app.SyncQueues.Lock()
		fifoSeqNumber := ""
		if app.SyncQueues.Queues[queueName].IsFIFO {
			fifoSeqNumber = app.SyncQueues.Queues[queueName].NextSequenceNumber(sendEntry.MessageGroupId)
		}

		if !app.SyncQueues.Queues[queueName].IsDuplicate(sendEntry.MessageDeduplicationId) {
			app.SyncQueues.Queues[queueName].Messages = append(app.SyncQueues.Queues[queueName].Messages, msg)
		} else {
			log.Debugf("Message with deduplicationId [%s] in queue [%s] is duplicate ", sendEntry.MessageDeduplicationId, queueName)
		}

		app.SyncQueues.Queues[queueName].InitDuplicatation(sendEntry.MessageDeduplicationId)

		app.SyncQueues.Unlock()
		se := app.SendMessageBatchResultEntry{
			Id:                     sendEntry.Id,
			MessageId:              msg.Uuid,
			MD5OfMessageBody:       msg.MD5OfMessageBody,
			MD5OfMessageAttributes: msg.MD5OfMessageAttributes,
			SequenceNumber:         fifoSeqNumber,
		}
		sentEntries = append(sentEntries, se)
		log.Infof("%s: Queue: %s, Message: %s\n", time.Now().Format("2006-01-02 15:04:05"), queueName, msg.MessageBody)
	}

	respStruct := app.SendMessageBatchResponse{
		"http://queue.amazonaws.com/doc/2012-11-05/",
		app.SendMessageBatchResult{Entry: sentEntries},
		app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000001"}}

	enc := xml.NewEncoder(w)
	enc.Indent(" ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func numberOfHiddenMessagesInQueue(queue app.Queue) int {
	num := 0
	for _, m := range queue.Messages {
		if m.ReceiptHandle != "" || m.DelaySecs > 0 && time.Now().Before(m.SentTime.Add(time.Duration(m.DelaySecs)*time.Second)) {
			num++
		}
	}
	return num
}

type DeleteEntry struct {
	Id            string
	ReceiptHandle string
	Error         string
	Deleted       bool
}

func DeleteMessageBatch(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	req.ParseForm()

	queueUrl := getQueueFromPath(req.FormValue("QueueUrl"), req.URL.String())
	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	deleteEntries := []DeleteEntry{}

	for k, v := range req.Form {
		keySegments := strings.Split(k, ".")
		if keySegments[0] == "DeleteMessageBatchRequestEntry" {
			keyIndex, err := strconv.Atoi(keySegments[1])
			if err != nil {
				createErrorResponse(w, req, "Error")
				return
			}

			if len(deleteEntries) < keyIndex {
				newDeleteEntries := make([]DeleteEntry, keyIndex)
				copy(newDeleteEntries, deleteEntries)
				deleteEntries = newDeleteEntries
			}

			if keySegments[2] == "Id" {
				deleteEntries[keyIndex-1].Id = v[0]
			}

			if keySegments[2] == "ReceiptHandle" {
				deleteEntries[keyIndex-1].ReceiptHandle = v[0]
			}
		}
	}

	deletedEntries := make([]app.DeleteMessageBatchResultEntry, 0)

	app.SyncQueues.Lock()
	if _, ok := app.SyncQueues.Queues[queueName]; ok {
		for _, deleteEntry := range deleteEntries {
			for i, msg := range app.SyncQueues.Queues[queueName].Messages {
				if msg.ReceiptHandle == deleteEntry.ReceiptHandle {
					// Unlock messages for the group
					log.Printf("FIFO Queue %s unlocking group %s:", queueName, msg.GroupID)
					app.SyncQueues.Queues[queueName].UnlockGroup(msg.GroupID)
					app.SyncQueues.Queues[queueName].Messages = append(app.SyncQueues.Queues[queueName].Messages[:i], app.SyncQueues.Queues[queueName].Messages[i+1:]...)
					delete(app.SyncQueues.Queues[queueName].Duplicates, msg.DeduplicationID)

					deleteEntry.Deleted = true
					deletedEntry := app.DeleteMessageBatchResultEntry{Id: deleteEntry.Id}
					deletedEntries = append(deletedEntries, deletedEntry)
					break
				}
			}
		}
	}
	app.SyncQueues.Unlock()

	notFoundEntries := make([]app.BatchResultErrorEntry, 0)
	for _, deleteEntry := range deleteEntries {
		if deleteEntry.Deleted {
			notFoundEntries = append(notFoundEntries, app.BatchResultErrorEntry{
				Code:        "1",
				Id:          deleteEntry.Id,
				Message:     "Message not found",
				SenderFault: true})
		}
	}

	respStruct := app.DeleteMessageBatchResponse{
		"http://queue.amazonaws.com/doc/2012-11-05/",
		app.DeleteMessageBatchResult{Entry: deletedEntries, Error: notFoundEntries},
		app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000001"}}

	enc := xml.NewEncoder(w)
	enc.Indent(" ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func DeleteQueue(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")

	// Retrieve FormValues required
	queueUrl := getQueueFromPath(req.FormValue("QueueUrl"), req.URL.String())
	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	log.Println("Deleting Queue:", queueName)
	app.SyncQueues.Lock()
	delete(app.SyncQueues.Queues, queueName)
	app.SyncQueues.Unlock()

	// Create, encode/xml and send response
	respStruct := app.DeleteQueueResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func GetQueueUrl(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")
	//
	//// Retrieve FormValues required
	queueName := req.FormValue("QueueName")
	if queue, ok := app.SyncQueues.Queues[queueName]; ok {
		url := queue.URL
		log.Println("Get Queue URL:", queueName)
		// Create, encode/xml and send response
		result := app.GetQueueUrlResult{QueueUrl: url}
		respStruct := app.GetQueueUrlResponse{"http://queue.amazonaws.com/doc/2012-11-05/", result, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
		enc := xml.NewEncoder(w)
		enc.Indent("  ", "    ")
		if err := enc.Encode(respStruct); err != nil {
			log.Printf("error: %v\n", err)
		}
	} else {
		log.Println("Get Queue URL:", queueName, ", queue does not exist!!!")
		createErrorResponse(w, req, "QueueNotFound")
	}
}

func getQueueFromPath(formVal string, theUrl string) string {
	if formVal != "" {
		return formVal
	}
	u, err := url.Parse(theUrl)
	if err != nil {
		return ""
	}
	return u.Path
}

func createErrorResponse(w http.ResponseWriter, req *http.Request, err string) {
	er := app.SqsErrors[err]
	respStruct := models.ErrorResponse{
		Result:    models.ErrorResult{Type: er.Type, Code: er.Code, Message: er.Message},
		RequestId: "00000000-0000-0000-0000-000000000000",
	}

	w.WriteHeader(er.HttpError)
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func createErrorResponseV1(err string) (int, interfaces.AbstractResponseBody) {
	er := app.SqsErrors[err]
	respStruct := models.ErrorResponse{
		Result:    models.ErrorResult{Type: er.Type, Code: er.Code, Message: er.Message},
		RequestId: "00000000-0000-0000-0000-000000000000", // TODO - fix
	}
	return er.HttpError, respStruct
}
