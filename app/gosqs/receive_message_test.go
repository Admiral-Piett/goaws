package gosqs

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/stretchr/testify/assert"
)

// TODO Admiral-Piett - fix all these tests, we don't want to be waiting for 5 seconds plus like this.
func TestReceiveMessageV1_with_WaitTimeEnforced(t *testing.T) {
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		models.ResetApp()
	}()

	q := &models.Queue{
		Name:                          "waiting-queue",
		ReceiveMessageWaitTimeSeconds: 2,
		//MaximumMessageSize:            262144,
	}
	models.SyncQueues.Queues["waiting-queue"] = q

	// receive message ensure delay
	_, r := test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl: "http://localhost:4100/queue/waiting-queue",
	}, true)

	start := time.Now()
	status, response := ReceiveMessageV1(r)
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusOK, status)
	if elapsed < 2*time.Second {
		t.Fatalf("handler didn't wait ReceiveMessageWaitTimeSeconds %s", response)
	}

	// mock sending a message
	q.Messages = append(q.Messages, models.SqsMessage{MessageBody: "1"})

	// receive message
	_, r = test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl: "http://localhost:4100/queue/waiting-queue",
	}, true)
	start = time.Now()
	status, resp := ReceiveMessageV1(r)
	elapsed = time.Since(start)

	assert.Equal(t, http.StatusOK, status)
	if elapsed > 1*time.Second {
		t.Fatal("handler waited when message was available, expected not to wait")
	}

	assert.Equal(t, "1", string(resp.GetResult().(models.ReceiveMessageResult).Messages[0].Body))
}

func TestReceiveMessageV1_CanceledByClient(t *testing.T) {
	// create a queue
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		models.ResetApp()
	}()

	q := &models.Queue{
		Name:                          "cancel-queue",
		ReceiveMessageWaitTimeSeconds: 20,
	}
	models.SyncQueues.Queues["cancel-queue"] = q

	var wg sync.WaitGroup
	ctx, cancelReceive := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		// receive message (that will be canceled)
		_, r := test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
			QueueUrl: "http://localhost:4100/queue/cancel-queue",
		}, true)
		r = r.WithContext(ctx)

		status, resp := ReceiveMessageV1(r)
		assert.Equal(t, http.StatusOK, status)

		if len(resp.GetResult().(models.ReceiveMessageResult).Messages) != 0 {
			t.Fatal("expecting this ReceiveMessage() to not pickup this message as it should canceled before the Send()")
		}
	}()
	time.Sleep(100 * time.Millisecond) // let enought time for the Receive go to wait mode
	cancelReceive()                    // cancel the first ReceiveMessage(), make sure it will not pickup the sent message below
	time.Sleep(5 * time.Millisecond)

	// send a message
	_, r := test.GenerateRequestInfo("POST", "/", models.SendMessageRequest{
		QueueUrl:    "http://localhost:4100/queue/cancel-queue",
		MessageBody: "12345",
	}, true)
	status, _ := SendMessageV1(r)
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v", status, http.StatusOK)
	}

	// receive message
	_, r = test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl: "http://localhost:4100/queue/cancel-queue",
	}, true)
	start := time.Now()
	status, resp := ReceiveMessageV1(r)
	assert.Equal(t, http.StatusOK, status)
	elapsed := time.Since(start)

	result, ok := resp.GetResult().(models.ReceiveMessageResult)
	if !ok {
		t.Fatal("handler should return a message")
	}

	if len(result.Messages) == 0 || string(result.Messages[0].Body) == "12345\n" {
		t.Fatal("handler should return a message")
	}
	if elapsed > 1*time.Second {
		t.Fatal("handler waited when message was available, expected not to wait")
	}

	if timedout := waitTimeout(&wg, 2*time.Second); timedout {
		t.Errorf("expected ReceiveMessage() in goroutine to exit quickly due to cancelReceive() called")
	}
}

func TestReceiveMessageV1_with_DelaySeconds(t *testing.T) {
	// create a queue
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		models.ResetApp()
	}()

	q := &models.Queue{
		Name:         "delay-seconds-queue",
		DelaySeconds: 2,
	}
	models.SyncQueues.Queues["delay-seconds-queue"] = q

	// send a message
	_, r := test.GenerateRequestInfo("POST", "/", models.SendMessageRequest{
		QueueUrl:    "http://localhost:4100/queue/delay-seconds-queue",
		MessageBody: "1",
	}, true)
	status, _ := SendMessageV1(r)
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v", status, http.StatusOK)
	}

	// receive message before delay is up
	_, r = test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{QueueUrl: "http://localhost:4100/queue/delay-seconds-queue"}, true)
	status, _ = ReceiveMessageV1(r)
	assert.Equal(t, http.StatusOK, status)

	// receive message with wait should return after delay
	_, r = test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl:        "http://localhost:4100/queue/delay-seconds-queue",
		WaitTimeSeconds: 10,
	}, true)
	start := time.Now()
	status, _ = ReceiveMessageV1(r)
	elapsed := time.Since(start)
	assert.Equal(t, http.StatusOK, status)
	if elapsed < 1*time.Second {
		t.Errorf("handler didn't wait at all")
	}
	if elapsed > 4*time.Second {
		t.Errorf("handler didn't need to wait all WaitTimeSeconds=10, only DelaySeconds=2")
	}
}

func TestReceiveMessageV1_with_MessageAttributes(t *testing.T) {
	// create a queue
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		models.ResetApp()
	}()

	q := &models.Queue{Name: "waiting-queue"}
	models.SyncQueues.Queues["waiting-queue"] = q

	// send a message
	q.Messages = append(q.Messages, models.SqsMessage{
		MessageBody: "1",
		MessageAttributes: map[string]models.MessageAttribute{
			"TestMessageAttrName": {
				DataType:    "String",
				StringValue: "TestMessageAttrValue",
			},
		},
	})

	// receive message
	_, r := test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{QueueUrl: "http://localhost:4100/queue/waiting-queue"}, true)
	status, resp := ReceiveMessageV1(r)
	result := resp.GetResult().(models.ReceiveMessageResult)

	assert.NotEmpty(t, result.Messages[0].Attributes["ApproximateFirstReceiveTimestamp"])
	assert.NotEmpty(t, result.Messages[0].Attributes["SenderId"])
	assert.NotEmpty(t, result.Messages[0].Attributes["ApproximateReceiveCount"])
	assert.NotEmpty(t, result.Messages[0].Attributes["SentTimestamp"])

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "1", string(result.Messages[0].Body))
	assert.Equal(t, 1, len(result.Messages[0].MessageAttributes))
	assert.Equal(t, "String", result.Messages[0].MessageAttributes["TestMessageAttrName"].DataType)
	assert.Equal(t, "TestMessageAttrValue", result.Messages[0].MessageAttributes["TestMessageAttrName"].StringValue)
}

func TestReceiveMessageV1_request_transformer_error(t *testing.T) {
	// TODO
}

// TODO - other tests
