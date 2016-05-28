package gosns

import (
	"net/http"
	"fmt"
	"log"
	"sync"
	"strings"
	"encoding/xml"

	"github.com/p4tin/goaws/common"
	sqs "github.com/p4tin/goaws/gosqs"
	"time"
	"gopkg.in/square/go-jose.v1/json"
)

type SnsErrorType struct {
	HttpError int
	Type string
	Code string
	Message string
}

var SnsErrors map[string]SnsErrorType

type Subscription struct {
	TopicArn 	string
	Protocol 	string
	SubscriptionArn string
	EndPoint 	string
	Raw 		bool
}

type Topic struct {
	Name 		string
	Arn 		string
	Subscriptions 	[]*Subscription
}


var SyncTopics = struct{
	sync.RWMutex
	Topics map[string]*Topic
}{Topics: make(map[string]*Topic)}

func init() {
	SyncTopics.Topics = make(map[string]*Topic)

	SnsErrors = make(map[string]SnsErrorType)
	err1 := SnsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleNotificationService.NonExistentTopic" , Message:"The specified topic does not exist for this wsdl version."}
	SnsErrors["TopicNotFound"] = err1
	err2 := SnsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleNotificationService.NonExistentSubscription" , Message:"The specified subscription does not exist for this wsdl version."}
	SnsErrors["SubscriptionNotFound"] = err2
	err3 := SnsErrorType{HttpError: http.StatusBadRequest, Type: "Duplicate", Code: "AWS.SimpleNotificationService.TopicAlreadyExists" , Message:"The specified topic already exists."}
	SnsErrors["TopicExists"] = err3
}

func ListTopics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	respStruct := ListTopicsResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	uuid, _ := common.NewUUID()
	respStruct.Metadata = ResponseMetadata{RequestId: uuid}

	respStruct.Result.Topics.Member = make([]TopicArnResult, 0, 0)
	log.Println("Listing Topics")
	for _, topic := range SyncTopics.Topics {
		fmt.Println(topic)
		ta := TopicArnResult{TopicArn: topic.Arn}
		respStruct.Result.Topics.Member = append(respStruct.Result.Topics.Member, ta)
	}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func CreateTopic(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	topicName := req.FormValue("Name")

	if _, ok := SyncTopics.Topics[topicName] ; ok {
		createErrorResponse(w, req, "TopicExists")
		return
	}

	topicArn := "arn:aws:sns:local:000000000000:" + topicName

	log.Println("Creating Topic:", topicName)
	topic := &Topic{Name: topicName, Arn: topicArn}
	topic.Subscriptions = make([]*Subscription, 0 ,0)
	SyncTopics.RLock()
	SyncTopics.Topics[topicName] = topic
	SyncTopics.RUnlock()

	uuid, _ := common.NewUUID()
	respStruct := CreateTopicResponse{"http://queue.amazonaws.com/doc/2012-11-05/", CreateTopicResult{TopicArn: topicArn}, ResponseMetadata{RequestId: uuid}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

// aws --endpoint-url http://localhost:47194 sns subscribe --topic-arn arn:aws:sns:us-west-2:0123456789012:my-topic --protocol email --notification-endpoint my-email@example.com
func Subscribe(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	topicArn := req.FormValue("TopicArn")
	protocol := req.FormValue("Protocol")
	endpoint := req.FormValue("Endpoint")

	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]

	log.Println("Creating Subscription from", topicName, "to", endpoint, "using protocol", protocol)
	subscription := &Subscription{EndPoint:endpoint, Protocol: protocol, TopicArn: topicArn, Raw: false}
	subArn, _ := common.NewUUID()
	subArn = topicArn + ":" + subArn
	subscription.SubscriptionArn = subArn

	if SyncTopics.Topics[topicName] != nil {
		SyncTopics.Lock()
		SyncTopics.Topics[topicName].Subscriptions = append(SyncTopics.Topics[topicName].Subscriptions, subscription)
		SyncTopics.Unlock()

		//Create the response
		uuid, _ := common.NewUUID()
		respStruct := SubscribeResponse{"http://queue.amazonaws.com/doc/2012-11-05/", SubscribeResult{SubscriptionArn: subArn}, ResponseMetadata{RequestId: uuid}}
		enc := xml.NewEncoder(w)
		enc.Indent("  ", "    ")
		if err := enc.Encode(respStruct); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	} else {
		// Return error
	}
}

func ListSubscriptions(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	uuid, _ := common.NewUUID()
	respStruct := ListSubscriptionsResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	respStruct.Metadata.RequestId = uuid
	respStruct.Result.Subscriptions.Member = make([]TopicMemberResult, 0, 0)

	for _, topic := range SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			tar := TopicMemberResult{TopicArn: topic.Arn, Protocol: sub.Protocol,
				SubscriptionArn: sub.SubscriptionArn, Endpoint: sub.EndPoint}
			respStruct.Result.Subscriptions.Member = append(respStruct.Result.Subscriptions.Member, tar)
		}
	}


	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func SetSubscriptionAttributes(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	subsArn := req.FormValue("SubscriptionArn")
	Attribute := req.FormValue("AttributeName")
	Value := req.FormValue("AttributeValue")

	for _, topic := range SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subsArn {
				if Attribute == "RawMessageDelivery" {
					log.Println("Before - Subs:", sub.EndPoint, "Raw:", sub.Raw)
					SyncTopics.Lock()
					if Value == "true" {
						sub.Raw = true
					} else {
						sub.Raw = false
					}
					SyncTopics.Unlock()
					log.Println("After - Subs:", sub.EndPoint, "Raw:", sub.Raw)
					//Good Response == return
					uuid, _ := common.NewUUID()
					respStruct := SetSubscriptionAttributesResponse{"http://queue.amazonaws.com/doc/2012-11-05/", ResponseMetadata{RequestId: uuid}}
					enc := xml.NewEncoder(w)
					enc.Indent("  ", "    ")
					if err := enc.Encode(respStruct); err != nil {
						fmt.Printf("error: %v\n", err)
					}
					return
				}
			}
		}
	}
	createErrorResponse(w, req, "SubscriptionNotFound")
}

func Unsubscribe(w http.ResponseWriter, req *http.Request) {

}

func DeleteTopic(w http.ResponseWriter, req *http.Request) {

}

// aws --endpoint-url http://localhost:47194 sns publish --topic-arn arn:aws:sns:yopa-local:000000000000:test1 --message "This is a test"
func Publish(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	topicArn := req.FormValue("TopicArn")
	messageBody := req.FormValue("Message")

	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]

	for _, subs := range SyncTopics.Topics[topicName].Subscriptions {
		if subs.Protocol == "sqs" {
			queueUrl := subs.EndPoint
			uriSegments := strings.Split(queueUrl, "/")
			queueName := uriSegments[len(uriSegments) - 1]

			msg := sqs.Message{}
			log.Println("Subscript:", subs.EndPoint, "Raw: ", subs.Raw)
			if subs.Raw == false {
				msg.MessageBody = CreateMessageBody(messageBody, topicArn)
			} else {
				msg.MessageBody = []byte(messageBody)
			}
			msg.MD5OfMessageAttributes = common.GetMD5Hash("GoAws")
			msg.MD5OfMessageBody = common.GetMD5Hash(messageBody)
			msg.Uuid, _ = common.NewUUID()
			sqs.SyncQueues.Lock()
			sqs.SyncQueues.Queues[queueName].Messages = append(sqs.SyncQueues.Queues[queueName].Messages, msg)
			sqs.SyncQueues.Unlock()
		}
	}

	//Create the response
	msgId, _ := common.NewUUID()
	uuid, _ := common.NewUUID()
	respStruct := PublishResponse{"http://queue.amazonaws.com/doc/2012-11-05/", PublishResult{MessageId: msgId}, ResponseMetadata{RequestId: uuid}}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

type TopicMessage struct {
	Type string
	MessageId string
	TopicArn string
	Subject string
	Message string
	TimeStamp string
}

func CreateMessageBody(msg string, topicArn string) []byte {
	msgId, _ := common.NewUUID()

	message := TopicMessage{}
	message.Type = "Notification"
	message.Message = msg
	message.MessageId = msgId
	message.TopicArn = topicArn
	t := time.Now()
	message.TimeStamp = fmt.Sprintln(t.Format("2006-01-02T15:04:05:001Z"))

	byteMsg, _ := json.Marshal(message)
	return byteMsg
}


func createErrorResponse(w http.ResponseWriter, req *http.Request, err string) {
	er := SnsErrors[err]
	respStruct := ErrorResponse{ErrorResult{Type: er.Type, Code: er.Code, Message: er.Message, RequestId: "00000000-0000-0000-0000-000000000000"}}

	w.WriteHeader(er.HttpError)
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
