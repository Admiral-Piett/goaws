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
// QueueAttributes and MessageAttributes to be maps, rather than the former slice
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

/*** Get Queue QueueAttributes ***/
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

/*** Get Queue Url Response */
type GetQueueUrlResult struct {
	QueueUrl string `xml:"QueueUrl,omitempty"`
}

type GetQueueUrlResponse struct {
	Xmlns    string               `xml:"xmlns,attr,omitempty"`
	Result   GetQueueUrlResult    `xml:"GetQueueUrlResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata,omitempty"`
}

func (r GetQueueUrlResponse) GetResult() interface{} {
	return r.Result
}

func (r GetQueueUrlResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type SendMessageBatchResultEntry struct {
	Id                     string `xml:"Id"`
	MessageId              string `xml:"MessageId"`
	MD5OfMessageBody       string `xml:"MD5OfMessageBody,omitempty"`
	MD5OfMessageAttributes string `xml:"MD5OfMessageAttributes,omitempty"`
	SequenceNumber         string `xml:"SequenceNumber"`
}

/*** Send Message Batch Response */
type SendMessageBatchResponse struct {
	Xmlns    string                 `xml:"xmlns,attr,omitempty"`
	Result   SendMessageBatchResult `xml:"SendMessageBatchResult"`
	Metadata app.ResponseMetadata   `xml:"ResponseMetadata,omitempty"`
}

type SendMessageBatchResult struct {
	Entry []SendMessageBatchResultEntry `xml:"SendMessageBatchResultEntry"`
	Error []BatchResultErrorEntry       `xml:"BatchResultErrorEntry,omitempty"`
}

func (r SendMessageBatchResponse) GetResult() interface{} {
	return r.Result
}

func (r SendMessageBatchResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type BatchResultErrorEntry struct {
	Code        string `xml:"Code"`
	Id          string `xml:"Id"`
	Message     string `xml:"Message,omitempty"`
	SenderFault bool   `xml:"SenderFault"`
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

/*** Purge Queue Response */
type PurgeQueueResponse struct {
	Xmlns    string               `xml:"xmlns,attr,omitempty"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata,omitempty"`
}

func (r PurgeQueueResponse) GetResult() interface{} {
	return nil
}

func (r PurgeQueueResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Queue Response */
type DeleteQueueResponse struct {
	Xmlns    string               `xml:"xmlns,attr,omitempty"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata,omitempty"`
}

func (r DeleteQueueResponse) GetResult() interface{} {
	return nil
}

func (r DeleteQueueResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Create Topic Response */
type CreateTopicResult struct {
	TopicArn string `xml:"TopicArn"`
}

type CreateTopicResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Result   CreateTopicResult    `xml:"CreateTopicResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r CreateTopicResponse) GetResult() interface{} {
	return r.Result
}

func (r CreateTopicResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Create Subscription ***/
type SubscribeResult struct {
	SubscriptionArn string `xml:"SubscriptionArn"`
}

type SubscribeResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Result   SubscribeResult      `xml:"SubscribeResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r SubscribeResponse) GetResult() interface{} {
	return r.Result
}

func (r SubscribeResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** ConfirmSubscriptionResponse ***/
type ConfirmSubscriptionResponse struct {
	Xmlns    string                    `xml:"xmlns,attr"`
	Result   ConfirmSubscriptionResult `xml:"ConfirmSubscriptionResult"`
	Metadata app.ResponseMetadata      `xml:"ResponseMetadata"`
}

type ConfirmSubscriptionResult struct {
	SubscriptionArn string `xml:"SubscriptionArn"`
}

func (r ConfirmSubscriptionResponse) GetResult() interface{} {
	return r.Result
}

func (r ConfirmSubscriptionResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Subscription ***/
type UnsubscribeResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r UnsubscribeResponse) GetResult() interface{} {
	return nil
}

func (r UnsubscribeResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type DeleteMessageBatchResultEntry struct {
	Id string `xml:"Id"`
}

type DeleteMessageBatchResult struct {
	Successful []DeleteMessageBatchResultEntry `xml:"DeleteMessageBatchResultEntry"`
	Failed     []BatchResultErrorEntry         `xml:"BatchResultErrorEntry,omitempty"`
}

/*** Delete Message Batch Response */
type DeleteMessageBatchResponse struct {
	Xmlns    string                   `xml:"xmlns,attr,omitempty"`
	Result   DeleteMessageBatchResult `xml:"DeleteMessageBatchResult"`
	Metadata app.ResponseMetadata     `xml:"ResponseMetadata,omitempty"`
}

func (r DeleteMessageBatchResponse) GetResult() interface{} {
	return r.Result
}

func (r DeleteMessageBatchResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Publish ***/
type PublishResult struct {
	MessageId string `xml:"MessageId"`
}

type PublishResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Result   PublishResult        `xml:"PublishResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r PublishResponse) GetResult() interface{} {
	return r.Result
}

func (r PublishResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** List Topics ***/
type TopicArnResult struct {
	TopicArn  string `xml:"TopicArn"`
	NextToken string `xml:"NextToken"` // not implemented
}
type TopicNamestype struct {
	Member []TopicArnResult `xml:"member"`
}

type ListTopicsResult struct {
	Topics TopicNamestype `xml:"Topics"`
}

type ListTopicsResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Result   ListTopicsResult     `xml:"ListTopicsResult"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r ListTopicsResponse) GetResult() interface{} {
	return r.Result
}

func (r ListTopicsResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Topic ***/
type DeleteTopicResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r DeleteTopicResponse) GetResult() interface{} {
	return nil
}

func (r DeleteTopicResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/** List Subcriptions **/

type TopicMemberResult struct {
	TopicArn        string `xml:"TopicArn"`
	Protocol        string `xml:"Protocol"`
	SubscriptionArn string `xml:"SubscriptionArn"`
	Owner           string `xml:"Owner"`
	Endpoint        string `xml:"Endpoint"`
}

type TopicSubscriptions struct {
	Member []TopicMemberResult `xml:"member"`
}

type ListSubscriptionsResult struct {
	Subscriptions TopicSubscriptions `xml:"Subscriptions"`
}

type ListSubscriptionsResponse struct {
	Xmlns    string                  `xml:"xmlns,attr"`
	Result   ListSubscriptionsResult `xml:"ListSubscriptionsResult"`
	Metadata app.ResponseMetadata    `xml:"ResponseMetadata"`
}

func (r ListSubscriptionsResponse) GetResult() interface{} {
	return r.Result
}

func (r ListSubscriptionsResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Get Subscription Attributes ***/
type GetSubscriptionAttributesResult struct {
	Attributes GetSubscriptionAttributes `xml:"Attributes,omitempty"`
}

type GetSubscriptionAttributes struct {
	/* SubscriptionArn, FilterPolicy */
	Entries []SubscriptionAttributeEntry `xml:"entry,omitempty"`
}

type SubscriptionAttributeEntry struct {
	Key   string `xml:"key,omitempty"`
	Value string `xml:"value,omitempty"`
}

type GetSubscriptionAttributesResponse struct {
	Xmlns    string                          `xml:"xmlns,attr,omitempty"`
	Result   GetSubscriptionAttributesResult `xml:"GetSubscriptionAttributesResult"`
	Metadata app.ResponseMetadata            `xml:"ResponseMetadata,omitempty"`
}

func (r GetSubscriptionAttributesResponse) GetResult() interface{} {
	return r.Result
}

func (r GetSubscriptionAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Set Subscription Attributes ***/
type SetSubscriptionAttributesResponse struct {
	Xmlns    string               `xml:"xmlns,attr"`
	Metadata app.ResponseMetadata `xml:"ResponseMetadata"`
}

func (r SetSubscriptionAttributesResponse) GetResult() interface{} {
	return nil
}

func (r SetSubscriptionAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** List Subscriptions By Topic Response */
type ListSubscriptionsByTopicResult struct {
	NextToken     string             `xml:"NextToken"` // not implemented
	Subscriptions TopicSubscriptions `xml:"Subscriptions"`
}

type ListSubscriptionsByTopicResponse struct {
	Xmlns    string                         `xml:"xmlns,attr"`
	Result   ListSubscriptionsByTopicResult `xml:"ListSubscriptionsByTopicResult"`
	Metadata app.ResponseMetadata           `xml:"ResponseMetadata"`
}

func (r ListSubscriptionsByTopicResponse) GetResult() interface{} {
	return r.Result
}

func (r ListSubscriptionsByTopicResponse) GetRequestId() string {
	return r.Metadata.RequestId
}
