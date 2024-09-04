package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Admiral-Piett/goaws/app"

	log "github.com/sirupsen/logrus"
)

func NewCreateTopicRequest() *CreateTopicRequest {
	return &CreateTopicRequest{
		Attributes: TopicAttributes{
			FifoTopic:                 false,
			SignatureVersion:          1,
			TracingConfig:             "Active",
			ContentBasedDeduplication: false,
		},
	}
}

type CreateTopicRequest struct {
	Name string `json:"Name" schema:"Name"`

	// Goaws unsupports below properties currently.
	DataProtectionPolicy string            `json:"DataProtectionPolicy" schema:"DataProtectionPolicy"`
	Attributes           TopicAttributes   `json:"Attributes" schema:"Attributes"`
	Tags                 map[string]string `json:"Tags" schema:"Tags"`
}

// Ref: https://docs.aws.amazon.com/sns/latest/api/API_CreateTopic.html
type TopicAttributes struct {
	DeliveryPolicy            map[string]interface{} `json:"DeliveryPolicy"`            // NOTE: not implemented
	DisplayName               string                 `json:"DisplayName"`               // NOTE: not implemented
	FifoTopic                 bool                   `json:"FifoTopic"`                 // NOTE: not implemented
	Policy                    map[string]interface{} `json:"Policy"`                    // NOTE: not implemented
	SignatureVersion          StringToInt            `json:"SignatureVersion"`          // NOTE: not implemented
	TracingConfig             string                 `json:"TracingConfig"`             // NOTE: not implemented
	KmsMasterKeyId            string                 `json:"KmsMasterKeyId"`            // NOTE: not implemented
	ArchivePolicy             map[string]interface{} `json:"ArchivePolicy"`             // NOTE: not implemented
	BeginningArchiveTime      string                 `json:"BeginningArchiveTime"`      // NOTE: not implemented
	ContentBasedDeduplication bool                   `json:"ContentBasedDeduplication"` // NOTE: not implemented
}

func (r *CreateTopicRequest) SetAttributesFromForm(values url.Values) {
	for i := 1; true; i++ {
		nameKey := fmt.Sprintf("Attribute.%d.Name", i)
		attrName := values.Get(nameKey)
		if attrName == "" {
			break
		}

		valueKey := fmt.Sprintf("Attribute.%d.Value", i)
		attrValue := values.Get(valueKey)
		if attrValue == "" {
			continue
		}

		switch attrName {
		case "DeliveryPolicy":
			var tmp map[string]interface{}
			err := json.Unmarshal([]byte(attrValue), &tmp)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.DeliveryPolicy = tmp
		case "DisplayName":
			r.Attributes.DisplayName = attrValue
		case "FifoTopic":
			tmp, err := strconv.ParseBool(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.FifoTopic = tmp
		case "Policy":
			var tmp map[string]interface{}
			err := json.Unmarshal([]byte(attrValue), &tmp)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.Policy = tmp
		case "SignatureVersion":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.SignatureVersion = StringToInt(tmp)
		case "TracingConfig":
			r.Attributes.TracingConfig = attrValue
		case "KmsMasterKeyId":
			r.Attributes.KmsMasterKeyId = attrValue
		case "ArchivePolicy":
			var tmp map[string]interface{}
			err := json.Unmarshal([]byte(attrValue), &tmp)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.ArchivePolicy = tmp
		case "BeginningArchiveTime":
			r.Attributes.BeginningArchiveTime = attrValue
		case "ContentBasedDeduplication":
			tmp, err := strconv.ParseBool(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.ContentBasedDeduplication = tmp
		}
	}
}

func NewSubscribeRequest() *SubscribeRequest {
	return &SubscribeRequest{}
}

type SubscribeRequest struct {
	TopicArn   string                 `json:"TopicArn" schema:"TopicArn"`
	Endpoint   string                 `json:"Endpoint" schema:"Endpoint"`
	Protocol   string                 `json:"Protocol" schema:"Protocol"`
	Attributes SubscriptionAttributes `json:"Attributes"`
}

func (r *SubscribeRequest) SetAttributesFromForm(values url.Values) {
	for i := 1; true; i++ {
		nameKey := fmt.Sprintf("Attributes.entry.%d.key", i)
		attrName := values.Get(nameKey)
		if attrName == "" {
			break
		}

		valueKey := fmt.Sprintf("Attributes.entry.%d.value", i)
		attrValue := values.Get(valueKey)
		if attrValue == "" {
			continue
		}
		switch attrName {
		case "RawMessageDelivery":
			tmp, err := strconv.ParseBool(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.RawMessageDelivery = tmp
		case "FilterPolicy":
			var tmp map[string][]string
			err := json.Unmarshal([]byte(attrValue), &tmp)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.FilterPolicy = tmp
		}
	}
	return
}

type SubscriptionAttributes struct {
	FilterPolicy       app.FilterPolicy `json:"FilterPolicy" schema:"FilterPolicy"`
	RawMessageDelivery bool             `json:"RawMessageDelivery" schema:"RawMessageDelivery"`
	//DeliveryPolicy      map[string]interface{} `json:"DeliveryPolicy" schema:"DeliveryPolicy"`
	//FilterPolicyScope   string                 `json:"FilterPolicyScope" schema:"FilterPolicyScope"`
	//RedrivePolicy       RedrivePolicy          `json:"RedrivePolicy" schema:"RawMessageDelivery"`
	//SubscriptionRoleArn string                 `json:"SubscriptionRoleArn" schema:"SubscriptionRoleArn"`
	//ReplayPolicy        string                 `json:"ReplayPolicy" schema:"ReplayPolicy"`
	//ReplayStatus        string                 `json:"ReplayStatus" schema:"ReplayStatus"`
}

func NewUnsubscribeRequest() *UnsubscribeRequest {
	return &UnsubscribeRequest{}
}

type UnsubscribeRequest struct {
	SubscriptionArn string `json:"SubscriptionArn" schema:"SubscriptionArn"`
}

func (r *UnsubscribeRequest) SetAttributesFromForm(values url.Values) {}

func NewPublishRequest() *PublishRequest {
	return &PublishRequest{}
}

type PublishRequest struct {
	Message                string                           `json:"Message" schema:"Message"`
	MessageAttributes      map[string]MessageAttributeValue `json:"MessageAttributes" schema:"MessageAttributes"`
	MessageDeduplicationId string                           `json:"MessageDeduplicationId" schema:"MessageDeduplicationId"` // Not implemented
	MessageGroupId         string                           `json:"MessageGroupId" schema:"MessageGroupId"`                 // Not implemented
	MessageStructure       string                           `json:"MessageStructure" schema:"MessageStructure"`
	PhoneNumber            string                           `json:"PhoneNumber" schema:"PhoneNumber"` // Not implemented
	Subject                string                           `json:"Subject" schema:"Subject"`
	TargetArn              string                           `json:"TargetArn" schema:"TargetArn"` // Not implemented
	TopicArn               string                           `json:"TopicArn" schema:"TopicArn"`
}

func (r *PublishRequest) SetAttributesFromForm(values url.Values) {
	for i := 1; true; i++ {
		nameKey := fmt.Sprintf("MessageAttributes.entry.%d.Name", i)
		name := values.Get(nameKey)
		if name == "" {
			break
		}

		dataTypeKey := fmt.Sprintf("MessageAttributes.entry.%d.Value.DataType", i)
		dataType := values.Get(dataTypeKey)
		if dataType == "" {
			log.Warnf("DataType of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
			continue
		}

		stringValue := values.Get(fmt.Sprintf("MessageAttributes.entry.%d.Value.StringValue", i))
		binaryValue := values.Get(fmt.Sprintf("MessageAttributes.entry.%d.Value.BinaryValue", i))

		r.MessageAttributes[name] = MessageAttributeValue{
			DataType:    dataType,
			StringValue: stringValue,
			BinaryValue: binaryValue,
		}
	}
}

// ListTopics

func NewListTopicsRequest() *ListTopicsRequest {
	return &ListTopicsRequest{}
}

type ListTopicsRequest struct {
	NextToken string `json:"NextToken" schema:"NextToken"` // not implemented
}

func (r *ListTopicsRequest) SetAttributesFromForm(values url.Values) {}

// DeleteTopicV1

func NewDeleteTopicRequest() *DeleteTopicRequest {
	return &DeleteTopicRequest{}
}

type DeleteTopicRequest struct {
	TopicArn string `json:"TopicArn" schema:"TopicArn"`
}

func (r *DeleteTopicRequest) SetAttributesFromForm(values url.Values) {}

// ListSubscriptionsV1

func NewListSubscriptionsRequest() *ListSubscriptionsRequest {
	return &ListSubscriptionsRequest{}
}

type ListSubscriptionsRequest struct {
	NextToken string `json:"NextToken" schema:"NextToken"` // not implemented
}

func (r *ListSubscriptionsRequest) SetAttributesFromForm(values url.Values) {}

// Get Subscription Attributes V1
func NewGetSubscriptionAttributesRequest() *GetSubscriptionAttributesRequest {
	return &GetSubscriptionAttributesRequest{}
}

type GetSubscriptionAttributesRequest struct {
	SubscriptionArn string `json:"SubscriptionArn" schema:"SubscriptionArn"`
}

func (r *GetSubscriptionAttributesRequest) SetAttributesFromForm(values url.Values) {}

// SetSubscriptionAttributes

func NewSetSubscriptionAttributesRequest() *SetSubscriptionAttributesRequest {
	return &SetSubscriptionAttributesRequest{}
}

// Ref: https://docs.aws.amazon.com/sns/latest/api/API_SetSubscriptionAttributes.html
type SetSubscriptionAttributesRequest struct {
	SubscriptionArn string `json:"SubscriptionArn" schema:"SubscriptionArn"`
	AttributeName   string `json:"AttributeName" schema:"AttributeName"`
	AttributeValue  string `json:"AttributeValue" schema:"AttributeValue"`
}

func (r *SetSubscriptionAttributesRequest) SetAttributesFromForm(values url.Values) {}

// List Subscriptions By Topic

func NewListSubscriptionsByTopicRequest() *ListSubscriptionsByTopicRequest {
	return &ListSubscriptionsByTopicRequest{}
}

type ListSubscriptionsByTopicRequest struct {
	NextToken string `json:"NextToken" schema:"NextToken"` // not implemented
	TopicArn  string `json:"TopicArn" schema:"TopicArn"`
}

func (r *ListSubscriptionsByTopicRequest) SetAttributesFromForm(values url.Values) {}

// Confirm Subscription V1

func NewConfirmSubscriptionRequest() *ConfirmSubscriptionRequest {
	return &ConfirmSubscriptionRequest{}
}

type ConfirmSubscriptionRequest struct {
	AuthenticateOnUnsubscribe bool   `json:"AuthenticateOnUnsubscribe" schema:"AuthenticateOnUnsubscribe"` // not implemented
	TopicArn                  string `json:"TopicArn" schema:"TopicArn"`
	Token                     string `json:"Token" schema:"Token"`
}

func (r *ConfirmSubscriptionRequest) SetAttributesFromForm(values url.Values) {}
