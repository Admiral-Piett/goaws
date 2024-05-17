package models

import (
	"encoding/json"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/aws/aws-sdk-go-v2/aws"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// NOTE: Every response in here MUST implement the `AbstractResponseBody` interface in order to be used
//  in `encodeResponse`

/*** Error Responses ***/
type ErrorResult struct {
	Type    string `xml:"Type,omitempty"`
	Code    string `xml:"Code,omitempty"`
	Message string `xml:"Message,omitempty"`
}

type ErrorResponse struct {
	Result    ErrorResult `xml:"Error"`
	RequestId string      `xml:"RequestId"`
}

func (r ErrorResponse) GetResult() interface{} {
	return r.Result
}

func (r ErrorResponse) GetRequestId() string {
	return r.RequestId
}

/*** Receive Message Response */
type ReceiveMessageResult struct {
	Messages []*ResultMessage `json:"Messages" xml:"Message,omitempty"`
}

type ReceiveMessageResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Result   ReceiveMessageResult `xml:"ReceiveMessageResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r ReceiveMessageResponse) GetResult() interface{} {
	return r.Result
}

func (r ReceiveMessageResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type ResultMessage struct {
	MessageId              string                    `xml:"MessageId,omitempty"`
	ReceiptHandle          string                    `xml:"ReceiptHandle,omitempty"`
	MD5OfBody              string                    `xml:"MD5OfBody,omitempty"`
	Body                   []byte                    `xml:"Body,omitempty"`
	MD5OfMessageAttributes string                    `xml:"MD5OfMessageAttributes,omitempty"`
	MessageAttributes      []*ResultMessageAttribute `xml:"MessageAttribute,omitempty"`
	Attributes             []*ResultAttribute        `xml:"Attribute,omitempty"`
}

// MarshalJSON first converts the ResultMessage to the shape which the SDKs
// expect. When receiving a response from the JSON API, it apparently expects
// Attributes and MessageAttributes to be maps, rather than the former slice
// shape.
func (r *ResultMessage) MarshalJSON() ([]byte, error) {
	m := &sqstypes.Message{
		MessageId:              &r.MessageId,
		ReceiptHandle:          &r.ReceiptHandle,
		MD5OfBody:              &r.MD5OfBody,
		Body:                   aws.String(string(r.Body)),
		MD5OfMessageAttributes: &r.MD5OfMessageAttributes,
		Attributes:             map[string]string{},
		MessageAttributes:      map[string]sqstypes.MessageAttributeValue{},
	}

	for _, attr := range r.Attributes {
		m.Attributes[attr.Name] = attr.Value
	}

	for _, attr := range r.MessageAttributes {
		m.MessageAttributes[attr.Name] = sqstypes.MessageAttributeValue{
			DataType:    &attr.Value.DataType,
			StringValue: &attr.Value.StringValue,
			BinaryValue: []byte(attr.Value.BinaryValue),
		}
	}

	return json.Marshal(m)
}

type ResultMessageAttributeValue struct {
	DataType    string `xml:"DataType,omitempty"`
	StringValue string `xml:"StringValue,omitempty"`
	BinaryValue string `xml:"BinaryValue,omitempty"`
}

type ResultMessageAttribute struct {
	Name  string                       `xml:"Name,omitempty"`
	Value *ResultMessageAttributeValue `xml:"Value,omitempty"`
}

type ResultAttribute struct {
	Name  string `xml:"Name,omitempty"`
	Value string `xml:"Value,omitempty"`
}

type ChangeMessageVisibilityResult struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r ChangeMessageVisibilityResult) GetResult() interface{} {
	return nil
}

func (r ChangeMessageVisibilityResult) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Create Queue Response */
type CreateQueueResult struct {
	QueueUrl string `json:"QueueUrl" xml:"QueueUrl"`
}

type CreateQueueResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Result   CreateQueueResult    `xml:"CreateQueueResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r CreateQueueResponse) GetResult() interface{} {
	return r.Result
}

func (r CreateQueueResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** List Queues Response */
type ListQueuesResult struct {
	// NOTE: the old XML sdks depend on QueueUrl, and the new JSON ones need QueueUrls
	QueueUrls []string `json:"QueueUrls" xml:"QueueUrl"`
}

type ListQueuesResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Result   ListQueuesResult     `xml:"ListQueuesResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r ListQueuesResponse) GetResult() interface{} {
	return r.Result
}

func (r ListQueuesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Get Queue Attributes ***/
type Attribute struct {
	Name  string `xml:"Name,omitempty"`
	Value string `xml:"Value,omitempty"`
}

type GetQueueAttributesResult struct {
	/* VisibilityTimeout, DelaySeconds, ReceiveMessageWaitTimeSeconds, ApproximateNumberOfMessages
	   ApproximateNumberOfMessagesNotVisible, CreatedTimestamp, LastModifiedTimestamp, QueueArn */
	Attrs []Attribute `xml:"Attribute,omitempty"`
}

type GetQueueAttributesResponse struct {
	Xmlns    string                   `xml:"xmlns,attr,omitempty"`
	Result   GetQueueAttributesResult `xml:"GetQueueAttributesResult"`
	Metadata app.ResponseMetadata     `xml:"ResponseMetadata,omitempty"`
}

func (r GetQueueAttributesResponse) GetResult() interface{} {
	result := map[string]string{}
	for _, attr := range r.Result.Attrs {
		result[attr.Name] = attr.Value
	}
	return map[string]map[string]string{"Attributes": result}
}

func (r GetQueueAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Send Message Response */
type SendMessageResult struct {
	MD5OfMessageAttributes string `xml:"MD5OfMessageAttributes"`
	MD5OfMessageBody       string `xml:"MD5OfMessageBody"`
	MessageId              string `xml:"MessageId"`
	SequenceNumber         string `xml:"SequenceNumber"`
}

type SendMessageResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Result   SendMessageResult    `xml:"SendMessageResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r SendMessageResponse) GetResult() interface{} {
	return r.Result
}

func (r SendMessageResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Message Response */
type DeleteMessageResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r DeleteMessageResponse) GetResult() interface{} {
	return nil
}

func (r DeleteMessageResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type SetQueueAttributesResponse struct {
	Xmlns    string               `xml:"xmlns,attr,omitempty"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata,omitempty"`
}

func (r SetQueueAttributesResponse) GetResult() interface{} {
	return nil
}

func (r SetQueueAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}