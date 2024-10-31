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
	Metadata ResponseMetadata     `xml:"ResponseMetadata"`
}

func (r ReceiveMessageResponse) GetResult() interface{} {
	return r.Result
}

func (r ReceiveMessageResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

type ResultMessage struct {
	MessageId              string                      `xml:"MessageId,omitempty"`
	ReceiptHandle          string                      `xml:"ReceiptHandle,omitempty"`
	MD5OfBody              string                      `xml:"MD5OfBody,omitempty"`
	Body                   string                      `xml:"Body,omitempty"`
	MD5OfMessageAttributes string                      `xml:"MD5OfMessageAttributes,omitempty"`
	MessageAttributes      map[string]MessageAttribute `xml:"MessageAttribute,omitempty,attr"`
	Attributes             map[string]string           `xml:"Attribute,omitempty,attr"`
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
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
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
	Xmlns    string            `xml:"xmlns,attr"`
	Result   CreateQueueResult `xml:"CreateQueueResult"`
	Metadata ResponseMetadata  `xml:"ResponseMetadata"`
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
	Xmlns    string           `xml:"xmlns,attr"`
	Result   ListQueuesResult `xml:"ListQueuesResult"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
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
	Metadata ResponseMetadata         `xml:"ResponseMetadata,omitempty"`
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
	Xmlns    string            `xml:"xmlns,attr"`
	Result   SendMessageResult `xml:"SendMessageResult"`
	Metadata ResponseMetadata  `xml:"ResponseMetadata"`
}

func (r SendMessageResponse) GetResult() interface{} {
	return r.Result
}

func (r SendMessageResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Message Response */
type DeleteMessageResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
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
	Xmlns    string            `xml:"xmlns,attr,omitempty"`
	Result   GetQueueUrlResult `xml:"GetQueueUrlResult"`
	Metadata ResponseMetadata  `xml:"ResponseMetadata,omitempty"`
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
	Metadata ResponseMetadata       `xml:"ResponseMetadata,omitempty"`
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
	Xmlns    string           `xml:"xmlns,attr,omitempty"`
	Metadata ResponseMetadata `xml:"ResponseMetadata,omitempty"`
}

func (r SetQueueAttributesResponse) GetResult() interface{} {
	return nil
}

func (r SetQueueAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Purge Queue Response */
type PurgeQueueResponse struct {
	Xmlns    string           `xml:"xmlns,attr,omitempty"`
	Metadata ResponseMetadata `xml:"ResponseMetadata,omitempty"`
}

func (r PurgeQueueResponse) GetResult() interface{} {
	return nil
}

func (r PurgeQueueResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Queue Response */
type DeleteQueueResponse struct {
	Xmlns    string           `xml:"xmlns,attr,omitempty"`
	Metadata ResponseMetadata `xml:"ResponseMetadata,omitempty"`
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
	Xmlns    string            `xml:"xmlns,attr"`
	Result   CreateTopicResult `xml:"CreateTopicResult"`
	Metadata ResponseMetadata  `xml:"ResponseMetadata"`
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
	Xmlns    string           `xml:"xmlns,attr"`
	Result   SubscribeResult  `xml:"SubscribeResult"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
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
	Metadata ResponseMetadata          `xml:"ResponseMetadata"`
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
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
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
	Metadata ResponseMetadata         `xml:"ResponseMetadata,omitempty"`
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
	Xmlns    string           `xml:"xmlns,attr"`
	Result   PublishResult    `xml:"PublishResult"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
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
	Xmlns    string           `xml:"xmlns,attr"`
	Result   ListTopicsResult `xml:"ListTopicsResult"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

func (r ListTopicsResponse) GetResult() interface{} {
	return r.Result
}

func (r ListTopicsResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Delete Topic ***/
type DeleteTopicResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
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
	Metadata ResponseMetadata        `xml:"ResponseMetadata"`
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
	Metadata ResponseMetadata                `xml:"ResponseMetadata,omitempty"`
}

func (r GetSubscriptionAttributesResponse) GetResult() interface{} {
	return r.Result
}

func (r GetSubscriptionAttributesResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Set Subscription Attributes ***/
type SetSubscriptionAttributesResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
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
	Metadata ResponseMetadata               `xml:"ResponseMetadata"`
}

func (r ListSubscriptionsByTopicResponse) GetResult() interface{} {
	return r.Result
}

func (r ListSubscriptionsByTopicResponse) GetRequestId() string {
	return r.Metadata.RequestId
}

/*** Publish Batch ***/
type PublishBatchFailed struct {
	ErrorEntries []BatchResultErrorEntry `xml:"member"`
}

type PublishBatchResultEntry struct {
	Id        string `xml:"Id"`
	MessageId string `xml:"MessageId"`
}

type PublishBatchSuccessful struct {
	SuccessEntries []PublishBatchResultEntry `xml:"member"`
}

type PublishBatchResult struct {
	Failed     PublishBatchFailed     `xml:"Failed"`
	Successful PublishBatchSuccessful `xml:"Successful"`
}

type PublishBatchResponse struct {
	Xmlns    string             `xml:"xmlns,attr"`
	Result   PublishBatchResult `xml:"PublishBatchResult"`
	Metadata ResponseMetadata   `xml:"ResponseMetadata"`
}

func (r PublishBatchResponse) GetResult() interface{} {
	return r.Result
}

func (r PublishBatchResponse) GetRequestId() string {
	return r.Metadata.RequestId
}
