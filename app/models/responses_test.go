package models

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: For now, we're only going to test those methods that do something other than just return a field

func TestGetQueueAttributesResponse_GetResult(t *testing.T) {
	gqa := GetQueueAttributesResponse{
		Result: GetQueueAttributesResult{Attrs: []Attribute{
			{Name: "attribute-name1", Value: "attribute-value1"},
			{Name: "attribute-name2", Value: "attribute-value2"},
		}},
	}

	expectedAttributes := map[string]map[string]string{
		"Attributes": {
			"attribute-name1": "attribute-value1",
			"attribute-name2": "attribute-value2",
		},
	}
	result := gqa.GetResult()

	assert.Equal(t, expectedAttributes, result)
}

func Test_ResultMessage_MarshalXML_success_with_attributes(t *testing.T) {
	input := &ResultMessage{
		MessageId:              "message-id",
		ReceiptHandle:          "receipt-handle",
		MD5OfBody:              "body-md5",
		Body:                   "message-body",
		MD5OfMessageAttributes: "message-attrs-md5",
		MessageAttributes: map[string]MessageAttribute{
			"attr1": {
				DataType:    "String",
				StringValue: "string-value",
			},
			"attr2": {
				DataType:    "Binary",
				BinaryValue: "binary-value",
			},
			"attr3": {
				DataType:    "Number",
				StringValue: "number-value",
			},
		},
		Attributes: map[string]string{
			"ApproximateFirstReceiveTimestamp": "1",
			"SenderId":                         "2",
			"ApproximateReceiveCount":          "3",
			"SentTimestamp":                    "4",
		},
	}
	result, err := xml.Marshal(input)

	assert.Nil(t, err)

	resultString := string(result)

	// We have to assert piecemeal like this, the maps go into their lists unordered, which will randomly break this.
	entry := "<ResultMessage><MessageId>message-id</MessageId><ReceiptHandle>receipt-handle</ReceiptHandle><MD5OfBody>body-md5</MD5OfBody><Body>message-body</Body>"
	assert.Contains(t, resultString, entry)

	entry = "<Attribute><Name>ApproximateFirstReceiveTimestamp</Name><Value>1</Value></Attribute>"
	assert.Contains(t, resultString, entry)

	entry = "<Attribute><Name>SenderId</Name><Value>2</Value></Attribute>"
	assert.Contains(t, resultString, entry)

	entry = "<Attribute><Name>ApproximateReceiveCount</Name><Value>3</Value></Attribute>"
	assert.Contains(t, resultString, entry)

	entry = "<Attribute><Name>SentTimestamp</Name><Value>4</Value></Attribute>"
	assert.Contains(t, resultString, entry)

	entry = "<MessageAttribute><Name>attr1</Name><Value><DataType>String</DataType><StringValue>string-value</StringValue></Value></MessageAttribute>"
	assert.Contains(t, resultString, entry)

	entry = "<MessageAttribute><Name>attr2</Name><Value><BinaryValue>binary-value</BinaryValue><DataType>Binary</DataType></Value></MessageAttribute>"
	assert.Contains(t, resultString, entry)

	entry = "<MessageAttribute><Name>attr3</Name><Value><DataType>Number</DataType><StringValue>number-value</StringValue></Value></MessageAttribute>"
	assert.Contains(t, resultString, entry)

	entry = "</ResultMessage>"
	assert.Contains(t, resultString, entry)
}

func Test_ResultMessage_MarshalXML_success_no_attributes(t *testing.T) {
	input := &ResultMessage{
		MessageId:              "message-id",
		ReceiptHandle:          "receipt-handle",
		MD5OfBody:              "body-md5",
		Body:                   "message-body",
		MD5OfMessageAttributes: "message-attrs-md5",
	}
	expectedOutput := "<ResultMessage><MessageId>message-id</MessageId><ReceiptHandle>receipt-handle</ReceiptHandle><MD5OfBody>body-md5</MD5OfBody><Body>message-body</Body></ResultMessage>"

	result, err := xml.Marshal(input)

	assert.Nil(t, err)

	resultString := string(result)
	assert.Equal(t, resultString, expectedOutput)
}
