package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Admiral-Piett/goaws/app"
	log "github.com/sirupsen/logrus"
)

type CreateQueueRequest struct {
	QueueName  string            `json:"QueueName" schema:"QueueName"`
	Attributes QueueAttributes   `json:"Attributes" schema:"Attribute"`
	Tags       map[string]string `json:"Tags" schema:"Tags"`
	Version    string            `json:"Version" schema:"Version"`
}

// TODO - is there an easier way to do this?  Similar to the StringToInt type?
func (r *CreateQueueRequest) SetAttributesFromForm(values url.Values) {
	for i := 1; true; i++ {
		nameKey := fmt.Sprintf("Attribute.%d.Name", i)
		attrName := values.Get(nameKey)
		if attrName == "" {
			break
		}

		valueKey := fmt.Sprintf("Attribute.%d.Value", i)
		attrValue := values.Get(valueKey)
		if attrValue == "" {
			continue
		}
		switch attrName {
		case "DelaySeconds":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.DelaySeconds = StringToInt(tmp)
		case "MaximumMessageSize":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.MaximumMessageSize = StringToInt(tmp)
		case "MessageRetentionPeriod":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.MessageRetentionPeriod = StringToInt(tmp)
		case "Policy":
			var tmp map[string]interface{}
			err := json.Unmarshal([]byte(attrValue), &tmp)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.Policy = tmp
		case "ReceiveMessageWaitTimeSeconds":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.ReceiveMessageWaitTimeSeconds = StringToInt(tmp)
		case "VisibilityTimeout":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.VisibilityTimeout = StringToInt(tmp)
		case "RedrivePolicy":
			tmp := RedrivePolicy{}
			var decodedPolicy struct {
				MaxReceiveCount     interface{} `json:"maxReceiveCount"`
				DeadLetterTargetArn string      `json:"deadLetterTargetArn"`
			}
			err := json.Unmarshal([]byte(attrValue), &decodedPolicy)
			if err != nil || decodedPolicy.DeadLetterTargetArn == "" {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			// Support both int and string types (historic processing), set a default of 10 if not provided.
			// Go will default into float64 for interface{} types when parsing numbers
			receiveCount, ok := decodedPolicy.MaxReceiveCount.(float64)
			if !ok {
				receiveCount = 10
				t, ok := decodedPolicy.MaxReceiveCount.(string)
				if ok {
					r, err := strconv.ParseFloat(t, 64)
					if err == nil {
						receiveCount = r
					} else {
						log.Debugf("Failed to parse form attribute (maxReceiveCount) - %s: %s", attrName, attrValue)
					}
				} else {
					log.Debugf("Failed to parse form attribute (maxReceiveCount) - %s: %s", attrName, attrValue)
				}
			}
			tmp.MaxReceiveCount = StringToInt(receiveCount)
			tmp.DeadLetterTargetArn = decodedPolicy.DeadLetterTargetArn
			r.Attributes.RedrivePolicy = tmp
		case "RedriveAllowPolicy":
			var tmp map[string]interface{}
			err := json.Unmarshal([]byte(attrValue), &tmp)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.RedriveAllowPolicy = tmp
		}
	}
	return
}

func NewListQueuesRequest() *ListQueueRequest {
	return &ListQueueRequest{}
}

type ListQueueRequest struct {
	MaxResults      int    `json:"MaxResults" schema:"MaxResults"`
	NextToken       string `json:"NextToken" schema:"NextToken"`
	QueueNamePrefix string `json:"QueueNamePrefix" schema:"QueueNamePrefix"`
}

func (r *ListQueueRequest) SetAttributesFromForm(values url.Values) {
	maxResults, err := strconv.Atoi(values.Get("MaxResults"))
	if err == nil {
		r.MaxResults = maxResults
	}
	r.NextToken = values.Get("NextToken")
	r.QueueNamePrefix = values.Get("QueueNamePrefix")
}

func NewGetQueueAttributesRequest() *GetQueueAttributesRequest {
	return &GetQueueAttributesRequest{}
}

type GetQueueAttributesRequest struct {
	QueueUrl       string   `json:"QueueUrl"`
	AttributeNames []string `json:"AttributeNames"`
}

func (r *GetQueueAttributesRequest) SetAttributesFromForm(values url.Values) {
	r.QueueUrl = values.Get("QueueUrl")
	// TODO - test me
	for i := 1; true; i++ {
		attrKey := fmt.Sprintf("AttributeName.%d", i)
		attrValue := values.Get(attrKey)
		if attrValue == "" {
			break
		}
		r.AttributeNames = append(r.AttributeNames, attrValue)
	}
}

/*** Send Message Request */
func NewSendMessageRequest() *SendMessageRequest {
	return &SendMessageRequest{
		MessageAttributes:       make(map[string]MessageAttributeValue),
		MessageSystemAttributes: make(map[string]MessageAttributeValue),
	}
}

type SendMessageRequest struct {
	DelaySeconds int `json:"DelaySeconds" schema:"DelaySeconds"`
	// MessageAttributes is custom attributes that users can add on the message as they like.
	// Please see: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SendMessage.html#SQS-SendMessage-request-MessageAttributes
	MessageAttributes      map[string]MessageAttributeValue `json:"MessageAttributes" schema:"MessageAttributes"`
	MessageBody            string                           `json:"MessageBody" schema:"MessageBody"`
	MessageDeduplicationId string                           `json:"MessageDeduplicationId" schema:"MessageDeduplicationId"`
	MessageGroupId         string                           `json:"MessageGroupId" schema:"MessageGroupId"`
	// MessageSystemAttributes is custom attributes for AWS services.
	// Please see: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SendMessage.html#SQS-SendMessage-request-MessageSystemAttributes
	// On AWS, the only supported attribute is "AWSTraceHeader" that is for AWS X-Ray.
	// Goaws does not contains X-Ray emulation, so currently MessageSystemAttributes is unsupported.
	// TODO: Replace with a struct with known attributes "AWSTraceHeader".
	MessageSystemAttributes map[string]MessageAttributeValue `json:"MessageSystemAttributes" schema:"MessageSystemAttributes"`
	QueueUrl                string                           `json:"QueueUrl" schema:"QueueUrl"`
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

		r.MessageAttributes[name] = MessageAttributeValue{
			DataType:    dataType,
			StringValue: stringValue,
			BinaryValue: binaryValue,
		}
	}
}

func NewSendMessageBatchRequest() *SendMessageBatchRequest {
	return &SendMessageBatchRequest{}
}

type SendMessageBatchRequest struct {
	Entries  []SendMessageBatchRequestEntry
	QueueUrl string
}

func (r *SendMessageBatchRequest) SetAttributesFromForm(values url.Values) {
	for key := range values {

		keySegments := strings.Split(key, ".")
		//If index value size is 3 or less, there is no attribute value
		if len(keySegments) <= 3 {
			continue
		}

		// Both patterns below are supported here.
		// strconv.Atoi(keySegments[1] - targets the index value in pattern: `Entries.1.MessageBody`
		// strconv.Atoi(keySegments[3] - targets the index value in pattern: `Entries.1.MessageAttributes.1.Name`
		entryIndex, err1 := strconv.Atoi(keySegments[1])
		attributeIndex, err2 := strconv.Atoi(keySegments[3])

		// If the entry index and attribute index cannot be obtained, the attribute will not be set, so skip
		if err1 != nil || err2 != nil {
			continue
		}

		nameKey := fmt.Sprintf("Entries.%d.MessageAttributes.%d.Name", entryIndex, attributeIndex)
		if key != nameKey {
			continue
		}
		name := values.Get(nameKey)
		dataTypeKey := fmt.Sprintf("Entries.%d.MessageAttributes.%d.Value.DataType", entryIndex, attributeIndex)
		dataType := values.Get(dataTypeKey)
		if dataType == "" {
			log.Warnf("DataType of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
			continue
		}

		stringValue := values.Get(fmt.Sprintf("Entries.%d.MessageAttributes.%d.Value.StringValue", entryIndex, attributeIndex))
		binaryValue := values.Get(fmt.Sprintf("Entries.%d.MessageAttributes.%d.Value.BinaryValue", entryIndex, attributeIndex))

		if r.Entries[entryIndex].MessageAttributes == nil {
			r.Entries[entryIndex].MessageAttributes = make(map[string]MessageAttributeValue)
		}

		r.Entries[entryIndex].MessageAttributes[name] = MessageAttributeValue{
			DataType:    dataType,
			StringValue: stringValue,
			BinaryValue: binaryValue,
		}

		if _, ok := r.Entries[entryIndex].MessageAttributes[name]; !ok {
			log.Warnf("StringValue or BinaryValue of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
		}
	}
}

type SendMessageBatchRequestEntry struct {
	Id                      string                           `json:"Id" schema:"Id"`
	MessageBody             string                           `json:"MessageBody" schema:"MessageBody"`
	DelaySeconds            int                              `json:"DelaySeconds" schema:"DelaySeconds"` // NOTE: not implemented
	MessageAttributes       map[string]MessageAttributeValue `json:"MessageAttributes" schema:"MessageAttributes"`
	MessageDeduplicationId  string                           `json:"MessageDeduplicationId" schema:"MessageDeduplicationId"`
	MessageGroupId          string                           `json:"MessageGroupId" schema:"MessageGroupId"`
	MessageSystemAttributes map[string]MessageAttributeValue `json:"MessageSystemAttributes" schema:"MessageSystemAttributes"` // NOTE: not implemented
}

// Get Queue Url Request
func NewGetQueueUrlRequest() *GetQueueUrlRequest {
	return &GetQueueUrlRequest{}
}

type GetQueueUrlRequest struct {
	QueueName              string `json:"QueueName"`
	QueueOwnerAWSAccountId string `json:"QueueOwnerAWSAccountId"` // NOTE: not implemented
}

func (r *GetQueueUrlRequest) SetAttributesFromForm(values url.Values) {}

func NewSetQueueAttributesRequest() *SetQueueAttributesRequest {
	return &SetQueueAttributesRequest{}
}

type SetQueueAttributesRequest struct {
	QueueUrl   string          `json:"QueueUrl"`
	Attributes QueueAttributes `json:"Attributes"`
}

func (r *SetQueueAttributesRequest) SetAttributesFromForm(values url.Values) {
	r.QueueUrl = values.Get("QueueUrl")
	// TODO - could we share with CreateQueueRequest?
	for i := 1; true; i++ {
		nameKey := fmt.Sprintf("Attribute.%d.Name", i)
		attrName := values.Get(nameKey)
		if attrName == "" {
			break
		}

		valueKey := fmt.Sprintf("Attribute.%d.Value", i)
		attrValue := values.Get(valueKey)
		if attrValue == "" {
			continue
		}
		switch attrName {
		case "DelaySeconds":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.DelaySeconds = StringToInt(tmp)
		case "MaximumMessageSize":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.MaximumMessageSize = StringToInt(tmp)
		case "MessageRetentionPeriod":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.MessageRetentionPeriod = StringToInt(tmp)
		case "Policy":
			var tmp map[string]interface{}
			err := json.Unmarshal([]byte(attrValue), &tmp)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.Policy = tmp
		case "ReceiveMessageWaitTimeSeconds":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.ReceiveMessageWaitTimeSeconds = StringToInt(tmp)
		case "VisibilityTimeout":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.VisibilityTimeout = StringToInt(tmp)
		case "RedrivePolicy":
			tmp := RedrivePolicy{}
			var decodedPolicy struct {
				MaxReceiveCount     interface{} `json:"maxReceiveCount"`
				DeadLetterTargetArn string      `json:"deadLetterTargetArn"`
			}
			err := json.Unmarshal([]byte(attrValue), &decodedPolicy)
			if err != nil || decodedPolicy.DeadLetterTargetArn == "" {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			// Support both int and string types (historic processing), set a default of 10 if not provided.
			// Go will default into float64 for interface{} types when parsing numbers
			receiveCount, ok := decodedPolicy.MaxReceiveCount.(float64)
			if !ok {
				receiveCount = 10
				t, ok := decodedPolicy.MaxReceiveCount.(string)
				if ok {
					r, err := strconv.ParseFloat(t, 64)
					if err == nil {
						receiveCount = r
					} else {
						log.Debugf("Failed to parse form attribute (maxReceiveCount) - %s: %s", attrName, attrValue)
					}
				} else {
					log.Debugf("Failed to parse form attribute (maxReceiveCount) - %s: %s", attrName, attrValue)
				}
			}
			tmp.MaxReceiveCount = StringToInt(receiveCount)
			tmp.DeadLetterTargetArn = decodedPolicy.DeadLetterTargetArn
			r.Attributes.RedrivePolicy = tmp
		case "RedriveAllowPolicy":
			var tmp map[string]interface{}
			err := json.Unmarshal([]byte(attrValue), &tmp)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.RedriveAllowPolicy = tmp
		}
	}
	return
}

// TODO - there are FIFO attributes and things too
// QueueAttributes - SQS QueueAttributes Available in create/set attributes requests.
// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_CreateQueue.html#SQS-CreateQueue-request-attributes
type QueueAttributes struct {
	DelaySeconds                  StringToInt            `json:"DelaySeconds"`
	MaximumMessageSize            StringToInt            `json:"MaximumMessageSize"`
	MessageRetentionPeriod        StringToInt            `json:"MessageRetentionPeriod"` // NOTE: not implemented
	Policy                        map[string]interface{} `json:"Policy"`                 // NOTE: not implemented
	ReceiveMessageWaitTimeSeconds StringToInt            `json:"ReceiveMessageWaitTimeSeconds"`
	VisibilityTimeout             StringToInt            `json:"VisibilityTimeout"`
	// Dead Letter Queues Only
	RedrivePolicy      RedrivePolicy          `json:"RedrivePolicy"`
	RedriveAllowPolicy map[string]interface{} `json:"RedriveAllowPolicy"` // NOTE: not implemented
}

type RedrivePolicy struct {
	MaxReceiveCount     StringToInt `json:"maxReceiveCount"`
	DeadLetterTargetArn string      `json:"deadLetterTargetArn"`
}

// UnmarshalJSON this will convert a JSON string of a Redrive Policy sub-doc (escaped characters and all) or
// a regular json document into the appropriate resulting struct.
func (r *RedrivePolicy) UnmarshalJSON(data []byte) error {
	type basicRequest RedrivePolicy

	err := json.Unmarshal(data, (*basicRequest)(r))
	if err == nil {
		return nil
	}

	tmp, _ := strconv.Unquote(string(data))
	err = json.Unmarshal([]byte(tmp), (*basicRequest)(r))
	if err != nil {
		return err
	}
	return nil
}

func NewReceiveMessageRequest() *ReceiveMessageRequest {
	return &ReceiveMessageRequest{}
}

type ReceiveMessageRequest struct {
	QueueUrl                    string   `json:"QueueUrl" schema:"QueueUrl"`
	AttributeNames              []string `json:"AttributeNames" schema:"AttributeNames"`
	MessageSystemAttributeNames []string `json:"MessageSystemAttributeNames" schema:"MessageSystemAttributeNames"`
	MessageAttributeNames       []string `json:"MessageAttributeNames" schema:"MessageAttributeNames"`
	MaxNumberOfMessages         int      `json:"MaxNumberOfMessages" schema:"MaxNumberOfMessages"`
	VisibilityTimeout           int      `json:"VisibilityTimeout" schema:"VisibilityTimeout"`
	WaitTimeSeconds             int      `json:"WaitTimeSeconds" schema:"WaitTimeSeconds"`
	ReceiveRequestAttemptId     string   `json:"ReceiveRequestAttemptId" schema:"ReceiveRequestAttemptId"`
}

func (r *ReceiveMessageRequest) SetAttributesFromForm(values url.Values) {}

func NewCreateQueueRequest() *CreateQueueRequest {
	return &CreateQueueRequest{
		Attributes: QueueAttributes{
			DelaySeconds:                  0,
			MaximumMessageSize:            StringToInt(app.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize),
			MessageRetentionPeriod:        StringToInt(app.CurrentEnvironment.QueueAttributeDefaults.MessageRetentionPeriod),
			ReceiveMessageWaitTimeSeconds: StringToInt(app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds),
			VisibilityTimeout:             StringToInt(app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout),
		},
	}
}

func NewChangeMessageVisibilityRequest() *ChangeMessageVisibilityRequest {
	return &ChangeMessageVisibilityRequest{}
}

type ChangeMessageVisibilityRequest struct {
	QueueUrl          string `json:"QueueUrl" schema:"QueueUrl"`
	ReceiptHandle     string `json:"ReceiptHandle" schema:"ReceiptHandle"`
	VisibilityTimeout int    `json:"VisibilityTimeout" schema:"VisibilityTimeout"`
}

func (r *ChangeMessageVisibilityRequest) SetAttributesFromForm(values url.Values) {}

func NewDeleteMessageRequest() *DeleteMessageRequest {
	return &DeleteMessageRequest{}
}

type DeleteMessageRequest struct {
	QueueUrl      string `json:"QueueUrl" schema:"QueueUrl"`
	ReceiptHandle string `json:"ReceiptHandle" schema:"ReceiptHandle"`
}

func (r *DeleteMessageRequest) SetAttributesFromForm(values url.Values) {}

func NewPurgeQueueRequest() *PurgeQueueRequest {
	return &PurgeQueueRequest{}
}

type PurgeQueueRequest struct {
	QueueUrl string `json:"QueueUrl" schema:"QueueUrl"`
}

func (r *PurgeQueueRequest) SetAttributesFromForm(values url.Values) {}

func NewDeleteQueueRequest() *DeleteQueueRequest {
	return &DeleteQueueRequest{}
}

type DeleteQueueRequest struct {
	QueueUrl string `json:"QueueUrl" schema:"QueueUrl"`
}

func (r *DeleteQueueRequest) SetAttributesFromForm(values url.Values) {}

type DeleteMessageBatchRequestEntry struct {
	Id            string `json:"Id" schema:"Id"`
	ReceiptHandle string `json:"ReceiptHandle" schema:"ReceiptHandle"`
}

type DeleteMessageBatchRequest struct {
	Entries  []DeleteMessageBatchRequestEntry `json:"Entries" schema:"Entries"`
	QueueUrl string                           `json:"QueueUrl" schema:"QueueUrl"`
}

func NewDeleteMessageBatchRequest() *DeleteMessageBatchRequest {
	return &DeleteMessageBatchRequest{}
}

func (r *DeleteMessageBatchRequest) SetAttributesFromForm(values url.Values) {}
