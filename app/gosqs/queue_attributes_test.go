package gosqs

import (
	"fmt"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/stretchr/testify/assert"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app"
)

func TestSetQueueAttributesV1_success_no_redrive_policy(t *testing.T) {
	var emptyQueue *app.Queue
	q := &app.Queue{}
	attrs := models.QueueAttributes{
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
	attrs := models.QueueAttributes{}
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
	attrs := models.QueueAttributes{}
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
		test.ResetApp()
	}()

	existingQueueName := "existing-queue"
	existingQueue := &app.Queue{Name: existingQueueName}
	app.SyncQueues.Queues[existingQueueName] = existingQueue

	q := &app.Queue{}
	attrs := models.QueueAttributes{
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
	attrs := models.QueueAttributes{
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
