package gosns

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/p4tin/goaws/app"
	"github.com/p4tin/goaws/app/common"
	sqs "github.com/p4tin/goaws/app/gosqs"
)

type SnsErrorType struct {
	HttpError int
	Type      string
	Code      string
	Message   string
}

var SnsErrors map[string]SnsErrorType

type Subscription struct {
	TopicArn        string
	Protocol        string
	SubscriptionArn string
	EndPoint        string
	Raw             bool
}

type Topic struct {
	Name          string
	Arn           string
	Subscriptions []*Subscription
}

type (
	Protocol         string
	MessageStructure string
)

const (
	ProtocolSQS     Protocol = "sqs"
	ProtocolDefault Protocol = "default"
)

const (
	MessageStructureJSON MessageStructure = "json"
)

// Predefined errors
const (
	ErrNoDefaultElementInJSON = "Invalid parameter: Message Structure - No default entry in JSON message body"
)

var SyncTopics = struct {
	sync.RWMutex
	Topics map[string]*Topic
}{Topics: make(map[string]*Topic)}

func init() {
	SyncTopics.Topics = make(map[string]*Topic)

	SnsErrors = make(map[string]SnsErrorType)
	err1 := SnsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleNotificationService.NonExistentTopic", Message: "The specified topic does not exist for this wsdl version."}
	SnsErrors["TopicNotFound"] = err1
	err2 := SnsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleNotificationService.NonExistentSubscription", Message: "The specified subscription does not exist for this wsdl version."}
	SnsErrors["SubscriptionNotFound"] = err2
	err3 := SnsErrorType{HttpError: http.StatusBadRequest, Type: "Duplicate", Code: "AWS.SimpleNotificationService.TopicAlreadyExists", Message: "The specified topic already exists."}
	SnsErrors["TopicExists"] = err3
}

func ListTopics(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")

	respStruct := app.ListTopicsResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	uuid, _ := common.NewUUID()
	respStruct.Metadata = app.ResponseMetadata{RequestId: uuid}

	respStruct.Result.Topics.Member = make([]app.TopicArnResult, 0, 0)
	log.Println("Listing Topics")
	for _, topic := range SyncTopics.Topics {
		ta := app.TopicArnResult{TopicArn: topic.Arn}
		respStruct.Result.Topics.Member = append(respStruct.Result.Topics.Member, ta)
	}

	SendResponseBack(w, req, respStruct, content)
}

func CreateTopic(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	topicName := req.FormValue("Name")
	topicArn := ""
	if _, ok := SyncTopics.Topics[topicName]; ok {
		topicArn = SyncTopics.Topics[topicName].Arn
	} else {
		topicArn = "arn:aws:sns:local:000000000000:" + topicName

		log.Println("Creating Topic:", topicName)
		topic := &Topic{Name: topicName, Arn: topicArn}
		topic.Subscriptions = make([]*Subscription, 0, 0)
		SyncTopics.Lock()
		SyncTopics.Topics[topicName] = topic
		SyncTopics.Unlock()
	}
	uuid, _ := common.NewUUID()
	respStruct := app.CreateTopicResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.CreateTopicResult{TopicArn: topicArn}, app.ResponseMetadata{RequestId: uuid}}
	SendResponseBack(w, req, respStruct, content)
}

// aws --endpoint-url http://localhost:47194 sns subscribe --topic-arn arn:aws:sns:us-west-2:0123456789012:my-topic --protocol email --notification-endpoint my-email@example.com
func Subscribe(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	topicArn := req.FormValue("TopicArn")
	protocol := req.FormValue("Protocol")
	endpoint := req.FormValue("Endpoint")

	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]

	log.Println("Creating Subscription from", topicName, "to", endpoint, "using protocol", protocol)
	subscription := &Subscription{EndPoint: endpoint, Protocol: protocol, TopicArn: topicArn, Raw: false}
	subArn, _ := common.NewUUID()
	subArn = topicArn + ":" + subArn
	subscription.SubscriptionArn = subArn

	if SyncTopics.Topics[topicName] != nil {
		SyncTopics.Lock()
		isDuplicate := false
		// Duplicate check
		for _, subscription := range SyncTopics.Topics[topicName].Subscriptions {
			if subscription.EndPoint == endpoint && subscription.TopicArn == topicArn {
				isDuplicate = true
				subArn = subscription.SubscriptionArn
			}
		}
		if !isDuplicate {
			SyncTopics.Topics[topicName].Subscriptions = append(SyncTopics.Topics[topicName].Subscriptions, subscription)
		}
		SyncTopics.Unlock()

		//Create the response
		uuid, _ := common.NewUUID()
		respStruct := app.SubscribeResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.SubscribeResult{SubscriptionArn: subArn}, app.ResponseMetadata{RequestId: uuid}}
		SendResponseBack(w, req, respStruct, content)
	} else {
		createErrorResponse(w, req, "TopicNotFound")
	}
}

func ListSubscriptions(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")

	uuid, _ := common.NewUUID()
	respStruct := app.ListSubscriptionsResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	respStruct.Metadata.RequestId = uuid
	respStruct.Result.Subscriptions.Member = make([]app.TopicMemberResult, 0, 0)

	for _, topic := range SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			tar := app.TopicMemberResult{TopicArn: topic.Arn, Protocol: sub.Protocol,
				SubscriptionArn: sub.SubscriptionArn, Endpoint: sub.EndPoint}
			respStruct.Result.Subscriptions.Member = append(respStruct.Result.Subscriptions.Member, tar)
		}
	}

	SendResponseBack(w, req, respStruct, content)
}

func ListSubscriptionsByTopic(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	topicArn := req.FormValue("TopicArn")

	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]

	if topic, ok := SyncTopics.Topics[topicName]; ok {
		uuid, _ := common.NewUUID()
		respStruct := app.ListSubscriptionsByTopicResponse{}
		respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
		respStruct.Metadata.RequestId = uuid
		respStruct.Result.Subscriptions.Member = make([]app.TopicMemberResult, 0, 0)

		for _, sub := range topic.Subscriptions {
			tar := app.TopicMemberResult{TopicArn: topic.Arn, Protocol: sub.Protocol,
				SubscriptionArn: sub.SubscriptionArn, Endpoint: sub.EndPoint}
			respStruct.Result.Subscriptions.Member = append(respStruct.Result.Subscriptions.Member, tar)
		}
		SendResponseBack(w, req, respStruct, content)
	} else {
		createErrorResponse(w, req, "TopicNotFound")
	}
}

func SetSubscriptionAttributes(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	subsArn := req.FormValue("SubscriptionArn")
	Attribute := req.FormValue("AttributeName")
	Value := req.FormValue("AttributeValue")

	for _, topic := range SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subsArn {
				if Attribute == "RawMessageDelivery" {
					SyncTopics.Lock()
					if Value == "true" {
						sub.Raw = true
					} else {
						sub.Raw = false
					}
					SyncTopics.Unlock()
					//Good Response == return
					uuid, _ := common.NewUUID()
					respStruct := app.SetSubscriptionAttributesResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: uuid}}
					SendResponseBack(w, req, respStruct, content)
					return
				}
			}
		}
	}
	createErrorResponse(w, req, "SubscriptionNotFound")
}

func Unsubscribe(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	subArn := req.FormValue("SubscriptionArn")

	log.Println("Unsubcribing:", subArn)
	for _, topic := range SyncTopics.Topics {
		for i, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subArn {
				SyncTopics.Lock()

				copy(topic.Subscriptions[i:], topic.Subscriptions[i+1:])
				topic.Subscriptions[len(topic.Subscriptions)-1] = nil
				topic.Subscriptions = topic.Subscriptions[:len(topic.Subscriptions)-1]

				SyncTopics.Unlock()

				uuid, _ := common.NewUUID()
				respStruct := app.UnsubscribeResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: uuid}}
				SendResponseBack(w, req, respStruct, content)
				return
			}
		}
	}
	createErrorResponse(w, req, "SubscriptionNotFound")
}

func DeleteTopic(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	topicArn := req.FormValue("TopicArn")

	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]

	log.Println("Delete Topic - TopicName:", topicName)

	_, ok := SyncTopics.Topics[topicName]
	if ok {
		SyncTopics.Lock()
		delete(SyncTopics.Topics, topicName)
		SyncTopics.Unlock()
		uuid, _ := common.NewUUID()
		respStruct := app.DeleteTopicResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: uuid}}
		SendResponseBack(w, req, respStruct, content)
	} else {
		createErrorResponse(w, req, "TopicNotFound")
	}

}

// aws --endpoint-url http://localhost:47194 sns publish --topic-arn arn:aws:sns:yopa-local:000000000000:test1 --message "This is a test"
func Publish(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	topicArn := req.FormValue("TopicArn")
	subject := req.FormValue("Subject")
	messageBody := req.FormValue("Message")
	messageStructure := req.FormValue("MessageStructure")

	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]

	_, ok := SyncTopics.Topics[topicName]
	if ok {
		log.Println("Publish to Topic:", topicName)
		for _, subs := range SyncTopics.Topics[topicName].Subscriptions {
			if Protocol(subs.Protocol) == ProtocolSQS {
				queueUrl := subs.EndPoint
				uriSegments := strings.Split(queueUrl, "/")
				queueName := uriSegments[len(uriSegments)-1]
				if _, ok := sqs.SyncQueues.Queues[queueName]; ok {
					parts := strings.Split(queueName, ":")
					if len(parts) > 0 {
						queueName = parts[len(parts)-1]
					}

					msg := sqs.Message{}
					if subs.Raw == false {
						m, err := CreateMessageBody(messageBody, subject, topicArn, subs.Protocol, messageStructure)
						if err != nil {
							createErrorResponse(w, req, err.Error())
							return
						}

						msg.MessageBody = m
					} else {
						msg.MessageBody = []byte(messageBody)
					}
					msg.MD5OfMessageAttributes = common.GetMD5Hash("GoAws")
					msg.MD5OfMessageBody = common.GetMD5Hash(messageBody)
					msg.Uuid, _ = common.NewUUID()
					sqs.SyncQueues.Lock()
					sqs.SyncQueues.Queues[queueName].Messages = append(sqs.SyncQueues.Queues[queueName].Messages, msg)
					sqs.SyncQueues.Unlock()
					common.LogMessage(fmt.Sprintf("%s: Topic: %s(%s), Message: %s\n", time.Now().Format("2006-01-02 15:04:05"), topicName, queueName, msg.MessageBody))
				} else {
					common.LogMessage(fmt.Sprintf("%s: Queue %s does not exits, message discarded\n", time.Now().Format("2006-01-02 15:04:05"), queueName))
				}
			}
		}
	} else {
		createErrorResponse(w, req, "TopicNotFound")
		return
	}

	//Create the response
	msgId, _ := common.NewUUID()
	uuid, _ := common.NewUUID()
	respStruct := app.PublishResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.PublishResult{MessageId: msgId}, app.ResponseMetadata{RequestId: uuid}}
	SendResponseBack(w, req, respStruct, content)
}

type TopicMessage struct {
	Type      string
	MessageId string
	TopicArn  string
	Subject   string
	Message   string
	TimeStamp string
}

func CreateMessageBody(msg string, subject string, topicArn string, protocol string, messageStructure string) ([]byte, error) {
	msgId, _ := common.NewUUID()

	message := TopicMessage{}
	message.Type = "Notification"
	message.Subject = subject

	if MessageStructure(messageStructure) == MessageStructureJSON {
		m, err := extractMessageFromJSON(msg, protocol)
		if err != nil {
			return nil, err
		}
		message.Message = m
	} else {
		message.Message = msg
	}

	message.MessageId = msgId
	message.TopicArn = topicArn
	t := time.Now()
	message.TimeStamp = fmt.Sprintln(t.Format("2006-01-02T15:04:05:001Z"))

	byteMsg, _ := json.Marshal(message)
	return byteMsg, nil
}

func extractMessageFromJSON(msg string, protocol string) (string, error) {
	var msgWithProtocols map[string]string
	if err := json.Unmarshal([]byte(msg), &msgWithProtocols); err != nil {
		return "", err
	}

	defaultMsg, ok := msgWithProtocols[string(ProtocolDefault)]
	if !ok {
		return "", errors.New(ErrNoDefaultElementInJSON)
	}

	if m, ok := msgWithProtocols[protocol]; ok {
		return m, nil
	}

	return defaultMsg, nil
}

func createErrorResponse(w http.ResponseWriter, req *http.Request, err string) {
	er := SnsErrors[err]
	respStruct := app.ErrorResponse{app.ErrorResult{Type: er.Type, Code: er.Code, Message: er.Message, RequestId: "00000000-0000-0000-0000-000000000000"}}

	w.WriteHeader(er.HttpError)
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func SendResponseBack(w http.ResponseWriter, req *http.Request, respStruct interface{}, content string) {
	if content == "JSON" {
		w.Header().Set("Content-Type", "application/json")
	} else {
		w.Header().Set("Content-Type", "application/xml")
	}

	if content == "JSON" {
		enc := json.NewEncoder(w)
		if err := enc.Encode(respStruct); err != nil {
			log.Printf("error: %v\n", err)
		}
	} else {
		enc := xml.NewEncoder(w)
		enc.Indent("  ", "    ")
		if err := enc.Encode(respStruct); err != nil {
			log.Printf("error: %v\n", err)
		}
	}
}
