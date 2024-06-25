package gosqs

import (
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/stretchr/testify/assert"
)

func TestChangeMessageVisibility_POST_SUCCESS(t *testing.T) {
	// create a queue
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
	}()

	q := &app.Queue{
		Name: "testing",
		Messages: []app.Message{{
			MessageBody:   []byte("test1"),
			ReceiptHandle: "123",
		}},
	}
	app.SyncQueues.Queues["testing"] = q

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
}
