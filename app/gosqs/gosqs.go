package gosqs

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/p4tin/goaws/app"
	"github.com/p4tin/goaws/app/common"
	"log"
)

type SqsErrorType struct {
	HttpError int
	Type      string
	Code      string
	Message   string
}

var SqsErrors map[string]SqsErrorType

type Message struct {
	MessageBody            []byte
	Uuid                   string
	MD5OfMessageAttributes string
	MD5OfMessageBody       string
	ReceiptHandle          string
	ReceiptTime            time.Time
}

type Queue struct {
	Name        string
	URL         string
	Arn         string
	TimeoutSecs int
	Messages    []Message
}

var SyncQueues = struct {
	sync.RWMutex
	Queues map[string]*Queue
}{Queues: make(map[string]*Queue)}

//var Queues map[string]*Queue

func init() {
	SyncQueues.Queues = make(map[string]*Queue)

	SqsErrors = make(map[string]SqsErrorType)
	err1 := SqsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleQueueService.NonExistentQueue", Message: "The specified ueue does not exist for this wsdl version."}
	SqsErrors["QueueNotFound"] = err1
	err2 := SqsErrorType{HttpError: http.StatusBadRequest, Type: "Duplicate", Code: "AWS.SimpleQueueService.QueueExists", Message: "The specified queue already exists."}
	SqsErrors["QueueExists"] = err2
	err3 := SqsErrorType{HttpError: http.StatusNotFound, Type: "Not Found", Code: "AWS.SimpleQueueService.QueueExists", Message: "The specified queue does not contain the message specified."}
	SqsErrors["MessageDoesNotExist"] = err3
	err4 := SqsErrorType{HttpError: http.StatusBadRequest, Type: "GeneralError", Code: "AWS.SimpleQueueService.GeneralError", Message: "General Error."}
	SqsErrors["GeneralError"] = err4
}

func ListQueues(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	respStruct := app.ListQueuesResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	respStruct.Metadata = app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}
	respStruct.Result.QueueUrl = make([]string, 0)

	log.Println("Listing Queues")
	for _, queue := range SyncQueues.Queues {
		SyncQueues.Lock()
		respStruct.Result.QueueUrl = append(respStruct.Result.QueueUrl, queue.URL)
		SyncQueues.Unlock()
	}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func CreateQueue(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	queueName := req.FormValue("QueueName")
	queueUrl := "http://" + req.Host + "/queue/" + queueName

	if _, ok := SyncQueues.Queues[queueName]; !ok {
		log.Println("Creating Queue:", queueName)
		queue := &Queue{Name: queueName, URL: queueUrl, Arn: queueUrl, TimeoutSecs: 30}
		SyncQueues.RLock()
		SyncQueues.Queues[queueName] = queue
		SyncQueues.RUnlock()
	}

	respStruct := app.CreateQueueResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.CreateQueueResult{QueueUrl: queueUrl}, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func SendMessage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	messageBody := req.FormValue("MessageBody")

	queueUrl := getQueueUrl(req.FormValue("QueueUrl"), req.URL.String())

	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	if _, ok := SyncQueues.Queues[queueName]; !ok {
		// Queue does not exists
		createErrorResponse(w, req, "QueueNotFound")
		return
	}

	log.Println("Putting Message in Queue:", queueName)
	msg := Message{MessageBody: []byte(messageBody)}
	msg.MD5OfMessageAttributes = common.GetMD5Hash("GoAws")
	msg.MD5OfMessageBody = common.GetMD5Hash(messageBody)
	msg.Uuid, _ = common.NewUUID()
	SyncQueues.Lock()
	SyncQueues.Queues[queueName].Messages = append(SyncQueues.Queues[queueName].Messages, msg)
	SyncQueues.Unlock()
	common.LogMessage(fmt.Sprintf("%s: Queue: %s, Message: %s\n", time.Now().Format("2006-01-02 15:04:05"), queueName, msg.MessageBody))

	respStruct := app.SendMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.SendMessageResult{MD5OfMessageAttributes: msg.MD5OfMessageAttributes, MD5OfMessageBody: msg.MD5OfMessageBody, MessageId: msg.Uuid}, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
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

	queueUrl := getQueueUrl(req.FormValue("QueueUrl"), req.URL.String())
	log.Println("Queue for receive:", queueUrl)

	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	if _, ok := SyncQueues.Queues[queueName]; !ok {
		createErrorResponse(w, req, "QueueNotFound")
		return
	}

	var message []*app.ResultMessage
	//	respMsg := ResultMessage{}
	respStruct := app.ReceiveMessageResponse{}

	loops := waitTimeSeconds * 10
	for len(SyncQueues.Queues[queueName].Messages)-numberOfHiddenMessagesInQueue(*SyncQueues.Queues[queueName]) == 0 && loops > 0 {
		time.Sleep(100 * time.Millisecond)
		loops--
	}
	log.Println("Getting Message from Queue:", queueName)

	if len(SyncQueues.Queues[queueName].Messages) > 0 {
		numMsg := 0
		message = make([]*app.ResultMessage, 0)
		for i, _ := range SyncQueues.Queues[queueName].Messages {
			if numMsg >= maxNumberOfMessages {
				break
			}
			timeout := time.Now().Add(time.Duration(-SyncQueues.Queues[queueName].TimeoutSecs) * time.Second)
			if (SyncQueues.Queues[queueName].Messages[i].ReceiptHandle != "") && (timeout.Before(SyncQueues.Queues[queueName].Messages[i].ReceiptTime)) {
				continue
			} else {
				SyncQueues.Lock() // Lock the Queues
				uuid, _ := common.NewUUID()
				SyncQueues.Queues[queueName].Messages[i].ReceiptHandle = SyncQueues.Queues[queueName].Messages[i].Uuid + "#" + uuid
				SyncQueues.Queues[queueName].Messages[i].ReceiptTime = time.Now()
				message = append(message, &app.ResultMessage{})
				message[numMsg].MessageId = SyncQueues.Queues[queueName].Messages[i].Uuid
				message[numMsg].Body = SyncQueues.Queues[queueName].Messages[i].MessageBody
				message[numMsg].ReceiptHandle = SyncQueues.Queues[queueName].Messages[i].ReceiptHandle
				message[numMsg].MD5OfBody = common.GetMD5Hash(string(message[numMsg].Body))
				SyncQueues.Unlock() // Unlock the Queues
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
		fmt.Printf("error: %v\n", err)
	}
}

func numberOfHiddenMessagesInQueue(queue Queue) int {
	num := 0
	for i, _ := range queue.Messages {
		timeout := time.Now().Add(time.Duration(-queue.TimeoutSecs) * time.Second)
		if (queue.Messages[i].ReceiptHandle != "") && (timeout.Before(queue.Messages[i].ReceiptTime)) {
			num++
		}
	}
	return num
}

func DeleteMessage(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")

	// Retrieve FormValues required
	receiptHandle := req.FormValue("ReceiptHandle")

	// Retrieve FormValues required
	queueUrl := getQueueUrl(req.FormValue("QueueUrl"), req.URL.String())
	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	log.Println("Deleting Message, Queue:", queueName, ", ReceiptHandle:", receiptHandle)

	// Find queue/message with the receipt handle and delete
	SyncQueues.Lock()
	if _, ok := SyncQueues.Queues[queueName]; ok {
		for i, msg := range SyncQueues.Queues[queueName].Messages {
			if msg.ReceiptHandle == receiptHandle {
				//Delete message from Q
				log.Println("Found receipt")
				SyncQueues.Queues[queueName].Messages = append(SyncQueues.Queues[queueName].Messages[:i], SyncQueues.Queues[queueName].Messages[i+1:]...)

				SyncQueues.Unlock()
				// Create, encode/xml and send response
				respStruct := app.DeleteMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000001"}}
				enc := xml.NewEncoder(w)
				enc.Indent("  ", "    ")
				if err := enc.Encode(respStruct); err != nil {
					fmt.Printf("error: %v\n", err)
				}
				return
			}
		}
		log.Println("Receipt Handle not found")
	} else {
		log.Println("Queue not found")
	}
	SyncQueues.Unlock()

	createErrorResponse(w, req, "MessageDoesNotExist")
}

func DeleteQueue(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")

	// Retrieve FormValues required
	queueUrl := getQueueUrl(req.FormValue("QueueUrl"), req.URL.String())
	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	log.Println("Deleting Queue:", queueName)
	SyncQueues.Lock()
	delete(SyncQueues.Queues, queueName)
	SyncQueues.Unlock()

	// Create, encode/xml and send response
	respStruct := app.DeleteMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func PurgeQueue(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")

	// Retrieve FormValues required
	queueUrl := getQueueUrl(req.FormValue("QueueUrl"), req.URL.String())

	uriSegments := strings.Split(queueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	log.Println("Purging Queue:", queueName)

	SyncQueues.Lock()
	if _, ok := SyncQueues.Queues[queueName]; ok {
		SyncQueues.Queues[queueName].Messages = nil
		respStruct := app.PurgeQueueResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
		enc := xml.NewEncoder(w)
		enc.Indent("  ", "    ")
		if err := enc.Encode(respStruct); err != nil {
			fmt.Printf("error: %v\n", err)
			createErrorResponse(w, req, "GeneralError")
		}
	} else {
		log.Println("Purge Queue:", queueName, ", queue does not exist!!!")
		createErrorResponse(w, req, "QueueNotFound")
	}
	SyncQueues.Unlock()
}

func GetQueueUrl(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")
	//
	//// Retrieve FormValues required
	queueName := req.FormValue("QueueName")
	SyncQueues.RLock()
	if queue, ok := SyncQueues.Queues[queueName]; ok {
		url := queue.URL
		log.Println("Get Queue URL:", queueName)
		// Create, encode/xml and send response
		result := app.GetQueueUrlResult{QueueUrl: url}
		respStruct := app.GetQueueUrlResponse{"http://queue.amazonaws.com/doc/2012-11-05/", result, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
		enc := xml.NewEncoder(w)
		enc.Indent("  ", "    ")
		if err := enc.Encode(respStruct); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	} else {
		log.Println("Get Queue URL:", queueName, ", queue does not exist!!!")
		createErrorResponse(w, req, "QueueNotFound")
	}
	SyncQueues.RUnlock()
}

func GetQueueAttributes(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")
	// Retrieve FormValues required
	queueUrl := getQueueUrl(req.FormValue("QueueUrl"), req.URL.String())
	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	log.Println("Get Queue Attributes:", queueName)
	if queue, ok := SyncQueues.Queues[queueName]; ok {
		SyncQueues.RLock()
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
		SyncQueues.RUnlock()

		result := app.GetQueueAttributesResult{Attrs: attribs}
		respStruct := app.GetQueueAttributesResponse{"http://queue.amazonaws.com/doc/2012-11-05/", result, app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
		enc := xml.NewEncoder(w)
		enc.Indent("  ", "    ")
		if err := enc.Encode(respStruct); err != nil {
			fmt.Printf("error: %v\n", err)
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
		fmt.Printf("error: %v\n", err)
		createErrorResponse(w, req, "GeneralError")
	}
}

func getQueueUrl(formVal string, url string) string {
	if formVal != "" {
		return formVal
	}
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func createErrorResponse(w http.ResponseWriter, req *http.Request, err string) {
	er := SqsErrors[err]
	respStruct := app.ErrorResponse{app.ErrorResult{Type: er.Type, Code: er.Code, Message: er.Message, RequestId: "00000000-0000-0000-0000-000000000000"}}

	w.WriteHeader(er.HttpError)
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
