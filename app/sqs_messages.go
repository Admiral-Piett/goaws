package app

/*** List Queues Response */
type ListQueuesResult struct {
	QueueUrl []string `xml:"QueueUrl" json:"QueueUrl"`
}

type ListQueuesResponse struct {
	Xmlns    string           `xml:"xmlns,attr" json:"-"`
	Result   ListQueuesResult `xml:"ListQueuesResult" json:"-"`
	Metadata ResponseMetadata `xml:"ResponseMetadata" json:"-"`
	ListQueuesResult
}

/*** Create Queue Response */
type CreateQueueResult struct {
	QueueUrl string `xml:"QueueUrl" json:"QueueUrl"`
}

type CreateQueueResponse struct {
	Xmlns    string            `xml:"xmlns,attr" json:"-"`
	Result   CreateQueueResult `xml:"CreateQueueResult" json:"-"`
	Metadata ResponseMetadata  `xml:"ResponseMetadata" json:"-"`
	CreateQueueResult
}

/*** Send Message Response */

type SendMessageResult struct {
	MD5OfMessageAttributes string `xml:"MD5OfMessageAttributes" json:"MD5OfMessageAttributes"`
	MD5OfMessageBody       string `xml:"MD5OfMessageBody" json:"MD5OfMessageBody"`
	MessageId              string `xml:"MessageId" json:"MessageId"`
	SequenceNumber         string `xml:"SequenceNumber" json:"SequenceNumber"`
}

type SendMessageResponse struct {
	Xmlns    string            `xml:"xmlns,attr" json:"-"`
	Result   SendMessageResult `xml:"SendMessageResult" json:"-"`
	Metadata ResponseMetadata  `xml:"ResponseMetadata" json:"-"`
	SendMessageResult
}

/*** Receive Message Response */

type ResultMessage struct {
	MessageId              string                                      `xml:"MessageId,omitempty" json:"MessageId,omitempty"`
	ReceiptHandle          string                                      `xml:"ReceiptHandle,omitempty" json:"ReceiptHandle,omitempty"`
	MD5OfBody              string                                      `xml:"MD5OfBody,omitempty" json:"MD5OfBody,omitempty"`
	Body                   []byte                                      `xml:"Body,omitempty" json:"-"`
	JSONBody               string                                      `json:"Body,omitempty"`
	MD5OfMessageAttributes string                                      `xml:"MD5OfMessageAttributes,omitempty" json:"MD5OfMessageAttributes,omitempty"`
	MessageAttributes      []*ResultMessageAttribute                   `xml:"MessageAttribute,omitempty" json:"-"`
	Attributes             []*ResultAttribute                          `xml:"Attribute,omitempty" json:"-"`
	JSONAttributes         map[string]string                           `json:"Attributes,omitempty"`
	JSONMessageAttributes  map[string]*ResultJSONMessageAttributeValue `json:"MessageAttributes,omitempty"`
}

type ResultMessageAttributeValue struct {
	DataType    string `xml:"DataType,omitempty" json:"DataType,omitempty"`
	StringValue string `xml:"StringValue,omitempty" json:"StringValue,omitempty"`
	BinaryValue string `xml:"BinaryValue,omitempty" json:"BinaryValue,omitempty"`
}

type ResultMessageAttribute struct {
	Name  string                       `xml:"Name,omitempty" json:"Name,omitempty"`
	Value *ResultMessageAttributeValue `xml:"Value,omitempty" json:"Value,omitempty"`
}

type ResultJSONMessageAttributeValue struct {
	DataType         string   `json:"DataType"`
	BinaryListValues []any    `json:"BinaryListValues"`
	BinaryValue      any      `json:"BinaryValue"`
	StringListValues []string `json:"StringListValues"`
	StringValue      string   `json:"StringValue"`
}

type ResultAttribute struct {
	Name  string `xml:"Name,omitempty" json:"Name,omitempty"`
	Value string `xml:"Value,omitempty" json:"Value,omitempty"`
}

type ReceiveMessageResult struct {
	Message []*ResultMessage `xml:"Message,omitempty" json:"Messages"`
}

type ReceiveMessageResponse struct {
	Xmlns    string               `xml:"xmlns,attr" json:"-"`
	Result   ReceiveMessageResult `xml:"ReceiveMessageResult" json:"-"`
	Metadata ResponseMetadata     `xml:"ResponseMetadata" json:"-"`
	ReceiveMessageResult
}

type ChangeMessageVisibilityResult struct {
	Xmlns    string           `xml:"xmlns,attr" json:"-"`
	Metadata ResponseMetadata `xml:"ResponseMetadata" json:"-"`
}

/*** Delete Message Response */
type DeleteMessageResponse struct {
	Xmlns    string           `xml:"xmlns,attr,omitempty" json:"-"`
	Metadata ResponseMetadata `xml:"ResponseMetadata,omitempty" json:"-"`
}

// Delete Queue
type DeleteQueueUrlResult struct {
	QueueUrl string `xml:"QueueUrl,omitempty" json:"QueueUrl,omitempty"`
}

type DeleteQueueResponse struct {
	Xmlns                string           `xml:"xmlns,attr,omitempty" json:"-"`
	Metadata             ResponseMetadata `xml:"ResponseMetadata,omitempty" json:"-"`
	DeleteQueueUrlResult `json:"QueueUrl,omitempty"`
}

type DeleteMessageBatchResultEntry struct {
	Id string `xml:"Id" json:""`
}

type SendMessageBatchResultEntry struct {
	Id                     string `xml:"Id" json:""`
	MessageId              string `xml:"MessageId" json:"MessageId"`
	MD5OfMessageBody       string `xml:"MD5OfMessageBody,omitempty" json:"MD5OfMessageBody,omitempty"`
	MD5OfMessageAttributes string `xml:"MD5OfMessageAttributes,omitempty" json:"MD5OfMessageAttributes,omitempty"`
	SequenceNumber         string `xml:"SequenceNumber" json:"SequenceNumber"`
}

type BatchResultErrorEntry struct {
	Code        string `xml:"Code" json:"Code"`
	Id          string `xml:"Id" json:"Id"`
	Message     string `xml:"Message,omitempty" json:"Message,omitempty"`
	SenderFault bool   `xml:"SenderFault" json:"SenderFault"`
}

type DeleteMessageBatchResult struct {
	Entry []DeleteMessageBatchResultEntry `xml:"DeleteMessageBatchResultEntry" json:"Successful"`
	Error []BatchResultErrorEntry         `xml:"BatchResultErrorEntry,omitempty" json:"Failed"`
}

/*** Delete Message Batch Response */
type DeleteMessageBatchResponse struct {
	Xmlns    string                   `xml:"xmlns,attr,omitempty" json:""`
	Result   DeleteMessageBatchResult `xml:"DeleteMessageBatchResult" json:"DeleteMessageBatchResult"`
	Metadata ResponseMetadata         `xml:"ResponseMetadata,omitempty"`
	DeleteMessageBatchResult
}

type SendMessageBatchResult struct {
	Entry []SendMessageBatchResultEntry `xml:"SendMessageBatchResultEntry" json:"Successful"`
	Error []BatchResultErrorEntry       `xml:"BatchResultErrorEntry,omitempty" json:"Failed"`
}

/*** Delete Message Batch Response */
type SendMessageBatchResponse struct {
	Xmlns    string                 `xml:"xmlns,attr,omitempty" json:"-"`
	Result   SendMessageBatchResult `xml:"SendMessageBatchResult" json:"-"`
	Metadata ResponseMetadata       `xml:"ResponseMetadata,omitempty" json:"-"`
	SendMessageBatchResult
}

/*** Purge Queue Response */
type PurgeQueueResponse struct {
	Xmlns    string           `xml:"xmlns,attr,omitempty" json:"-"`
	Metadata ResponseMetadata `xml:"ResponseMetadata,omitempty" json:"-"`
}

/*** Get Queue Url Response */
type GetQueueUrlResult struct {
	QueueUrl string `xml:"QueueUrl,omitempty" json:"QueueUrl,omitempty"`
}

type GetQueueUrlResponse struct {
	Xmlns    string            `xml:"xmlns,attr,omitempty" json:"-"`
	Result   GetQueueUrlResult `xml:"GetQueueUrlResult" json:"-"`
	Metadata ResponseMetadata  `xml:"ResponseMetadata,omitempty" json:"-"`
	GetQueueUrlResult
}

/*** Get Queue Attributes ***/
type Attribute struct {
	Name  string `xml:"Name,omitempty" json:"Name,omitempty"`
	Value string `xml:"Value,omitempty" json:"Value,omitempty"`
}

type GetQueueAttributesResult struct {
	/* VisibilityTimeout, DelaySeconds, ReceiveMessageWaitTimeSeconds, ApproximateNumberOfMessages
	   ApproximateNumberOfMessagesNotVisible, CreatedTimestamp, LastModifiedTimestamp, QueueArn */
	Attrs []Attribute `xml:"Attribute,omitempty" json:"Attribute,omitempty"`
}

type GetQueueAttributesResponse struct {
	Xmlns    string                   `xml:"xmlns,attr,omitempty" json:""`
	Result   GetQueueAttributesResult `xml:"GetQueueAttributesResult" json:"-"`
	Metadata ResponseMetadata         `xml:"ResponseMetadata,omitempty" json:"-"`
	GetQueueAttributesResult
}

type SetQueueAttributesResponse struct {
	Xmlns    string           `xml:"xmlns,attr,omitempty" json:"-"`
	Metadata ResponseMetadata `xml:"ResponseMetadata,omitempty"`
}
