package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Admiral-Piett/goaws/app"
	log "github.com/sirupsen/logrus"
)

func NewCreateQueueRequest() *CreateQueueRequest {
	return &CreateQueueRequest{
		Attributes: Attributes{
			DelaySeconds:                  0,
			MaximumMessageSize:            app.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize,
			MessageRetentionPeriod:        app.CurrentEnvironment.QueueAttributeDefaults.MessageRetentionPeriod,
			ReceiveMessageWaitTimeSeconds: app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds,
			VisibilityTimeout:             app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout,
		},
	}
}

type CreateQueueRequest struct {
	QueueName  string            `json:"QueueName" schema:"QueueName"`
	Attributes Attributes        `json:"Attributes" schema:"Attribute"`
	Tags       map[string]string `json:"Tags" schema:"Tags"`
}

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
			r.Attributes.DelaySeconds = tmp
		case "MaximumMessageSize":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.MaximumMessageSize = tmp
		case "MessageRetentionPeriod":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.MessageRetentionPeriod = tmp
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
			r.Attributes.ReceiveMessageWaitTimeSeconds = tmp
		case "VisibilityTimeout":
			tmp, err := strconv.Atoi(attrValue)
			if err != nil {
				log.Debugf("Failed to parse form attribute - %s: %s", attrName, attrValue)
				continue
			}
			r.Attributes.VisibilityTimeout = tmp
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
			tmp.MaxReceiveCount = int(receiveCount)
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
	DelaySeconds                  int                    `json:"DelaySeconds"`
	MaximumMessageSize            int                    `json:"MaximumMessageSize"`
	MessageRetentionPeriod        int                    `json:"MessageRetentionPeriod"` // NOTE: not implemented
	Policy                        map[string]interface{} `json:"Policy"`                 // NOTE: not implemented
	ReceiveMessageWaitTimeSeconds int                    `json:"ReceiveMessageWaitTimeSeconds"`
	VisibilityTimeout             int                    `json:"VisibilityTimeout"`
	// Dead Letter Queues Only
	RedrivePolicy      RedrivePolicy          `json:"RedrivePolicy"`
	RedriveAllowPolicy map[string]interface{} `json:"RedriveAllowPolicy"` // NOTE: not implemented
}

type RedrivePolicy struct {
	MaxReceiveCount     int    `json:"maxReceiveCount"`
	DeadLetterTargetArn string `json:"deadLetterTargetArn"`
}
