package gosqs

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/p4tin/goaws/app"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestListQueues_POST_NoQueues(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	handler := http.HandlerFunc(ListQueues)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	const expected = "<ListQueuesResult></ListQueuesResult>"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func prepareRequest(t *testing.T, method string, url string, body io.Reader) (*assert.Assertions, *httptest.ResponseRecorder, *http.Request) {
	assert := assert.New(t)
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest(method, url, body)
	assert.NoError(err)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	recorder := httptest.NewRecorder()
	return assert, recorder, req
}

func TestListQueues_POST_Success(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	handler := http.HandlerFunc(ListQueues)

	app.SyncQueues.Queues["foo"] = &app.Queue{Name: "foo", URL: "http://:/queue/foo"}
	app.SyncQueues.Queues["bar"] = &app.Queue{Name: "bar", URL: "http://:/queue/bar"}
	app.SyncQueues.Queues["foobar"] = &app.Queue{Name: "foobar", URL: "http://:/queue/foobar"}

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = "<QueueUrl>http://:/queue/bar</QueueUrl>"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")

	// Arrange
	// Filter lists by the given QueueNamePrefix
	form := url.Values{}
	form.Add("QueueNamePrefix", "fo")
	req, _ = http.NewRequest("POST", "/", nil)
	req.PostForm = form
	rr = httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	assert.NotContainsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestCreateQueuehandler_POST_CreateQueue(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	const queueName = "UnitTestQueue1"
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "UnitTestQueue1")
	form.Add("Attribute.1.Name", "VisibilityTimeout")
	form.Add("Attribute.1.Value", "60")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler := http.HandlerFunc(CreateQueue)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = queueName
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")

	actualQueue := app.SyncQueues.Queues[queueName]
	assert.Equal(queueName, actualQueue.Name)
	assert.Equal("http://://"+queueName, actualQueue.URL)
	assert.Equal("arn:aws:sqs:::"+queueName, actualQueue.Arn)
	assert.Equal(60, actualQueue.TimeoutSecs)
}

func TestCreateFIFOQueuehandler_POST_CreateQueue(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	const queueName = "UnitTestQueue1.fifo"
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "UnitTestQueue1.fifo")
	form.Add("Attribute.1.Name", "VisibilityTimeout")
	form.Add("Attribute.1.Value", "60")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	handler := http.HandlerFunc(CreateQueue)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = queueName
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")

	actualQueue := app.SyncQueues.Queues[queueName]
	assert.Equal(queueName, actualQueue.Name)
	assert.Equal("http://://"+queueName, actualQueue.URL)
	assert.Equal("arn:aws:sqs:::"+queueName, actualQueue.Arn)
	assert.Equal(60, actualQueue.TimeoutSecs)
	assert.True(actualQueue.IsFIFO)
}

func TestSendQueue_POST_NotExists(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	form := url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/NON-EXISTANT")
	form.Add("MessageBody", "Test123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	handler := http.HandlerFunc(SendMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusBadRequest,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = "NonExistentQueue"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestSendMessageBatch_POST_QueueNotFound(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(SendMessageBatch)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusBadRequest,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = "NonExistentQueue"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestSendMessageBatch_POST_NoEntry(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	app.SyncQueues.Queues["testing"] = &app.Queue{Name: "testing"}
	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(SendMessageBatch)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusBadRequest,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = "EmptyBatchRequest"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")

	// Arrange
	req, _ = http.NewRequest("POST", "/", nil)
	form.Add("SendMessageBatchRequestEntry", "")
	req.PostForm = form
	rr = httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusBadRequest,
		rr.Code,
		"handler returned wrong status code")

	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestSendMessageBatch_POST_IdNotDistinct(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	app.SyncQueues.Queues["testing"] = &app.Queue{Name: "testing"}
	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("SendMessageBatchRequestEntry.1.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.1.MessageBody", "test%20message%20body%201")
	form.Add("SendMessageBatchRequestEntry.2.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.2.MessageBody", "test%20message%20body%202")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(SendMessageBatch)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusBadRequest,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = "BatchEntryIdsNotDistinct"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestSendMessageBatch_POST_TooManyEntries(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	app.SyncQueues.Queues["testing"] = &app.Queue{Name: "testing"}
	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("SendMessageBatchRequestEntry.1.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.1.MessageBody", "test%20message%20body%201")
	form.Add("SendMessageBatchRequestEntry.2.Id", "test_msg_002")
	form.Add("SendMessageBatchRequestEntry.2.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.3.Id", "test_msg_003")
	form.Add("SendMessageBatchRequestEntry.3.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.4.Id", "test_msg_004")
	form.Add("SendMessageBatchRequestEntry.4.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.5.Id", "test_msg_005")
	form.Add("SendMessageBatchRequestEntry.5.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.6.Id", "test_msg_006")
	form.Add("SendMessageBatchRequestEntry.6.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.7.Id", "test_msg_007")
	form.Add("SendMessageBatchRequestEntry.7.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.8.Id", "test_msg_008")
	form.Add("SendMessageBatchRequestEntry.8.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.9.Id", "test_msg_009")
	form.Add("SendMessageBatchRequestEntry.9.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.10.Id", "test_msg_010")
	form.Add("SendMessageBatchRequestEntry.10.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.11.Id", "test_msg_011")
	form.Add("SendMessageBatchRequestEntry.11.MessageBody", "test%20message%20body%202")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(SendMessageBatch)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusBadRequest,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = "TooManyEntriesInBatchRequest"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestSendMessageBatch_POST_Success(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	app.SyncQueues.Queues["testing"] = &app.Queue{Name: "testing"}
	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("SendMessageBatchRequestEntry.1.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.1.MessageBody", "test%20message%20body%201")
	form.Add("SendMessageBatchRequestEntry.2.Id", "test_msg_002")
	form.Add("SendMessageBatchRequestEntry.2.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Name", "test_attribute_name_1")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Value.StringValue", "test_attribute_value_1")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Value.DataType", "String")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(SendMessageBatch)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = "<MD5OfMessageBody>1c538b76fce1a234bce865025c02b042</MD5OfMessageBody>"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestSendMessageBatchToFIFOQueue_POST_Success(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	app.SyncQueues.Queues["testing.fifo"] = &app.Queue{
		Name:   "testing.fifo",
		IsFIFO: true,
	}
	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing.fifo")
	form.Add("SendMessageBatchRequestEntry.1.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.1.MessageGroupId", "GROUP-X")
	form.Add("SendMessageBatchRequestEntry.1.MessageBody", "test%20message%20body%201")
	form.Add("SendMessageBatchRequestEntry.2.Id", "test_msg_002")
	form.Add("SendMessageBatchRequestEntry.2.MessageGroupId", "GROUP-X")
	form.Add("SendMessageBatchRequestEntry.2.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Name", "test_attribute_name_1")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Value.StringValue", "test_attribute_value_1")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Value.DataType", "String")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(SendMessageBatch)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = "<MD5OfMessageBody>1c538b76fce1a234bce865025c02b042</MD5OfMessageBody>"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestChangeMessageVisibility_POST_SUCCESS(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	app.SyncQueues.Queues["testing"] = &app.Queue{Name: "testing"}
	app.SyncQueues.Queues["testing"].Messages = []app.Message{{
		MessageBody:   []byte("test1"),
		ReceiptHandle: "123",
	}}
	form := url.Values{}
	form.Add("Action", "ChangeMessageVisibility")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("VisibilityTimeout", "0")
	form.Add("ReceiptHandle", "123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(ChangeMessageVisibility)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Check the response body is what we expect.
	const expected = `<ChangeMessageVisibilityResult xmlns="http://queue.amazonaws.com/doc/2012-11-05/">`
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestRequeue_VisibilityTimeoutExpires(t *testing.T) {
	done := make(chan struct{}, 0)
	go PeriodicTasks(1*time.Second, done)

	// Arrange
	log.Info("Create queue")
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "requeue")
	form.Add("Attribute.1.Name", "VisibilityTimeout")
	form.Add("Attribute.1.Value", "1")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(CreateQueue)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	log.Info("Send message to queue")
	req, err := http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue")
	form.Add("MessageBody", "1")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(SendMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// receive message
	log.Info("Receive message from queue immediately")
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	const expected = "<Message>"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler should return a message")

	// Arrange
	// try to receive another message.
	log.Info("Try to receive message from queue second time")
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	assert.NotContainsf(
		rr.Body.String(),
		expected,
		"handler should not return a message")

	log.Info("Waiting for 2 sec")
	time.Sleep(2 * time.Second)

	// Arrange
	// message needs to be requeue
	log.Info("Try to receive message from queue third time")
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler should return a message")
	done <- struct{}{}
}

func TestRequeueing_ResetVisibilityTimeout(t *testing.T) {
	done := make(chan struct{}, 0)
	go PeriodicTasks(1*time.Second, done)

	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "requeue-reset")
	form.Add("Attribute.1.Name", "VisibilityTimeout")
	form.Add("Attribute.1.Value", "10")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(CreateQueue)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// send a message
	req, err := http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset")
	form.Add("MessageBody", "1")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(SendMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// receive message
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	const expected = "<Message>"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler should return a message")

	resp := app.ReceiveMessageResponse{}
	err = xml.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(err, "unexpected unmarshal error: %s", err)
	assert.NotEmptyf(resp.Result.Message, "messages should not be empty")
	receiptHandle := ""
	if len(resp.Result.Message) > 0 {
		receiptHandle = resp.Result.Message[0].ReceiptHandle
	}

	// Arrange
	// try to receive another message.
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	assert.NotContainsf(
		rr.Body.String(),
		expected,
		"handler should not return a message")

	// Arrange
	// reset message visibility timeout to requeue it
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ChangeMessageVisibility")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset")
	form.Add("VisibilityTimeout", "0")
	form.Add("ReceiptHandle", receiptHandle)
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ChangeMessageVisibility)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// message needs to be requeue
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler should return a message")
	done <- struct{}{}
}

func TestDeadLetterQueue(t *testing.T) {
	done := make(chan struct{}, 0)
	go PeriodicTasks(1*time.Second, done)

	// Arrange
	// create a queue
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	deadLetterQueue := &app.Queue{
		Name:     "failed-messages",
		Messages: []app.Message{},
	}
	app.SyncQueues.Lock()
	app.SyncQueues.Queues["failed-messages"] = deadLetterQueue
	app.SyncQueues.Unlock()
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "testing-deadletter")
	form.Add("Attribute.1.Name", "VisibilityTimeout")
	form.Add("Attribute.1.Value", "1")
	form.Add("Attribute.2.Name", "RedrivePolicy")
	form.Add("Attribute.2.Value", `{"maxReceiveCount": 1, "deadLetterTargetArn":"arn:aws:sqs::000000000000:failed-messages"}`)
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(CreateQueue)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// send a message
	req, err := http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing-deadletter")
	form.Add("MessageBody", "1")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(SendMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// receive message
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing-deadletter")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	const expected = "<Message>"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler should return a message")

	log.Info("Waiting for 2 sec")
	time.Sleep(2 * time.Second)

	// Arrange
	// receive the message one more time
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler should return a message")

	log.Info("Waiting for 2 sec")
	time.Sleep(2 * time.Second)

	// Arrange
	// another receive attempt
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	assert.NotContainsf(
		rr.Body.String(),
		expected,
		"handler should not return a message")
	assert.NotEmpty(deadLetterQueue.Messages, "expected a message")
}

func TestReceiveMessageWaitTimeEnforced(t *testing.T) {
	// Arrange
	// create a queue
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "waiting-queue")
	form.Add("Attribute.1.Name", "ReceiveMessageWaitTimeSeconds")
	form.Add("Attribute.1.Value", "2")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(CreateQueue)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// receive message ensure delay
	req, err := http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/waiting-queue")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	start := time.Now()
	handler.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	const expected = "<Message>"
	assert.NotContainsf(
		rr.Body.String(),
		expected,
		"handler should not return a message")
	assert.Condition(
		func() bool { return elapsed >= 2*time.Second },
		"handler didn't wait ReceiveMessageWaitTimeSeconds")

	// Arrange
	// send a message
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/waiting-queue")
	form.Add("MessageBody", "1")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(SendMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// receive message
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/waiting-queue")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	start = time.Now()
	handler.ServeHTTP(rr, req)
	elapsed = time.Since(start)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler should return a message")
	assert.Condition(
		func() bool { return elapsed < 1*time.Second },
		"handler waited when message was available, expected not to wait")
}
func TestReceiveMessage_CanceledByClient(t *testing.T) {
	// Arrange
	// create a queue
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "cancel-queue")
	form.Add("Attribute.1.Name", "ReceiveMessageWaitTimeSeconds")
	form.Add("Attribute.1.Value", "20")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(CreateQueue)

	// Act
	handler.ServeHTTP(rr, req)

	var wg sync.WaitGroup
	ctx, cancelReceive := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		// receive message (that will be canceled)
		assert, rr, req := prepareRequest(t, "POST", "/", nil)
		req = req.WithContext(ctx)
		form := url.Values{}
		form.Add("Action", "ReceiveMessage")
		form.Add("QueueUrl", "http://localhost:4100/queue/cancel-queue")
		form.Add("Version", "2012-11-05")
		req.PostForm = form
		handler := http.HandlerFunc(ReceiveMessage)

		// Act
		handler.ServeHTTP(rr, req)

		// Asserts
		assert.Equalf(
			http.StatusOK,
			rr.Code,
			"handler returned wrong status code")
		const expected = "12345"
		assert.NotContainsf(
			rr.Body.String(),
			expected,
			"expecting this ReceiveMessage() to not pickup this message as it should canceled before the Send()")
	}()

	time.Sleep(100 * time.Millisecond) // let enough time for the Receive go to wait mode
	cancelReceive()                    // cancel the first ReceiveMessage(), make sure it will not pickup the sent message below
	time.Sleep(5 * time.Millisecond)

	// Arrange
	// send a message
	req, err := http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/cancel-queue")
	form.Add("MessageBody", "12345")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(SendMessage)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	// Arrange
	// receive message
	req, err = http.NewRequest("POST", "/", nil)
	assert.NoError(err)
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/cancel-queue")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(ReceiveMessage)

	// Act
	start := time.Now()
	handler.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")
	const expected = "12345"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler should return a message")
	assert.Conditionf(
		func() bool { return elapsed <= 1*time.Second },
		"handler waited when message was available, expected not to wait")
	assert.Falsef(
		waitTimeout(&wg, 2*time.Second),
		"expected ReceiveMessage() in goroutine to exit quickly due to cancelReceive() called")
}

func TestReceiveMessage_WithConcurrentDeleteQueue(t *testing.T) {
	// Arrange
	// create a queue
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "waiting-queue")
	form.Add("Attribute.1.Name", "ReceiveMessageWaitTimeSeconds")
	form.Add("Attribute.1.Value", "1")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(CreateQueue)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusOK,
		rr.Code,
		"handler returned wrong status code")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Arrange
		// receive message
		assert, rr, req := prepareRequest(t, "POST", "/", nil)
		form := url.Values{}
		form.Add("Action", "ReceiveMessage")
		form.Add("QueueUrl", "http://localhost:4100/queue/waiting-queue")
		form.Add("Version", "2012-11-05")
		req.PostForm = form
		handler := http.HandlerFunc(ReceiveMessage)

		// Act
		handler.ServeHTTP(rr, req)

		// Asserts
		assert.Equalf(
			http.StatusBadRequest,
			rr.Code,
			"handler returned wrong status code")

		// Check the response body is what we expect.
		const expected = "Not Found"
		assert.Containsf(
			rr.Body.String(),
			expected,
			"handler returned unexpected body")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // 10ms to let the ReceiveMessage() block

		// Arrange
		// delete queue message
		req, err := http.NewRequest("POST", "/", nil)
		assert.NoError(err)
		form := url.Values{}
		form.Add("Action", "DeleteQueue")
		form.Add("QueueUrl", "http://localhost:4100/queue/waiting-queue")
		form.Add("Version", "2012-11-05")
		req.PostForm = form
		rr := httptest.NewRecorder()
		handler = http.HandlerFunc(DeleteQueue)

		// Act
		handler.ServeHTTP(rr, req)

		// Asserts
		assert.Equalf(
			http.StatusOK,
			rr.Code,
			"handler returned wrong status code")
	}()

	// Asserts
	assert.Falsef(
		waitTimeout(&wg, 2*time.Second),
		"concurrent handlers timeout, expecting both to return within timeout")
}

func TestSetQueueAttributes_POST_QueueNotFound(t *testing.T) {
	// Arrange
	assert, rr, req := prepareRequest(t, "POST", "/", nil)
	form := url.Values{}
	form.Add("Action", "SetQueueAttributes")
	form.Add("QueueUrl", "http://localhost:4100/queue/not-existing")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	handler := http.HandlerFunc(SetQueueAttributes)

	// Act
	handler.ServeHTTP(rr, req)

	// Asserts
	assert.Equalf(
		http.StatusBadRequest,
		rr.Code,
		"handler returned wrong status code")
	const expected = "NonExistentQueue"
	assert.Containsf(
		rr.Body.String(),
		expected,
		"handler returned unexpected body")
}

func TestSendingAndReceivingFromFIFOQueueReturnsSameMessageOnError(t *testing.T) {
	done := make(chan struct{}, 0)
	go PeriodicTasks(1*time.Second, done)

	// create a queue
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "requeue-reset.fifo")
	form.Add("Attribute.1.Name", "VisibilityTimeout")
	form.Add("Attribute.1.Value", "2")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	rr := httptest.NewRecorder()
	http.HandlerFunc(CreateQueue).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}

	// send a message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("MessageBody", "1")
	form.Add("MessageGroupId", "GROUP-X")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	rr = httptest.NewRecorder()
	http.HandlerFunc(SendMessage).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}

	// send a message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("MessageBody", "2")
	form.Add("MessageGroupId", "GROUP-X")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	rr = httptest.NewRecorder()
	http.HandlerFunc(SendMessage).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}

	// receive message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	rr = httptest.NewRecorder()
	http.HandlerFunc(ReceiveMessage).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if ok := strings.Contains(rr.Body.String(), "<Message>"); !ok {
		t.Fatal("handler should return a message")
	}

	resp := app.ReceiveMessageResponse{}
	err = xml.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %s", err)
	}
	receiptHandleFirst := resp.Result.Message[0].ReceiptHandle
	if string(resp.Result.Message[0].Body) != "1" {
		t.Fatalf("should have received body 1: %s", err)
	}

	// try to receive another message and we should get none
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	rr = httptest.NewRecorder()
	http.HandlerFunc(ReceiveMessage).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if ok := strings.Contains(rr.Body.String(), "<Message>"); ok {
		t.Fatal("handler should not return a message")
	}

	if len(app.SyncQueues.Queues["requeue-reset.fifo"].FIFOMessages) != 1 {
		t.Fatal("there should be only 1 group locked")
	}

	if app.SyncQueues.Queues["requeue-reset.fifo"].FIFOMessages["GROUP-X"] != 0 {
		t.Fatal("there should be GROUP-X locked")
	}

	// remove message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "DeleteMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("ReceiptHandle", receiptHandleFirst)
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	rr = httptest.NewRecorder()
	http.HandlerFunc(DeleteMessage).ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if len(app.SyncQueues.Queues["requeue-reset.fifo"].Messages) != 1 {
		t.Fatal("there should be only 1 message in queue")
	}

	// receive message - loop until visibility timeouts
	for {
		req, err = http.NewRequest("POST", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		form = url.Values{}
		form.Add("Action", "ReceiveMessage")
		form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
		form.Add("Version", "2012-11-05")
		req.PostForm = form

		rr = httptest.NewRecorder()
		http.HandlerFunc(ReceiveMessage).ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got \n%v want %v",
				status, http.StatusOK)
		}
		if ok := strings.Contains(rr.Body.String(), "<Message>"); !ok {
			continue
		}

		resp = app.ReceiveMessageResponse{}
		err = xml.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %s", err)
		}
		if string(resp.Result.Message[0].Body) != "2" {
			t.Fatalf("should have received body 2: %s", err)
		}
		break
	}

	done <- struct{}{}
}

// waitTimeout waits for the wait group for the specified max timeout.
// Returns true if waiting timed out.
// credits: https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
