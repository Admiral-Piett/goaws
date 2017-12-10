package gosqs

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/p4tin/goaws/app"
	"github.com/p4tin/goaws/app/common"
	"github.com/p4tin/goaws/app/models"
)

func periodicTasks() {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			for _, queue := range models.SyncQueues.Queues {
				for _, msg := range queue.Messages {
					if msg.ReceiptHandle != "" {
						if val, ok := models.ReceiptInfos[msg.ReceiptHandle]; ok {
							if val.Timeout.Before(time.Now()) {
								log.Debugf("Making message visible again %s", msg.ReceiptHandle)
								delete(models.ReceiptInfos, msg.ReceiptHandle)
							}
						}
						models.SyncQueues.Lock()
						msg.ReceiptHandle = ""
						models.SyncQueues.Unlock()
					}
				}
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func init() {
	models.ReceiptInfos = make(map[string]*models.ReceiptInfo)
	models.SyncQueues.Queues = make(map[string]*models.Queue)

	app.SqsErrors = make(map[string]app.SqsErrorType)
	app.SqsErrors = make(map[string]app.SqsErrorType)
	err1 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleQueueService.NonExistentQueue", Message: "The specified queue does not exist for this wsdl version."}
	app.SqsErrors["QueueNotFound"] = err1
	err2 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "Duplicate", Code: "AWS.SimpleQueueService.QueueExists", Message: "The specified queue already exists."}
	app.SqsErrors["QueueExists"] = err2
	err3 := app.SqsErrorType{HttpError: http.StatusNotFound, Type: "Not Found", Code: "AWS.SimpleQueueService.QueueExists", Message: "The specified queue does not contain the message specified."}
	app.SqsErrors["MessageDoesNotExist"] = err3
	err4 := app.SqsErrorType{HttpError: http.StatusBadRequest, Type: "GeneralError", Code: "AWS.SimpleQueueService.GeneralError", Message: "General Error."}
	app.SqsErrors["GeneralError"] = err4

	go periodicTasks()
}

func ListQueues(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	respStruct := app.ListQueuesResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	respStruct.Metadata = app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}
	respStruct.Result.QueueUrl = make([]string, 0)

	log.Println("Listing Queues")
	for _, queue := range models.SyncQueues.Queues {
		models.SyncQueues.Lock()
		respStruct.Result.QueueUrl = append(respStruct.Result.QueueUrl, queue.URL)
		models.SyncQueues.Unlock()
	}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func CreateQueue(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	queueName := req.FormValue("QueueName")

	q := models.CreateQueue(req.Host, queueName, 30)

	respStruct := app.CreateQueueResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.CreateQueueResult{QueueUrl: q.URL}, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func SendMessage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	messageBody := req.FormValue("MessageBody")
	messageAttributes := models.ExtractMessageAttributes(req)

	queueUrl := getQueueFromPath(req.FormValue("QueueUrl"), req.URL.String())

	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	msg, err := models.AddMessageToQueue(queueName, messageAttributes, messageBody)
	if err != nil && err.Error() == "QueueNotFound" {
		createErrorResponse(w, req, "QueueNotFound")
		return
	}

	respStruct := app.SendMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.SendMessageResult{MD5OfMessageAttributes: msg.MD5OfMessageAttributes, MD5OfMessageBody: msg.MD5OfMessageBody, MessageId: msg.Uuid}, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func ReceiveMessage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	waitTimeSeconds := 0
	wts := req.FormValue("WaitTimeSeconds")
	if wts != "" {
		waitTimeSeconds, _ = strconv.Atoi(wts)
	}
	maxNumberOfMessages := 1
	mom := req.FormValue("MaxNumberOfMessages")
	if mom != "" {
		maxNumberOfMessages, _ = strconv.Atoi(mom)
	}

	queueUrl := getQueueFromPath(req.FormValue("QueueUrl"), req.URL.String())

	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	if _, ok := models.SyncQueues.Queues[queueName]; !ok {
		createErrorResponse(w, req, "QueueNotFound")
		return
	}

	var message []*app.ResultMessage
	//	respMsg := ResultMessage{}
	respStruct := app.ReceiveMessageResponse{}

	loops := waitTimeSeconds * 10
	for len(models.SyncQueues.Queues[queueName].Messages)-numberOfHiddenMessagesInQueue(*models.SyncQueues.Queues[queueName]) == 0 && loops > 0 {
		time.Sleep(100 * time.Millisecond)
		loops--
	}
	log.Println("Getting Message from Queue:", queueName)

	if len(models.SyncQueues.Queues[queueName].Messages) > 0 {
		numMsg := 0
		message = make([]*app.ResultMessage, 0)
		for i := range models.SyncQueues.Queues[queueName].Messages {
			if numMsg >= maxNumberOfMessages {
				break
			}
			timeout := time.Now().Add(time.Duration(models.SyncQueues.Queues[queueName].TimeoutSecs) * time.Second)
			handle := models.SyncQueues.Queues[queueName].Messages[i].ReceiptHandle
			if (handle != "") && (timeout.Before(models.SyncQueues.Queues[queueName].Messages[i].ReceiptTime)) {
				continue
			} else {
				models.SyncQueues.Lock() // Lock the Queues
				uuid, _ := common.NewUUID()
				handle = models.SyncQueues.Queues[queueName].Messages[i].Uuid + "#" + uuid
				models.SyncQueues.Queues[queueName].Messages[i].ReceiptHandle = handle
				models.SyncQueues.Queues[queueName].Messages[i].ReceiptTime = time.Now().Add(time.Second * time.Duration(models.SyncQueues.Queues[queueName].TimeoutSecs))
				message = append(message, &app.ResultMessage{})
				message[numMsg].MessageId = models.SyncQueues.Queues[queueName].Messages[i].Uuid
				message[numMsg].Body = models.SyncQueues.Queues[queueName].Messages[i].MessageBody
				message[numMsg].ReceiptHandle = handle
				message[numMsg].MD5OfBody = common.GetMD5Hash(string(message[numMsg].Body))
				models.SyncQueues.Unlock() // Unlock the Queues
				ri := &models.ReceiptInfo{
					Timeout:      timeout,
					MessageIndex: i,
					Message:      &models.SyncQueues.Queues[queueName].Messages[i],
				}
				models.ReceiptInfos[handle] = ri
				numMsg++
			}
		}

		//		respMsg = ResultMessage{MessageId: message.Uuid, ReceiptHandle: message.ReceiptHandle, MD5OfBody: message.MD5OfMessageBody, Body: message.MessageBody, MD5OfMessageAttributes: message.MD5OfMessageAttributes}
		respStruct = app.ReceiveMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ReceiveMessageResult{Message: message}, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	} else {
		log.Println("No messages in Queue:", queueName)
		respStruct = app.ReceiveMessageResponse{Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/", Result: app.ReceiveMessageResult{}, Metadata: app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func ChangeMessageVisibility(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	req.ParseForm()
	vars := mux.Vars(req)

	queueUrl := getQueueFromPath(req.FormValue("QueueUrl"), req.URL.String())
	queueName := ""
	if queueUrl == "" {
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}
	receiptHandle := vars["ReceiptHandle"]
	visibilityTimeout, err := strconv.Atoi(vars["VisibilityTimeout"])
	if err != nil {
		createErrorResponse(w, req, "Error")
		return
	}
	if visibilityTimeout/60/60 > 12 {
		createErrorResponse(w, req, "Visibility Timeout too big")
		return
	}

	models.SyncQueues.Lock()
	if _, ok := models.SyncQueues.Queues[queueName]; ok {
		for _, msg := range models.SyncQueues.Queues[queueName].Messages {
			if msg.ReceiptHandle == receiptHandle {
				msg.ChangeMessageVisiblity(visibilityTimeout)
			}
		}
		models.SyncQueues.Unlock()
		createErrorResponse(w, req, "Message not in flight")
		return
	}

	respStruct := app.ChangeMessageVisibilityResult{
		"http://queue.amazonaws.com/doc/2012-11-05/",
		app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000001"}}

	enc := xml.NewEncoder(w)
	enc.Indent(" ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
		createErrorResponse(w, req, "ChangeMessageVisibility - Could not encode response")
		return
	}
}

func numberOfHiddenMessagesInQueue(queue models.Queue) int {
	num := 0
	for i := range queue.Messages {
		timeout := time.Now().Add(time.Duration(-queue.TimeoutSecs) * time.Second)
		if (queue.Messages[i].ReceiptHandle != "") && (timeout.Before(queue.Messages[i].ReceiptTime)) {
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

	models.SyncQueues.Lock()
	if _, ok := models.SyncQueues.Queues[queueName]; ok {
		for i, msg := range models.SyncQueues.Queues[queueName].Messages {
			for _, deleteEntry := range deleteEntries {
				if msg.ReceiptHandle == deleteEntry.ReceiptHandle {
					models.SyncQueues.Queues[queueName].Messages = append(models.SyncQueues.Queues[queueName].Messages[:i], models.SyncQueues.Queues[queueName].Messages[i+1:]...)

					deleteEntry.Deleted = true
					deletedEntry := app.DeleteMessageBatchResultEntry{Id: deleteEntry.Id}
					deletedEntries = append(deletedEntries, deletedEntry)
				}
			}
		}
	}
	models.SyncQueues.Unlock()

	notFoundEntries := make([]app.BatchResultErrorEntry, 0)
	for _, deleteEntry := range deleteEntries {
		if deleteEntry.Deleted == false {
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

func DeleteMessage(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")

	// Retrieve FormValues required
	receiptHandle := req.FormValue("ReceiptHandle")

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

	log.Println("Deleting Message, Queue:", queueName, ", ReceiptHandle:", receiptHandle)

	err := models.RemoveMessageFromQueue(queueName, receiptHandle)
	if err != nil {
		createErrorResponse(w, req, "MessageDoesNotExist")
		return
	}

	//Create, encode/xml and send response
	respStruct := app.DeleteMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000001"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}

	//// Find queue/message with the receipt handle and delete
	//models.SyncQueues.Lock()
	//if _, ok := models.SyncQueues.Queues[queueName]; ok {
	//	for i, msg := range models.SyncQueues.Queues[queueName].Messages {
	//		if msg.ReceiptHandle == receiptHandle {
	//			//Delete message from Q
	//			models.SyncQueues.Queues[queueName].Messages = append(models.SyncQueues.Queues[queueName].Messages[:i], models.SyncQueues.Queues[queueName].Messages[i+1:]...)
	//
	//			models.SyncQueues.Unlock()
	//			// Create, encode/xml and send response
	//			respStruct := app.DeleteMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000001"}}
	//			enc := xml.NewEncoder(w)
	//			enc.Indent("  ", "    ")
	//			if err := enc.Encode(respStruct); err != nil {
	//				log.Printf("error: %v\n", err)
	//			}
	//			models.ReceiptInfos.Lock()
	//			delete(models.ReceiptInfos.Receipts, receiptHandle)
	//			models.ReceiptInfos.Unlock()
	//			return
	//		}
	//	}
	//	log.Println("Receipt Handle not found")
	//} else {
	//	log.Println("Queue not found")
	//}
	//models.SyncQueues.Unlock()

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
	models.SyncQueues.Lock()
	delete(models.SyncQueues.Queues, queueName)
	models.SyncQueues.Unlock()

	// Create, encode/xml and send response
	respStruct := app.DeleteMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func PurgeQueue(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")

	// Retrieve FormValues required
	queueUrl := getQueueFromPath(req.FormValue("QueueUrl"), req.URL.String())

	uriSegments := strings.Split(queueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	log.Println("Purging Queue:", queueName)

	models.SyncQueues.Lock()
	if _, ok := models.SyncQueues.Queues[queueName]; ok {
		models.SyncQueues.Queues[queueName].Messages = nil
		respStruct := app.PurgeQueueResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
		enc := xml.NewEncoder(w)
		enc.Indent("  ", "    ")
		if err := enc.Encode(respStruct); err != nil {
			log.Printf("error: %v\n", err)
			createErrorResponse(w, req, "GeneralError")
		}
	} else {
		log.Println("Purge Queue:", queueName, ", queue does not exist!!!")
		createErrorResponse(w, req, "QueueNotFound")
	}
	models.SyncQueues.Unlock()
}

func GetQueueUrl(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")
	//
	//// Retrieve FormValues required
	queueName := req.FormValue("QueueName")
	models.SyncQueues.RLock()
	if queue, ok := models.SyncQueues.Queues[queueName]; ok {
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
	models.SyncQueues.RUnlock()
}

func GetQueueAttributes(w http.ResponseWriter, req *http.Request) {
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

	log.Println("Get Queue Attributes:", queueName)
	if queue, ok := models.SyncQueues.Queues[queueName]; ok {
		models.SyncQueues.RLock()
		// Create, encode/xml and send response
		attribs := make([]app.Attribute, 0, 0)
		attr := app.Attribute{Name: "VisibilityTimeout", Value: strconv.Itoa(queue.TimeoutSecs)}
		attribs = append(attribs, attr)
		attr = app.Attribute{Name: "DelaySeconds", Value: "0"}
		attribs = append(attribs, attr)
		attr = app.Attribute{Name: "ReceiveMessageWaitTimeSeconds", Value: "0"}
		attribs = append(attribs, attr)
		attr = app.Attribute{Name: "ApproximateNumberOfMessages", Value: strconv.Itoa(len(queue.Messages))}
		attribs = append(attribs, attr)
		attr = app.Attribute{Name: "ApproximateNumberOfMessagesNotVisible", Value: strconv.Itoa(numberOfHiddenMessagesInQueue(*queue))}
		attribs = append(attribs, attr)
		attr = app.Attribute{Name: "CreatedTimestamp", Value: "0000000000"}
		attribs = append(attribs, attr)
		attr = app.Attribute{Name: "LastModifiedTimestamp", Value: "0000000000"}
		attribs = append(attribs, attr)
		attr = app.Attribute{Name: "QueueArn", Value: queue.Arn}
		attribs = append(attribs, attr)
		models.SyncQueues.RUnlock()

		result := app.GetQueueAttributesResult{Attrs: attribs}
		respStruct := app.GetQueueAttributesResponse{"http://queue.amazonaws.com/doc/2012-11-05/", result, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
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

func SetQueueAttributes(w http.ResponseWriter, req *http.Request) {
	log.Println("setQueueAttributes was called but it's not implemented")
	respStruct := app.SetQueueAttributesResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
		createErrorResponse(w, req, "GeneralError")
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
	respStruct := app.ErrorResponse{app.ErrorResult{Type: er.Type, Code: er.Code, Message: er.Message, RequestId: "00000000-0000-0000-0000-000000000000"}}

	w.WriteHeader(er.HttpError)
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}
