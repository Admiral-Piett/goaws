package models

import "github.com/Admiral-Piett/goaws/app"

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
