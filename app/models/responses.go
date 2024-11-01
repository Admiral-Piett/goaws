package models

import (
	"encoding/xml"
)

type ResponseMetadata struct {
	RequestId string `xml:"RequestId"`
}

// NOTE: Every response in here MUST implement the `AbstractResponseBody` interface in order to be used
//  in `encodeResponse`

/*** Error Responses ***/
type ErrorResult struct {
	Type    string `json:"Type,omitempty" xml:"Type,omitempty"`
	Code    string `json:"Code,omitempty" xml:"Code,omitempty"`
	Message string `json:"Message,omitempty" xml:"Message,omitempty"`
}

type ErrorResponse struct {
	Result    ErrorResult `json:"Error" xml:"Error"`
	RequestId string      `json:"RequestId" xml:"RequestId"`
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
	Xmlns    string               `json:"Xmlns" xml:"xmlns,attr"`
	Result   ReceiveMessageResult `json:"ReceiveMessageResult" xml:"ReceiveMessageResult"`
	Metadata ResponseMetadata     `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r ReceiveMessageResponse) GetResult() interface{} {
	return r.Result
}

func (r ReceiveMessageResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type ResultMessage struct {
	MessageId              string                      `json:"MessageId,omitempty" xml:"MessageId,omitempty"`
	ReceiptHandle          string                      `json:"ReceiptHandle,omitempty" xml:"ReceiptHandle,omitempty"`
	MD5OfBody              string                      `json:"MD5OfBody,omitempty" xml:"MD5OfBody,omitempty"`
	Body                   string                      `json:"Body,omitempty" xml:"Body,omitempty"`
	MD5OfMessageAttributes string                      `json:"MD5OfMessageAttributes,omitempty" xml:"MD5OfMessageAttributes,omitempty"`
	MessageAttributes      map[string]MessageAttribute `json:"MessageAttributes,omitempty" xml:"MessageAttribute,omitempty,attr"`
	Attributes             map[string]string           `json:"Attributes,omitempty" xml:"Attribute,omitempty,attr"`
}

// MarshalXML is a custom marshaler for the ResultMessage struct.  We need it because we need to convert the
// maps into something that can be shown as XML.  If we ever get rid of the XML response parsing this can go,
// and that would be glorious.
func (r *ResultMessage) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type Attributes struct {
		Name  string `xml:"Name,omitempty"`
		Value string `xml:"Value,omitempty"`
	}
	var attrs []Attributes
	for key, value := range r.Attributes {
		attribute := Attributes{
			Name:  key,
			Value: value,
		}
		attrs = append(attrs, attribute)
	}

	type MessageAttributes struct {
		Name  string           `xml:"Name,omitempty"`
		Value MessageAttribute `xml:"Value,omitempty"`
	}
	var messageAttrs []MessageAttributes
	for key, value := range r.MessageAttributes {
		attribute := MessageAttributes{
			Name:  key,
			Value: value,
		}
		messageAttrs = append(messageAttrs, attribute)
	}
	e.EncodeToken(start)

	// Encode the fields
	e.EncodeElement(r.MessageId, xml.StartElement{Name: xml.Name{Local: "MessageId"}})
	e.EncodeElement(r.ReceiptHandle, xml.StartElement{Name: xml.Name{Local: "ReceiptHandle"}})
	e.EncodeElement(r.MD5OfBody, xml.StartElement{Name: xml.Name{Local: "MD5OfBody"}})
	e.EncodeElement(r.Body, xml.StartElement{Name: xml.Name{Local: "Body"}})
	e.EncodeElement(attrs, xml.StartElement{Name: xml.Name{Local: "Attribute"}})
	e.EncodeElement(messageAttrs, xml.StartElement{Name: xml.Name{Local: "MessageAttribute"}})
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

type ChangeMessageVisibilityResult struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
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
	Xmlns    string            `json:"Xmlns" xml:"xmlns,attr"`
	Result   CreateQueueResult `json:"CreateQueueResult" xml:"CreateQueueResult"`
	Metadata ResponseMetadata  `json:"ResponseMetadata" xml:"ResponseMetadata"`
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
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Result   ListQueuesResult `json:"ListQueuesResult" xml:"ListQueuesResult"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r ListQueuesResponse) GetResult() interface{} {
	return r.Result
}

func (r ListQueuesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Get Queue QueueAttributes ***/
type Attribute struct {
	Name  string `json:"Name,omitempty" xml:"Name,omitempty"`
	Value string `json:"Value,omitempty" xml:"Value,omitempty"`
}

type GetQueueAttributesResult struct {
	/* VisibilityTimeout, DelaySeconds, ReceiveMessageWaitTimeSeconds, ApproximateNumberOfMessages
	   ApproximateNumberOfMessagesNotVisible, CreatedTimestamp, LastModifiedTimestamp, QueueArn */
	Attrs []Attribute `json:"Attributes,omitempty" xml:"Attribute,omitempty"`
}

type GetQueueAttributesResponse struct {
	Xmlns    string                   `json:"Xmlns" xml:"xmlns,attr"`
	Result   GetQueueAttributesResult `json:"GetQueueAttributesResult" xml:"GetQueueAttributesResult"`
	Metadata ResponseMetadata         `json:"ResponseMetadata" xml:"ResponseMetadata"`
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
	MD5OfMessageAttributes string `json:"MD5OfMessageAttributes,omitempty" xml:"MD5OfMessageAttributes,omitempty"`
	MD5OfMessageBody       string `json:"MD5OfMessageBody" xml:"MD5OfMessageBody"`
	MessageId              string `json:"MessageId" xml:"MessageId"`
	SequenceNumber         string `json:"SequenceNumber,omitempty" xml:"SequenceNumber,omitempty"`
}

type SendMessageResponse struct {
	Xmlns    string            `json:"Xmlns" xml:"xmlns,attr"`
	Result   SendMessageResult `json:"SendMessageResult" xml:"SendMessageResult"`
	Metadata ResponseMetadata  `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r SendMessageResponse) GetResult() interface{} {
	return r.Result
}

func (r SendMessageResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Message Response */
type DeleteMessageResponse struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r DeleteMessageResponse) GetResult() interface{} {
	return nil
}

func (r DeleteMessageResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Get Queue Url Response */
type GetQueueUrlResult struct {
	QueueUrl string `json:"QueueUrl,omitempty" xml:"QueueUrl,omitempty"`
}

type GetQueueUrlResponse struct {
	Xmlns    string            `json:"Xmlns" xml:"xmlns,attr"`
	Result   GetQueueUrlResult `json:"GetQueueUrlResult" xml:"GetQueueUrlResult"`
	Metadata ResponseMetadata  `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r GetQueueUrlResponse) GetResult() interface{} {
	return r.Result
}

func (r GetQueueUrlResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type SendMessageBatchResultEntry struct {
	Id                     string `json:"Id" xml:"Id"`
	MessageId              string `json:"MessageId" xml:"MessageId"`
	MD5OfMessageBody       string `json:"MD5OfMessageBody,omitempty" xml:"MD5OfMessageBody,omitempty"`
	MD5OfMessageAttributes string `json:"MD5OfMessageAttributes,omitempty" xml:"MD5OfMessageAttributes,omitempty"`
	SequenceNumber         string `json:"SequenceNumber" xml:"SequenceNumber"`
}

/*** Send Message Batch Response */
type SendMessageBatchResponse struct {
	Xmlns    string                 `json:"Xmlns" xml:"xmlns,attr"`
	Result   SendMessageBatchResult `json:"SendMessageBatchResult" xml:"SendMessageBatchResult"`
	Metadata ResponseMetadata       `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

type SendMessageBatchResult struct {
	Entry []SendMessageBatchResultEntry `json:"SendMessageBatchResultEntry" xml:"SendMessageBatchResultEntry"`
	Error []BatchResultErrorEntry       `json:"BatchResultErrorEntry,omitempty" xml:"BatchResultErrorEntry,omitempty"`
}

func (r SendMessageBatchResponse) GetResult() interface{} {
	return r.Result
}

func (r SendMessageBatchResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type BatchResultErrorEntry struct {
	Code        string `json:"Code" xml:"Code"`
	Id          string `json:"Id" xml:"Id"`
	Message     string `json:"Message,omitempty" xml:"Message,omitempty"`
	SenderFault bool   `json:"SenderFault" xml:"SenderFault"`
}

type SetQueueAttributesResponse struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r SetQueueAttributesResponse) GetResult() interface{} {
	return nil
}

func (r SetQueueAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Purge Queue Response */
type PurgeQueueResponse struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r PurgeQueueResponse) GetResult() interface{} {
	return nil
}

func (r PurgeQueueResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Queue Response */
type DeleteQueueResponse struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r DeleteQueueResponse) GetResult() interface{} {
	return nil
}

func (r DeleteQueueResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Create Topic Response */
type CreateTopicResult struct {
	TopicArn string `json:"TopicArn" xml:"TopicArn"`
}

type CreateTopicResponse struct {
	Xmlns    string            `json:"Xmlns" xml:"xmlns,attr"`
	Result   CreateTopicResult `json:"CreateTopicResult" xml:"CreateTopicResult"`
	Metadata ResponseMetadata  `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r CreateTopicResponse) GetResult() interface{} {
	return r.Result
}

func (r CreateTopicResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Create Subscription ***/
type SubscribeResult struct {
	SubscriptionArn string `json:"SubscriptionArn" xml:"SubscriptionArn"`
}

type SubscribeResponse struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Result   SubscribeResult  `json:"SubscribeResult" xml:"SubscribeResult"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r SubscribeResponse) GetResult() interface{} {
	return r.Result
}

func (r SubscribeResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** ConfirmSubscriptionResponse ***/
type ConfirmSubscriptionResponse struct {
	Xmlns    string                    `json:"Xmlns" xml:"xmlns,attr"`
	Result   ConfirmSubscriptionResult `json:"ConfirmSubscriptionResult" xml:"ConfirmSubscriptionResult"`
	Metadata ResponseMetadata          `json:"ResponseMetadata" xml:"ResponseMetadata"`
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
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
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
	Successful []DeleteMessageBatchResultEntry `json:"Successful" xml:"DeleteMessageBatchResultEntry"`
	Failed     []BatchResultErrorEntry         `json:"Failed,omitempty" xml:"BatchResultErrorEntry,omitempty"`
}

/*** Delete Message Batch Response */
type DeleteMessageBatchResponse struct {
	Xmlns    string                   `json:"Xmlns" xml:"xmlns,attr"`
	Result   DeleteMessageBatchResult `json:"DeleteMessageBatchResult" xml:"DeleteMessageBatchResult"`
	Metadata ResponseMetadata         `json:"ResponseMetadata" xml:"ResponseMetadata"`
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
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Result   PublishResult    `json:"PublishResult" xml:"PublishResult"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r PublishResponse) GetResult() interface{} {
	return r.Result
}

func (r PublishResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** List Topics ***/
type TopicArnResult struct {
	TopicArn  string `json:"TopicArn" xml:"TopicArn"`
	NextToken string `json:"NextToken" xml:"NextToken"` // not implemented
}
type TopicNamestype struct {
	Member []TopicArnResult `json:"Members" xml:"member"`
}

type ListTopicsResult struct {
	Topics TopicNamestype `json:"Topics" xml:"Topics"`
}

type ListTopicsResponse struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Result   ListTopicsResult `json:"ListTopicsResult" xml:"ListTopicsResult"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r ListTopicsResponse) GetResult() interface{} {
	return r.Result
}

func (r ListTopicsResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Topic ***/
type DeleteTopicResponse struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r DeleteTopicResponse) GetResult() interface{} {
	return nil
}

func (r DeleteTopicResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/** List Subcriptions **/

type TopicMemberResult struct {
	TopicArn        string `json:"TopicArn" xml:"TopicArn"`
	Protocol        string `json:"Protocol" xml:"Protocol"`
	SubscriptionArn string `json:"SubscriptionArn" xml:"SubscriptionArn"`
	Owner           string `json:"Owner" xml:"Owner"`
	Endpoint        string `json:"Endpoint" xml:"Endpoint"`
}

type TopicSubscriptions struct {
	Member []TopicMemberResult `json:"Members" xml:"member"`
}

type ListSubscriptionsResult struct {
	Subscriptions TopicSubscriptions `json:"Subscriptions" xml:"Subscriptions"`
}

type ListSubscriptionsResponse struct {
	Xmlns    string                  `json:"Xmlns" xml:"xmlns,attr"`
	Result   ListSubscriptionsResult `json:"ListSubscriptionsResult" xml:"ListSubscriptionsResult"`
	Metadata ResponseMetadata        `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r ListSubscriptionsResponse) GetResult() interface{} {
	return r.Result
}

func (r ListSubscriptionsResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Get Subscription Attributes ***/
type GetSubscriptionAttributesResult struct {
	Attributes GetSubscriptionAttributes `json:"Attributes,omitempty" xml:"Attributes,omitempty"`
}

type GetSubscriptionAttributes struct {
	/* SubscriptionArn, FilterPolicy */
	Entries []SubscriptionAttributeEntry `json:"Entries,omitempty" xml:"entry,omitempty"`
}

type SubscriptionAttributeEntry struct {
	Key   string `json:"Key,omitempty" xml:"key,omitempty"`
	Value string `json:"Value,omitempty" xml:"value,omitempty"`
}

type GetSubscriptionAttributesResponse struct {
	Xmlns    string                          `json:"Xmlns" xml:"xmlns,attr"`
	Result   GetSubscriptionAttributesResult `json:"GetSubscriptionAttributesResult" xml:"GetSubscriptionAttributesResult"`
	Metadata ResponseMetadata                `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r GetSubscriptionAttributesResponse) GetResult() interface{} {
	return r.Result
}

func (r GetSubscriptionAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Set Subscription Attributes ***/
type SetSubscriptionAttributesResponse struct {
	Xmlns    string           `json:"Xmlns" xml:"xmlns,attr"`
	Metadata ResponseMetadata `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r SetSubscriptionAttributesResponse) GetResult() interface{} {
	return nil
}

func (r SetSubscriptionAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** List Subscriptions By Topic Response */
type ListSubscriptionsByTopicResult struct {
	NextToken     string             `json:"NextToken" xml:"NextToken"` // not implemented
	Subscriptions TopicSubscriptions `json:"Subscriptions" xml:"Subscriptions"`
}

type ListSubscriptionsByTopicResponse struct {
	Xmlns    string                         `json:"Xmlns" xml:"xmlns,attr"`
	Result   ListSubscriptionsByTopicResult `json:"ListSubscriptionsByTopicResult" xml:"ListSubscriptionsByTopicResult"`
	Metadata ResponseMetadata               `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r ListSubscriptionsByTopicResponse) GetResult() interface{} {
	return r.Result
}

func (r ListSubscriptionsByTopicResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Publish Batch ***/
type PublishBatchFailed struct {
	ErrorEntries []BatchResultErrorEntry `json:"ErrorEntries" xml:"member"`
}

type PublishBatchResultEntry struct {
	Id        string `xml:"Id"`
	MessageId string `xml:"MessageId"`
}

type PublishBatchSuccessful struct {
	SuccessEntries []PublishBatchResultEntry `json:"SuccessEntries" xml:"member"`
}

type PublishBatchResult struct {
	Failed     PublishBatchFailed     `json:"Failed" xml:"Failed"`
	Successful PublishBatchSuccessful `json:"Successful" xml:"Successful"`
}

type PublishBatchResponse struct {
	Xmlns    string             `json:"Xmlns" xml:"xmlns,attr"`
	Result   PublishBatchResult `json:"PublishBatchResult" xml:"PublishBatchResult"`
	Metadata ResponseMetadata   `json:"ResponseMetadata" xml:"ResponseMetadata"`
}

func (r PublishBatchResponse) GetResult() interface{} {
	return r.Result
}

func (r PublishBatchResponse) GetRequestId() string {
	return r.Metadata.RequestId
}
