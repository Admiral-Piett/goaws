package gosqs

import (
	"net/http"
	"encoding/xml"
	"strings"
	"fmt"
	"time"
	"log"
)

type Message struct {
	messageBody 		[]byte
	Uuid 			string
	MD5OfMessageAttributes 	string
	MD5OfMessageBody 	string
	ReceiptHandle		string
	ReceiptTime 		time.Time
}

type Queue struct {
	Name 		string
	URL 		string
	Arn 		string
	Messages 	[]Message
}

var Queues map[string]*Queue

func init() {
	Queues = make(map[string]*Queue)
}

func ListQueues(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	respStruct := ListQueuesResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	respStruct.Metadata = ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}
	respStruct.Result.QueueUrl = make([]string, 0)

	log.Println("Listing Queues")
	for _, queue := range Queues {
		respStruct.Result.QueueUrl = append(respStruct.Result.QueueUrl, queue.URL)
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
	queueUrl := "http://" + req.Host + req.URL.RequestURI() + queueName

	log.Println("Creating Queue:", queueName)
	queue := &Queue{Name: queueName, URL: queueUrl, Arn: queueUrl}
	Queues[queueName] = queue

	respStruct := CreateQueueResponse{"http://queue.amazonaws.com/doc/2012-11-05/", CreateQueueResult{QueueUrl: queueUrl}, ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func SendMessage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	messageBody := req.FormValue("MessageBody")
	queueUrl := req.FormValue("QueueUrl")

	uriSegments := strings.Split(queueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	log.Println("Putting Message in Queue:", queueName)
	msg := Message{messageBody: []byte(messageBody)}
	msg.MD5OfMessageAttributes = GetMD5Hash("GoAws")
	msg.MD5OfMessageBody = GetMD5Hash(messageBody)
	msg.Uuid, _ = NewUUID()
	Queues[queueName].Messages = append(Queues[queueName].Messages, msg)

	respStruct := SendMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", SendMessageResult{MD5OfMessageAttributes: msg.MD5OfMessageAttributes, MD5OfMessageBody: msg.MD5OfMessageBody, MessageId: msg.Uuid}, ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}


func ReceiveMessage(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	queueUrl := req.FormValue("QueueUrl")

	uriSegments := strings.Split(queueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	var message Message
	respMsg := ResultMessage{}
	respStruct := ReceiveMessageResponse{}

	log.Println("Getting Message from Queue:", queueName)
	if len(Queues[queueName].Messages) > 0 {
		for i, _ := range Queues[queueName].Messages {
			if Queues[queueName].Messages[i].ReceiptHandle != "" {
				continue
			}
			uuid, _ := NewUUID()
			Queues[queueName].Messages[i].ReceiptHandle = Queues[queueName].Messages[i].Uuid + "#" + uuid
			Queues[queueName].Messages[i].ReceiptTime = time.Now()
			message = Queues[queueName].Messages[i]
		}

		respMsg = ResultMessage{MessageId: message.Uuid, ReceiptHandle: message.ReceiptHandle, MD5OfBody: message.MD5OfMessageBody, Body: message.messageBody, MD5OfMessageAttributes: message.MD5OfMessageAttributes}
		respStruct = ReceiveMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", ReceiveMessageResult{Message: &respMsg}, ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	} else {
		log.Println("No messages in Queue:", queueName)
		respStruct = ReceiveMessageResponse{xmlns: "http://queue.amazonaws.com/doc/2012-11-05/", Result: ReceiveMessageResult{}, Metadata: ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func DeleteMessage(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")

	// Retrieve FormValues required
	receiptHandle := req.FormValue("ReceiptHandle")
	queueUrl := req.FormValue("QueueUrl")

	uriSegments := strings.Split(queueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	log.Println("Deleting Message, Queue:", queueName, ", ReceiptHandle:", receiptHandle)

	// Find queue/message with the receipt handle and delete
	if Queues[queueName] != nil {
		for i, msg := range Queues[queueName].Messages {
			if msg.ReceiptHandle == receiptHandle {
				//Delete message from Q
				Queues[queueName].Messages = append(Queues[queueName].Messages[:i], Queues[queueName].Messages[i + 1:]...)
			}
		}
	} else {
		log.Println("Queue not found")
	}

	// Create, encode/xml and send response
	respStruct := DeleteMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func DeleteQueue(w http.ResponseWriter, req *http.Request) {
	// Sent response type
	w.Header().Set("Content-Type", "application/xml")

	// Retrieve FormValues required
	queueUrl := req.FormValue("QueueUrl")

	uriSegments := strings.Split(queueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	log.Println("Deleting Queue:", queueName)
	delete(Queues, queueName)

	// Create, encode/xml and send response
	respStruct := DeleteMessageResponse{"http://queue.amazonaws.com/doc/2012-11-05/", ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
