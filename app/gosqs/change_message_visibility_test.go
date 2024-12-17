package gosqs

import (
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/stretchr/testify/assert"
)

func TestChangeMessageVisibility_success(t *testing.T) {
	// create a queue
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		models.ResetApp()
	}()

	q := &models.Queue{
		Name: "testing",
		Messages: []models.SqsMessage{{
			MessageBody:   "test1",
			ReceiptHandle: "123",
		}},
	}
	models.SyncQueues.Queues["testing"] = q

	// The default value for the VisibilityTimeout is the zero value of time.Time
	assert.Zero(t, q.Messages[0].VisibilityTimeout)

	_, r := test.GenerateRequestInfo("POST", "/", models.ChangeMessageVisibilityRequest{
		QueueUrl:          "http://localhost:4100/queue/testing",
		ReceiptHandle:     "123",
		VisibilityTimeout: 0,
	}, true)
	status, _ := ChangeMessageVisibilityV1(r)
	assert.Equal(t, status, http.StatusOK)

	// Changing the message visibility increments the time.Time by N seconds
	// from the current time.
	//
	// Given that the current time is relative between calling the endpoint and
	// the time being set, we can't reliably assert an exact value. So assert
	// that the time.Time value is no longer the default zero value.
	assert.NotZero(t, q.Messages[0].VisibilityTimeout)
	assert.NotZero(t, q.Messages[0].ReceiptTime)
	assert.Equal(t, "", q.Messages[0].ReceiptHandle)
	assert.Equal(t, 1, q.Messages[0].Retry)
}

func TestChangeMessageVisibility_success_adds_to_existing_visibility_timeout(t *testing.T) {
	// TODO
}

func TestChangeMessageVisibility_success_transfers_to_dead_letter_queue(t *testing.T) {
	// TODO
}

func TestChangeMessageVisibility_request_transformer_error(t *testing.T) {
	// TODO
}

func TestChangeMessageVisibility_visibility_timeout_too_large(t *testing.T) {
	// TODO
}

func TestChangeMessageVisibility_missing_queue(t *testing.T) {
	// TODO
}

func TestChangeMessageVisibility_missing_message(t *testing.T) {
	// TODO - mismatch receipt handle
}
