package gosns

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/p4tin/goaws/app"
	"github.com/p4tin/goaws/app/common"
	log "github.com/sirupsen/logrus"
)

func init() {
	app.SyncTopics.Topics = make(map[string]*app.Topic)

	app.SnsErrors = make(map[string]app.SnsErrorType)
	err1 := app.SnsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleNotificationService.NonExistentTopic", Message: "The specified topic does not exist for this wsdl version."}
	app.SnsErrors["TopicNotFound"] = err1
	err2 := app.SnsErrorType{HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleNotificationService.NonExistentSubscription", Message: "The specified subscription does not exist for this wsdl version."}
	app.SnsErrors["SubscriptionNotFound"] = err2
	err3 := app.SnsErrorType{HttpError: http.StatusBadRequest, Type: "Duplicate", Code: "AWS.SimpleNotificationService.TopicAlreadyExists", Message: "The specified topic already exists."}
	app.SnsErrors["TopicExists"] = err3
	err4 := app.SnsErrorType{HttpError: http.StatusBadRequest, Type: "InvalidParameter", Code: "AWS.SimpleNotificationService.ValidationError", Message: "The input fails to satisfy the constraints specified by an AWS service."}
	app.SnsErrors["ValidationError"] = err4
}

func ListTopics(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")

	respStruct := app.ListTopicsResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	uuid, _ := common.NewUUID()
	respStruct.Metadata = app.ResponseMetadata{RequestId: uuid}

	respStruct.Result.Topics.Member = make([]app.TopicArnResult, 0, 0)
	log.Println("Listing Topics")
	for _, topic := range app.SyncTopics.Topics {
		ta := app.TopicArnResult{TopicArn: topic.Arn}
		respStruct.Result.Topics.Member = append(respStruct.Result.Topics.Member, ta)
	}

	SendResponseBack(w, req, respStruct, content)
}

func CreateTopic(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	topicName := req.FormValue("Name")
	topicArn := ""
	if _, ok := app.SyncTopics.Topics[topicName]; ok {
		topicArn = app.SyncTopics.Topics[topicName].Arn
	} else {
		topicArn = "arn:aws:sns:" + app.CurrentEnvironment.Region + ":000000000000:" + topicName

		log.Println("Creating Topic:", topicName)
		topic := &app.Topic{Name: topicName, Arn: topicArn}
		topic.Subscriptions = make([]*app.Subscription, 0, 0)
		app.SyncTopics.Lock()
		app.SyncTopics.Topics[topicName] = topic
		app.SyncTopics.Unlock()
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
	subscription := &app.Subscription{EndPoint: endpoint, Protocol: protocol, TopicArn: topicArn, Raw: false}
	subArn, _ := common.NewUUID()
	subArn = topicArn + ":" + subArn
	subscription.SubscriptionArn = subArn

	if app.SyncTopics.Topics[topicName] != nil {
		app.SyncTopics.Lock()
		isDuplicate := false
		// Duplicate check
		for _, subscription := range app.SyncTopics.Topics[topicName].Subscriptions {
			if subscription.EndPoint == endpoint && subscription.TopicArn == topicArn {
				isDuplicate = true
				subArn = subscription.SubscriptionArn
			}
		}
		if !isDuplicate {
			app.SyncTopics.Topics[topicName].Subscriptions = append(app.SyncTopics.Topics[topicName].Subscriptions, subscription)
		}
		app.SyncTopics.Unlock()

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

	for _, topic := range app.SyncTopics.Topics {
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

	if topic, ok := app.SyncTopics.Topics[topicName]; ok {
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

	for _, topic := range app.SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subsArn {
				if Attribute == "RawMessageDelivery" {
					app.SyncTopics.Lock()
					if Value == "true" {
						sub.Raw = true
					} else {
						sub.Raw = false
					}
					app.SyncTopics.Unlock()
					//Good Response == return
					uuid, _ := common.NewUUID()
					respStruct := app.SetSubscriptionAttributesResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: uuid}}
					SendResponseBack(w, req, respStruct, content)
					return
				}

				if Attribute == "FilterPolicy" {
					filterPolicy := &app.FilterPolicy{}
					err := json.Unmarshal([]byte(Value), filterPolicy)
					if err != nil {
						createErrorResponse(w, req, "ValidationError")
						return
					}

					app.SyncTopics.Lock()
					sub.FilterPolicy = filterPolicy
					app.SyncTopics.Unlock()

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

func GetSubscriptionAttributes(w http.ResponseWriter, req *http.Request) {

	content := req.FormValue("ContentType")
	subsArn := req.FormValue("SubscriptionArn")

	for _, topic := range app.SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subsArn {

				entries := make([]app.SubscriptionAttributeEntry, 0, 0)
				entry := app.SubscriptionAttributeEntry{Key: "SubscriptionArn", Value: sub.SubscriptionArn}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "Protocol", Value: sub.Protocol}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "Endpoint", Value: sub.EndPoint}
				entries = append(entries, entry)

				if sub.FilterPolicy != nil {
					filterPolicyBytes, _ := json.Marshal(sub.FilterPolicy)
					entry = app.SubscriptionAttributeEntry{Key: "FilterPolicy", Value: string(filterPolicyBytes)}
					entries = append(entries, entry)
				}

				result := app.GetSubscriptionAttributesResult{SubscriptionAttributes: app.SubscriptionAttributes{Entries: entries}}
				uuid, _ := common.NewUUID()
				respStruct := app.GetSubscriptionAttributesResponse{"http://sns.amazonaws.com/doc/2010-03-31", result, app.ResponseMetadata{RequestId: uuid}}

				SendResponseBack(w, req, respStruct, content)

				return

			}
		}
	}
	createErrorResponse(w, req, "SubscriptionNotFound")
}

func Unsubscribe(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	subArn := req.FormValue("SubscriptionArn")

	log.Println("Unsubcribing:", subArn)
	for _, topic := range app.SyncTopics.Topics {
		for i, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subArn {
				app.SyncTopics.Lock()

				copy(topic.Subscriptions[i:], topic.Subscriptions[i+1:])
				topic.Subscriptions[len(topic.Subscriptions)-1] = nil
				topic.Subscriptions = topic.Subscriptions[:len(topic.Subscriptions)-1]

				app.SyncTopics.Unlock()

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

	_, ok := app.SyncTopics.Topics[topicName]
	if ok {
		app.SyncTopics.Lock()
		delete(app.SyncTopics.Topics, topicName)
		app.SyncTopics.Unlock()
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
	messageAttributes := getMessageAttributesFromRequest(req)

	arnSegments := strings.Split(topicArn, ":")
	topicName := arnSegments[len(arnSegments)-1]

	_, ok := app.SyncTopics.Topics[topicName]
	if ok {
		log.Println("Publish to Topic:", topicName)
		for _, subs := range app.SyncTopics.Topics[topicName].Subscriptions {
			if app.Protocol(subs.Protocol) == app.ProtocolSQS {

				if subs.FilterPolicy != nil && !subs.FilterPolicy.IsSatisfiedBy(messageAttributes) {
					continue
				}

				endPoint := subs.EndPoint
				uriSegments := strings.Split(endPoint, "/")
				queueName := uriSegments[len(uriSegments)-1]
				arnSegments := strings.Split(queueName, ":")
				queueName = arnSegments[len(arnSegments)-1]

				if _, ok := app.SyncQueues.Queues[queueName]; ok {
					msg := app.Message{}

					if subs.Raw == false {
						m, err := CreateMessageBody(messageBody, subject, topicArn, subs.Protocol, messageStructure)
						if err != nil {
							createErrorResponse(w, req, err.Error())
							return
						}

						msg.MessageBody = m
					} else {
						msg.MessageAttributes = messageAttributes
						msg.MD5OfMessageAttributes = common.HashAttributes(messageAttributes)
						msg.MessageBody = []byte(messageBody)
					}

					msg.MD5OfMessageBody = common.GetMD5Hash(messageBody)
					msg.Uuid, _ = common.NewUUID()
					app.SyncQueues.Lock()
					app.SyncQueues.Queues[queueName].Messages = append(app.SyncQueues.Queues[queueName].Messages, msg)
					app.SyncQueues.Unlock()

					common.LogMessage(fmt.Sprintf("%s: Topic: %s(%s), Message: %s\n", time.Now().Format("2006-01-02 15:04:05"), topicName, queueName, msg.MessageBody))
				} else {
					common.LogMessage(fmt.Sprintf("%s: Queue %s does not exist, message discarded\n", time.Now().Format("2006-01-02 15:04:05"), queueName))
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

func getMessageAttributesFromRequest(req *http.Request) map[string]app.MessageAttributeValue {
	attributes := make(map[string]app.MessageAttributeValue)

	for i := 1; true; i++ {
		name := req.FormValue(fmt.Sprintf("MessageAttributes.entry.%d.Name", i))
		if name == "" {
			break
		}

		dataType := req.FormValue(fmt.Sprintf("MessageAttributes.entry.%d.Value.DataType", i))
		if dataType == "" {
			log.Warnf("DataType of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
			continue
		}

		// StringListValue and BinaryListValue is currently not implemented
		for _, valueKey := range [...]string{"StringValue", "BinaryValue"} {
			value := req.FormValue(fmt.Sprintf("MessageAttributes.entry.%d.Value.%s", i, valueKey))
			if value != "" {
				attributes[name] = app.MessageAttributeValue{name, dataType, value, valueKey}
			}
		}

		if _, ok := attributes[name]; !ok {
			log.Warnf("StringValue or BinaryValue of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
		}
	}

	return attributes
}

type TopicMessage struct {
	Type      string
	MessageId string
	TopicArn  string
	Subject   string
	Message   string
	Timestamp string
}

func CreateMessageBody(msg string, subject string, topicArn string, protocol string, messageStructure string) ([]byte, error) {
	msgId, _ := common.NewUUID()

	message := TopicMessage{}
	message.Type = "Notification"
	message.Subject = subject

	if app.MessageStructure(messageStructure) == app.MessageStructureJSON {
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
	message.Timestamp = fmt.Sprint(t.Format("2006-01-02T15:04:05.001Z"))

	byteMsg, _ := json.Marshal(message)
	return byteMsg, nil
}

func extractMessageFromJSON(msg string, protocol string) (string, error) {
	var msgWithProtocols map[string]string
	if err := json.Unmarshal([]byte(msg), &msgWithProtocols); err != nil {
		return "", err
	}

	defaultMsg, ok := msgWithProtocols[string(app.ProtocolDefault)]
	if !ok {
		return "", errors.New(app.ErrNoDefaultElementInJSON)
	}

	if m, ok := msgWithProtocols[protocol]; ok {
		return m, nil
	}

	return defaultMsg, nil
}

func createErrorResponse(w http.ResponseWriter, req *http.Request, err string) {
	er := app.SnsErrors[err]
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
