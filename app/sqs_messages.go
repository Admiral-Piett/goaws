package app

type DeleteMessageBatchResultEntry struct {
	Id string `xml:"Id"`
}

type SendMessageBatchResultEntry struct {
	Id                     string `xml:"Id"`
	MessageId              string `xml:"MessageId"`
	MD5OfMessageBody       string `xml:"MD5OfMessageBody,omitempty"`
	MD5OfMessageAttributes string `xml:"MD5OfMessageAttributes,omitempty"`
	SequenceNumber         string `xml:"SequenceNumber"`
}

type BatchResultErrorEntry struct {
	Code        string `xml:"Code"`
	Id          string `xml:"Id"`
	Message     string `xml:"Message,omitempty"`
	SenderFault bool   `xml:"SenderFault"`
}

type DeleteMessageBatchResult struct {
	Entry []DeleteMessageBatchResultEntry `xml:"DeleteMessageBatchResultEntry"`
	Error []BatchResultErrorEntry         `xml:"BatchResultErrorEntry,omitempty"`
}

/*** Delete Message Batch Response */
type DeleteMessageBatchResponse struct {
	Xmlns    string                   `xml:"xmlns,attr,omitempty"`
	Result   DeleteMessageBatchResult `xml:"DeleteMessageBatchResult"`
	Metadata ResponseMetadata         `xml:"ResponseMetadata,omitempty"`
}

type SendMessageBatchResult struct {
	Entry []SendMessageBatchResultEntry `xml:"SendMessageBatchResultEntry"`
	Error []BatchResultErrorEntry       `xml:"BatchResultErrorEntry,omitempty"`
}

/*** Delete Message Batch Response */
type SendMessageBatchResponse struct {
	Xmlns    string                 `xml:"xmlns,attr,omitempty"`
	Result   SendMessageBatchResult `xml:"SendMessageBatchResult"`
	Metadata ResponseMetadata       `xml:"ResponseMetadata,omitempty"`
}
