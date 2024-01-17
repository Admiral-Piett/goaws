package gosqs

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/stretchr/testify/assert"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app"
)

func TestApplyQueueAttributes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		deadLetterQueue := &app.Queue{Name: "failed-messages"}
		app.SyncQueues.Lock()
		app.SyncQueues.Queues["failed-messages"] = deadLetterQueue
		app.SyncQueues.Unlock()
		q := &app.Queue{VisibilityTimeout: 30}
		u := url.Values{}
		u.Add("Attribute.1.Name", "DelaySeconds")
		u.Add("Attribute.1.Value", "25")
		u.Add("Attribute.2.Name", "VisibilityTimeout")
		u.Add("Attribute.2.Value", "60")
		u.Add("Attribute.3.Name", "Policy")
		u.Add("Attribute.4.Name", "RedrivePolicy")
		u.Add("Attribute.4.Value", `{"maxReceiveCount": "4", "deadLetterTargetArn":"arn:aws:sqs::000000000000:failed-messages"}`)
		u.Add("Attribute.5.Name", "ReceiveMessageWaitTimeSeconds")
		u.Add("Attribute.5.Value", "20")
		if err := validateAndSetQueueAttributesFromForm(q, u); err != nil {
			t.Fatalf("expected nil, got %s", err)
		}
		expected := &app.Queue{
			VisibilityTimeout:             60,
			ReceiveMessageWaitTimeSeconds: 20,
			DelaySeconds:                  25,
			MaxReceiveCount:               4,
			DeadLetterQueue:               deadLetterQueue,
		}
		if ok := reflect.DeepEqual(q, expected); !ok {
			t.Fatalf("expected %+v, got %+v", expected, q)
		}
	})
	t.Run("missing_deadletter_arn", func(t *testing.T) {
		q := &app.Queue{VisibilityTimeout: 30}
		u := url.Values{}
		u.Add("Attribute.1.Name", "RedrivePolicy")
		u.Add("Attribute.1.Value", `{"maxReceiveCount": "4"}`)
		err := validateAndSetQueueAttributesFromForm(q, u)
		if err != ErrInvalidParameterValue {
			t.Fatalf("expected %s, got %s", ErrInvalidParameterValue, err)
		}
	})
	t.Run("invalid_redrive_policy", func(t *testing.T) {
		q := &app.Queue{VisibilityTimeout: 30}
		u := url.Values{}
		u.Add("Attribute.1.Name", "RedrivePolicy")
		u.Add("Attribute.1.Value", `{invalidinput}`)
		err := validateAndSetQueueAttributesFromForm(q, u)
		if err != ErrInvalidAttributeValue {
			t.Fatalf("expected %s, got %s", ErrInvalidAttributeValue, err)
		}
	})
}

func TestSetQueueAttributesV1_success_no_redrive_policy(t *testing.T) {
	var emptyQueue *app.Queue
	q := &app.Queue{}
	attrs := models.Attributes{
		DelaySeconds:                  1,
		MaximumMessageSize:            2,
		MessageRetentionPeriod:        3,
		ReceiveMessageWaitTimeSeconds: 4,
		VisibilityTimeout:             5,
	}
	err := setQueueAttributesV1(q, attrs)

	assert.Nil(t, err)
	assert.Equal(t, 1, q.DelaySeconds)
	assert.Equal(t, 2, q.MaximumMessageSize)
	assert.Equal(t, 3, q.MessageRetentionPeriod)
	assert.Equal(t, 4, q.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 5, q.VisibilityTimeout)
	assert.Equal(t, emptyQueue, q.DeadLetterQueue)
	assert.Equal(t, 0, q.MaxReceiveCount)
}

func TestSetQueueAttributesV1_success_no_request_attributes(t *testing.T) {
	var emptyQueue *app.Queue
	q := &app.Queue{}
	attrs := models.Attributes{}
	err := setQueueAttributesV1(q, attrs)

	assert.Nil(t, err)
	assert.Equal(t, 0, q.DelaySeconds)
	assert.Equal(t, 0, q.MaximumMessageSize)
	assert.Equal(t, 0, q.MessageRetentionPeriod)
	assert.Equal(t, 0, q.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 0, q.VisibilityTimeout)
	assert.Equal(t, emptyQueue, q.DeadLetterQueue)
	assert.Equal(t, 0, q.MaxReceiveCount)
}

func TestSetQueueAttributesV1_success_can_set_0_values_where_applicable(t *testing.T) {
	var emptyQueue *app.Queue
	q := &app.Queue{
		DelaySeconds:                  1,
		MaximumMessageSize:            2,
		MessageRetentionPeriod:        3,
		ReceiveMessageWaitTimeSeconds: 4,
		VisibilityTimeout:             5,
	}
	attrs := models.Attributes{}
	err := setQueueAttributesV1(q, attrs)

	assert.Nil(t, err)
	assert.Equal(t, 0, q.DelaySeconds)
	assert.Equal(t, 0, q.MaximumMessageSize)
	assert.Equal(t, 3, q.MessageRetentionPeriod)
	assert.Equal(t, 4, q.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 0, q.VisibilityTimeout)
	assert.Equal(t, emptyQueue, q.DeadLetterQueue)
	assert.Equal(t, 0, q.MaxReceiveCount)
}

func TestSetQueueAttributesV1_success_with_redrive_policy(t *testing.T) {
	defer func() {
		utils.ResetApp()
	}()

	existingQueueName := "existing-queue"
	existingQueue := &app.Queue{Name: existingQueueName}
	app.SyncQueues.Queues[existingQueueName] = existingQueue

	q := &app.Queue{}
	attrs := models.Attributes{
		DelaySeconds:                  1,
		MaximumMessageSize:            2,
		MessageRetentionPeriod:        3,
		ReceiveMessageWaitTimeSeconds: 4,
		VisibilityTimeout:             5,
		RedrivePolicy: models.RedrivePolicy{
			MaxReceiveCount:     10,
			DeadLetterTargetArn: fmt.Sprintf("arn:aws:sqs:region:account-id:%s", existingQueueName),
		},
	}
	err := setQueueAttributesV1(q, attrs)

	assert.Nil(t, err)
	assert.Equal(t, 1, q.DelaySeconds)
	assert.Equal(t, 2, q.MaximumMessageSize)
	assert.Equal(t, 3, q.MessageRetentionPeriod)
	assert.Equal(t, 4, q.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 5, q.VisibilityTimeout)
	assert.Equal(t, existingQueue, q.DeadLetterQueue)
	assert.Equal(t, 10, q.MaxReceiveCount)
}

func TestSetQueueAttributesV1_error_redrive_policy_targets_missing_queue(t *testing.T) {
	existingQueueName := "existing-queue"

	q := &app.Queue{}
	attrs := models.Attributes{
		DelaySeconds:                  1,
		MaximumMessageSize:            2,
		MessageRetentionPeriod:        3,
		ReceiveMessageWaitTimeSeconds: 4,
		VisibilityTimeout:             5,
		RedrivePolicy: models.RedrivePolicy{
			MaxReceiveCount:     10,
			DeadLetterTargetArn: fmt.Sprintf("arn:aws:sqs:region:account-id:%s", existingQueueName),
		},
	}
	err := setQueueAttributesV1(q, attrs)

	assert.Error(t, err)
}
