package gosqs

import (
	"context"
	"fmt"
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

func TestReceiveMessageV1_with_CustomVisibilityTimeout(t *testing.T) {
	// create a queue
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		models.ResetApp()
	}()

	// Create a queue with a default visibility timeout of 30 seconds
	q := &models.Queue{
		Name:             "custom-visibility-queue",
		VisibilityTimeout: 30,
	}
	models.SyncQueues.Queues["custom-visibility-queue"] = q

	// Add a message to the queue
	q.Messages = append(q.Messages, models.SqsMessage{
		MessageBody: "test-message",
		Uuid:        "test-uuid",
	})

	// Test 1: Receive message with custom visibility timeout
	customTimeout := 60 // 60 seconds
	_, r := test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl:          "http://localhost:4100/queue/custom-visibility-queue",
		VisibilityTimeout: customTimeout,
	}, true)

	status, resp := ReceiveMessageV1(r)
	assert.Equal(t, http.StatusOK, status)

	result := resp.GetResult().(models.ReceiveMessageResult)
	assert.Equal(t, 1, len(result.Messages))
	assert.Equal(t, "test-message", string(result.Messages[0].Body))

	// Verify the message in the queue has the custom visibility timeout
	// We can't directly check the exact time, but we can verify it's not using the queue's default
	// by checking that the visibility timeout is greater than now + default timeout - 1 second
	// and less than now + custom timeout + 1 second
	now := time.Now()
	defaultExpiry := now.Add(time.Duration(q.VisibilityTimeout) * time.Second)
	customExpiry := now.Add(time.Duration(customTimeout) * time.Second)

	// The first message should have the custom visibility timeout
	msgVisibilityTimeout := q.Messages[0].VisibilityTimeout
	assert.True(t, msgVisibilityTimeout.After(defaultExpiry.Add(-1*time.Second)),
		"Message visibility timeout should be greater than default timeout")
	assert.True(t, msgVisibilityTimeout.Before(customExpiry.Add(1*time.Second)),
		"Message visibility timeout should be less than custom timeout + 1 second")

	// Test 2: Reset the queue and test with zero visibility timeout (should use queue default)
	models.SyncQueues.Queues["custom-visibility-queue"] = &models.Queue{
		Name:             "custom-visibility-queue",
		VisibilityTimeout: 30,
	}
	q = models.SyncQueues.Queues["custom-visibility-queue"]
	q.Messages = append(q.Messages, models.SqsMessage{
		MessageBody: "test-message-2",
		Uuid:        "test-uuid-2",
	})

	// Receive message with zero visibility timeout (should use queue default)
	_, r = test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl:          "http://localhost:4100/queue/custom-visibility-queue",
		VisibilityTimeout: 0, // Zero should use queue default
	}, true)

	status, resp = ReceiveMessageV1(r)
	assert.Equal(t, http.StatusOK, status)

	// Verify the message in the queue has the default visibility timeout
	now = time.Now()
	defaultExpiry = now.Add(time.Duration(q.VisibilityTimeout) * time.Second)

	// The message should have the default visibility timeout
	msgVisibilityTimeout = q.Messages[0].VisibilityTimeout
	assert.True(t, msgVisibilityTimeout.After(defaultExpiry.Add(-1*time.Second)),
		"Message visibility timeout should be greater than default timeout - 1 second")
	assert.True(t, msgVisibilityTimeout.Before(defaultExpiry.Add(1*time.Second)),
		"Message visibility timeout should be less than default timeout + 1 second")
}

func TestReceiveMessageV1_FIFOSecondMessageAvailableAfterDelete(t *testing.T) {
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		models.ResetApp()
	}()

	queueName := "fifo-delay-queue"
	queueURL := fmt.Sprintf("http://localhost:4100/queue/%s", queueName)
	now := time.Now().Add(-1 * time.Minute)

	// create a queue with a visibility timeout of 10 seconds
	q := &models.Queue{
		Name:                queueName,
		VisibilityTimeout:   10,
		IsFIFO:              true,
		FIFOMessages:        map[string]int{},
		FIFOSequenceNumbers: map[string]int{},
		Duplicates:          map[string]time.Time{},
		Messages: []models.SqsMessage{
			{
				MessageBody: "first",
				Uuid:        "first-uuid",
				GroupID:     "company#worker",
				SentTime:    now,
			},
			{
				MessageBody: "second",
				Uuid:        "second-uuid",
				GroupID:     "company#worker",
				SentTime:    now,
			},
		},
	}
	models.SyncQueues.Queues[queueName] = q

	// receive the first FIFO message only
	_, r := test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl:            queueURL,
		MaxNumberOfMessages: 1,
	}, true)
	status, resp := ReceiveMessageV1(r)
	assert.Equal(t, http.StatusOK, status)
	result := resp.GetResult().(models.ReceiveMessageResult)
	if len(result.Messages) != 1 {
		t.Fatalf("expected to receive the first FIFO message, got %d", len(result.Messages))
	}
	assert.Equal(t, "first", result.Messages[0].Body)

	firstReceipt := result.Messages[0].ReceiptHandle

	// verify the second FIFO message is blocked while the first is in flight
	_, r = test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl: queueURL,
	}, true)
	status, resp = ReceiveMessageV1(r)
	assert.Equal(t, http.StatusOK, status)
	result = resp.GetResult().(models.ReceiveMessageResult)
	if len(result.Messages) != 0 {
		t.Fatalf("expected no FIFO message while the first is outstanding, got %d", len(result.Messages))
	}

	// delete the first FIFO message
	_, deleteReq := test.GenerateRequestInfo("POST", "/", models.DeleteMessageRequest{
		QueueUrl:      queueURL,
		ReceiptHandle: firstReceipt,
	}, true)
	deleteStatus, _ := DeleteMessageV1(deleteReq)
	assert.Equal(t, http.StatusOK, deleteStatus)

	// receive the second FIFO message and ensure it does not wait for full 10 second visibility timeout
	_, r = test.GenerateRequestInfo("POST", "/", models.ReceiveMessageRequest{
		QueueUrl: queueURL,
	}, true)
	start := time.Now()
	status, resp = ReceiveMessageV1(r)
	elapsed := time.Since(start)
	assert.Equal(t, http.StatusOK, status)
	result = resp.GetResult().(models.ReceiveMessageResult)
	if len(result.Messages) != 1 {
		t.Fatalf("expected second FIFO message to be available immediately, got %d", len(result.Messages))
	}
	if elapsed > time.Second {
		t.Fatalf("expected second FIFO message without waiting on visibility timeout, took %s", elapsed)
	}
	assert.Equal(t, "second", result.Messages[0].Body)

	// delete the second FIFO message
	_, deleteReq = test.GenerateRequestInfo("POST", "/", models.DeleteMessageRequest{
		QueueUrl:      queueURL,
		ReceiptHandle: result.Messages[0].ReceiptHandle,
	}, true)
	deleteStatus, _ = DeleteMessageV1(deleteReq)
	assert.Equal(t, http.StatusOK, deleteStatus)

	if len(q.Messages) != 0 {
		t.Fatalf("expected all FIFO messages to be deleted, remaining %d", len(q.Messages))
	}
}

// TODO - other tests
