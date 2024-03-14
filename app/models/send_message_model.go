package models

import (
	"fmt"
	"net/url"

	"github.com/Admiral-Piett/goaws/app"
	log "github.com/sirupsen/logrus"
)

/*** Request */

func NewSendMessageRequest() *SendMessageRequest {
	return &SendMessageRequest{
		MessageAttributes:       make(map[string]MessageAttributes),
		MessageSystemAttributes: make(map[string]MessageAttributes),
	}
}

type SendMessageRequest struct {
	DelaySeconds            int                          `json:"Del1aySeconds" schema:"DelaySeconds"`
	MessageAttributes       map[string]MessageAttributes `json:"MessageAttributes" schema:"MessageAttributes"`
	MessageBody             string                       `json:"MessageBody" schema:"MessageBody"`
	MessageDeduplicationId  string                       `json:"MessageDeduplicationId" schema:"MessageDeduplicationId"`
	MessageGroupId          string                       `json:"MessageGroupId" schema:"MessageGroupId"`
	MessageSystemAttributes map[string]MessageAttributes `json:"MessageSystemAttributes" schema:"MessageSystemAttributes"`
	QueueUrl                string                       `json:"QueueUrl" schema:"QueueUrl"`
}
type MessageAttributes struct {
	BinaryListValues []string `json:"BinaryListValues"` // goaws does not supported yet
	BinaryValue      string   `json:"BinaryValue"`
	DataType         string   `json:"DataType"`
	StringListValues []string `json:"StringListValues"` // goaws does not supported yet
	StringValue      string   `json:"StringValue"`
}

func (r *SendMessageRequest) SetAttributesFromForm(values url.Values) {
	for i := 1; true; i++ {
		nameKey := fmt.Sprintf("MessageAttribute.%d.Name", i)
		name := values.Get(nameKey)
		if name == "" {
			break
		}

		dataTypeKey := fmt.Sprintf("MessageAttribute.%d.Value.DataType", i)
		dataType := values.Get(dataTypeKey)
		if dataType == "" {
			log.Warnf("DataType of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
			continue
		}

		stringValue := values.Get(fmt.Sprintf("MessageAttribute.%d.Value.StringValue", i))
		binaryValue := values.Get(fmt.Sprintf("MessageAttribute.%d.Value.BinaryValue", i))

		r.MessageAttributes[name] = MessageAttributes{
			DataType:    dataType,
			StringValue: stringValue,
			BinaryValue: binaryValue,
		}

		if _, ok := r.MessageAttributes[name]; !ok {
			log.Warnf("StringValue or BinaryValue of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
		}
	}
}

/*** Response */

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
