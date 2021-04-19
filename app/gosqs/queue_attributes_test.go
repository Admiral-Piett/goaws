package gosqs

import (
	"net/url"
	"testing"

	"github.com/p4tin/goaws/app"
	"github.com/stretchr/testify/assert"
)

func TestApplyQueueAttributes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert := assert.New(t)
		deadLetterQueue := &app.Queue{Name: "failed-messages"}
		app.SyncQueues.Lock()
		app.SyncQueues.Queues["failed-messages"] = deadLetterQueue
		app.SyncQueues.Unlock()
		q := &app.Queue{TimeoutSecs: 30}
		u := url.Values{}
		u.Add("Attribute.1.Name", "DelaySeconds")
		u.Add("Attribute.1.Value", "20")
		u.Add("Attribute.2.Name", "VisibilityTimeout")
		u.Add("Attribute.2.Value", "60")
		u.Add("Attribute.3.Name", "Policy")
		u.Add("Attribute.4.Name", "RedrivePolicy")
		u.Add("Attribute.4.Value", `{"maxReceiveCount": "4", "deadLetterTargetArn":"arn:aws:sqs::000000000000:failed-messages"}`)
		u.Add("Attribute.5.Name", "ReceiveMessageWaitTimeSeconds")
		u.Add("Attribute.5.Value", "20")
		assert.NoError(validateAndSetQueueAttributes(q, u))
		assert.Equal(60, q.TimeoutSecs)
		assert.Equal(20, q.ReceiveWaitTimeSecs)
		assert.Equal(4, q.MaxReceiveCount)
		assert.Equal(deadLetterQueue, q.DeadLetterQueue)
	})
	t.Run("missing_deadletter_arn", func(t *testing.T) {
		assert := assert.New(t)
		q := &app.Queue{TimeoutSecs: 30}
		u := url.Values{}
		u.Add("Attribute.1.Name", "RedrivePolicy")
		u.Add("Attribute.1.Value", `{"maxReceiveCount": "4"}`)
		assert.Equal(ErrInvalidParameterValue, validateAndSetQueueAttributes(q, u))
	})
	t.Run("invalid_redrive_policy", func(t *testing.T) {
		assert := assert.New(t)
		q := &app.Queue{TimeoutSecs: 30}
		u := url.Values{}
		u.Add("Attribute.1.Name", "RedrivePolicy")
		u.Add("Attribute.1.Value", `{invalidinput}`)
		assert.Equal(ErrInvalidAttributeValue, validateAndSetQueueAttributes(q, u))
	})
}

func TestExtractQueueAttributes(t *testing.T) {
	assert := assert.New(t)
	u := url.Values{}
	u.Add("Attribute.1.Name", "DelaySeconds")
	u.Add("Attribute.1.Value", "20")
	u.Add("Attribute.2.Name", "VisibilityTimeout")
	u.Add("Attribute.2.Value", "30")
	u.Add("Attribute.3.Name", "Policy")
	attr := extractQueueAttributes(u)
	expected := map[string]string{
		"DelaySeconds":      "20",
		"VisibilityTimeout": "30",
	}

	assert.EqualValues(expected, attr)
}
