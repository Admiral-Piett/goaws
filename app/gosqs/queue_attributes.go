package gosqs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/archa347/goaws/app"
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

// validateQueueAttributes applies the requested queue attributes to the given
// queue.
// TODO Currently it only supports VisibilityTimeout and RedrivePolicy attributes.
func validateQueueAttributes(q *app.Queue, u url.Values) error {
	attr := extractQueueAttributes(u)
	visibilityTimeout, _ := strconv.Atoi(attr["VisibilityTimeout"])
	if visibilityTimeout != 0 {
		q.TimeoutSecs = visibilityTimeout
	}
	redrivePolicy := attr["RedrivePolicy"]
	if redrivePolicy != "" {
		str := struct {
			MaxReceiveCount     int    `json:"maxReceiveCount"`
			DeadLetterTargetArn string `json:"deadLetterTargetArn"`
		}{}
		err := json.Unmarshal([]byte(redrivePolicy), &str)
		if err != nil {
			return ErrInvalidAttributeValue
		}

		if (str.DeadLetterTargetArn != "" && str.MaxReceiveCount == 0) ||
			(str.DeadLetterTargetArn == "" && str.MaxReceiveCount != 0) {
			return ErrInvalidParameterValue
		}
		dlt := strings.Split(str.DeadLetterTargetArn, ":")
		deadLetterQueueName := dlt[len(dlt)-1]
		deadLetterQueue, ok := app.SyncQueues.Queues[deadLetterQueueName]
		if !ok {
			return ErrInvalidParameterValue
		}
		q.DeadLetterQueue = deadLetterQueue
		q.MaxReceiveCount = str.MaxReceiveCount
	}

	return nil
}

func extractQueueAttributes(u url.Values) map[string]string {
	attr := map[string]string{}
	for i := 1; true; i++ {
		nameKey := fmt.Sprintf("Attribute.%d.Name", i)
		attrName := u.Get(nameKey)
		if attrName == "" {
			break
		}

		valueKey := fmt.Sprintf("Attribute.%d.Value", i)
		attrValue := u.Get(valueKey)
		if attrValue != "" {
			attr[attrName] = attrValue
		}
	}
	return attr
}
