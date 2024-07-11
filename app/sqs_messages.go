package app

type DeleteMessageBatchResultEntry struct {
	Id string `xml:"Id"`
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
