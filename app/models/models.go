package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Admiral-Piett/goaws/app"
	log "github.com/sirupsen/logrus"
)

var BASE_XMLNS = "http://queue.amazonaws.com/doc/2012-11-05/"
var BASE_RESPONSE_METADATA = app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}

var AVAILABLE_QUEUE_ATTRIBUTES = map[string]bool{
	"DelaySeconds":                          true,
	"MaximumMessageSize":                    true,
	"MessageRetentionPeriod":                true,
	"Policy":                                true,
	"ReceiveMessageWaitTimeSeconds":         true,
	"VisibilityTimeout":                     true,
	"RedrivePolicy":                         true,
	"RedriveAllowPolicy":                    true,
	"ApproximateNumberOfMessages":           true,
	"ApproximateNumberOfMessagesDelayed":    true,
	"ApproximateNumberOfMessagesNotVisible": true,
	"CreatedTimestamp":                      true,
	"LastModifiedTimestamp":                 true,
	"QueueArn":                              true,
}

type CreateQueueRequest struct {
	QueueName  string            `json:"QueueName" schema:"QueueName"`
	Attributes Attributes        `json:"Attributes" schema:"Attribute"`
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
type MessageAttributeValue struct {
	BinaryListValues []string `json:"BinaryListValues"` // currently unsupported by AWS
	BinaryValue      string   `json:"BinaryValue"`
	DataType         string   `json:"DataType"`
	StringListValues []string `json:"StringListValues"` // currently unsupported by AWS
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

		r.MessageAttributes[name] = MessageAttributeValue{
			DataType:    dataType,
			StringValue: stringValue,
			BinaryValue: binaryValue,
		}

		if _, ok := r.MessageAttributes[name]; !ok {
			log.Warnf("StringValue or BinaryValue of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
		}
	}
}

func NewSetQueueAttributesRequest() *SetQueueAttributesRequest {
	return &SetQueueAttributesRequest{}
}

type SetQueueAttributesRequest struct {
	QueueUrl   string     `json:"QueueUrl"`
	Attributes Attributes `json:"Attributes"`
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

// TODO - copy Attributes for SNS

// TODO - there are FIFO attributes and things too
// Attributes - SQS Attributes Available in create/set attributes requests.
// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_CreateQueue.html#SQS-CreateQueue-request-attributes
type Attributes struct {
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
		Attributes: Attributes{
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
