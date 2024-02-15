package gosqs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app"
)

var (
	ErrInvalidParameterValue = &app.SqsErrorType{
		HttpError: http.StatusBadRequest,
		Type:      "InvalidParameterValue",
		Code:      "AWS.SimpleQueueService.InvalidParameterValue",
		Message:   "An invalid or out-of-range value was supplied for the input parameter.",
	}
	ErrInvalidAttributeValue = &app.SqsErrorType{
		HttpError: http.StatusBadRequest,
		Type:      "InvalidAttributeValue",
		Code:      "AWS.SimpleQueueService.InvalidAttributeValue",
		Message:   "Invalid Value for the parameter RedrivePolicy.",
	}
)

// validateAndSetQueueAttributesFromForm applies the requested queue attributes to the given
// queue.
// TODO Currently it only supports VisibilityTimeout, MaximumMessageSize, DelaySeconds, RedrivePolicy and ReceiveMessageWaitTimeSeconds attributes.
func validateAndSetQueueAttributesFromForm(q *app.Queue, u url.Values) error {
	attr := utils.ExtractQueueAttributes(u)

	return validateAndSetQueueAttributes(q, attr)
}

func validateAndSetQueueAttributes(q *app.Queue, attr map[string]string) error {
	visibilityTimeout, _ := strconv.Atoi(attr["VisibilityTimeout"])
	if visibilityTimeout != 0 {
		q.VisibilityTimeout = visibilityTimeout
	}
	receiveWaitTime, _ := strconv.Atoi(attr["ReceiveMessageWaitTimeSeconds"])
	if receiveWaitTime != 0 {
		q.ReceiveMessageWaitTimeSeconds = receiveWaitTime
	}
	maximumMessageSize, _ := strconv.Atoi(attr["MaximumMessageSize"])
	if maximumMessageSize != 0 {
		q.MaximumMessageSize = maximumMessageSize
	}
	strRedrivePolicy := attr["RedrivePolicy"]
	if strRedrivePolicy != "" {
		// support both int and string maxReceiveCount (Amazon clients use string)
		redrivePolicy1 := struct {
			MaxReceiveCount     int    `json:"maxReceiveCount"`
			DeadLetterTargetArn string `json:"deadLetterTargetArn"`
		}{}
		redrivePolicy2 := struct {
			MaxReceiveCount     string `json:"maxReceiveCount"`
			DeadLetterTargetArn string `json:"deadLetterTargetArn"`
		}{}
		err1 := json.Unmarshal([]byte(strRedrivePolicy), &redrivePolicy1)
		err2 := json.Unmarshal([]byte(strRedrivePolicy), &redrivePolicy2)
		maxReceiveCount := redrivePolicy1.MaxReceiveCount
		deadLetterQueueArn := redrivePolicy1.DeadLetterTargetArn
		if err1 != nil && err2 != nil {
			return ErrInvalidAttributeValue
		} else if err1 != nil {
			maxReceiveCount, _ = strconv.Atoi(redrivePolicy2.MaxReceiveCount)
			deadLetterQueueArn = redrivePolicy2.DeadLetterTargetArn
		}

		if (deadLetterQueueArn != "" && maxReceiveCount == 0) ||
			(deadLetterQueueArn == "" && maxReceiveCount != 0) {
			return ErrInvalidParameterValue
		}
		dlt := strings.Split(deadLetterQueueArn, ":")
		deadLetterQueueName := dlt[len(dlt)-1]
		deadLetterQueue, ok := app.SyncQueues.Queues[deadLetterQueueName]
		if !ok {
			return ErrInvalidParameterValue
		}
		q.DeadLetterQueue = deadLetterQueue
		q.MaxReceiveCount = maxReceiveCount
	}
	delaySecs, _ := strconv.Atoi(attr["DelaySeconds"])
	if delaySecs != 0 {
		q.DelaySeconds = delaySecs
	}

	return nil
}

// TODO - Support:
//   - attr.MessageRetentionPeriod
//   - attr.Policy
//   - attr.RedriveAllowPolicy
func setQueueAttributesV1(q *app.Queue, attr models.Attributes) error {
	// FIXME - are there better places to put these bottom-limit validations?
	if attr.DelaySeconds >= 0 {
		q.DelaySeconds = attr.DelaySeconds
	}
	if attr.MaximumMessageSize >= 0 {
		q.MaximumMessageSize = attr.MaximumMessageSize
	}
	// TODO - bottom limit should be the AWS limits
	// The following 2 don't support zero values
	if attr.MessageRetentionPeriod > 0 {
		q.MessageRetentionPeriod = attr.MessageRetentionPeriod
	}
	if attr.ReceiveMessageWaitTimeSeconds > 0 {
		q.ReceiveMessageWaitTimeSeconds = attr.ReceiveMessageWaitTimeSeconds
	}
	if attr.VisibilityTimeout >= 0 {
		q.VisibilityTimeout = attr.VisibilityTimeout
	}
	if attr.RedrivePolicy != (models.RedrivePolicy{}) {
		arnArray := strings.Split(attr.RedrivePolicy.DeadLetterTargetArn, ":")
		queueName := arnArray[len(arnArray)-1]
		deadLetterQueue, ok := app.SyncQueues.Queues[queueName]
		if !ok {
			log.Error("Invalid RedrivePolicy Attribute")
			return fmt.Errorf(ErrInvalidAttributeValue.Type)
		}
		q.DeadLetterQueue = deadLetterQueue
		q.MaxReceiveCount = attr.RedrivePolicy.MaxReceiveCount
	}
	return nil
}
