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

func TestDeleteMessage(t *testing.T) {
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

	_, r := test.GenerateRequestInfo("POST", "/", models.DeleteMessageRequest{
		QueueUrl:      "http://localhost:4100/queue/testing",
		ReceiptHandle: "123",
	}, true)
	status, _ := DeleteMessageV1(r)

	assert.Equal(t, status, http.StatusOK)
	assert.Empty(t, q.Messages)
}
