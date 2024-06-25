package gosqs

import (
	"net/http"
	"testing"
	"time"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestSendMessageV1_Success(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_success := models.SendMessageRequest{
		QueueUrl:    "http://localhost:4200/new-queue-1",
		MessageBody: "Test Message",
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageRequest)
		*v = sendMessageRequest_success
		return true
	}

	q := &app.Queue{
		Name:               "new-queue-1",
		MaximumMessageSize: 1024,
	}
	app.SyncQueues.Queues["new-queue-1"] = q

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageV1(r)

	// Check the queue
	assert.Equal(t, 1, len(q.Messages))
	msg := q.Messages[0]
	assert.Equal(t, "Test Message", string(msg.MessageBody))

	// Check the response
	assert.Equal(t, http.StatusOK, status)
	sendMessageResponse, ok := response.(models.SendMessageResponse)
	assert.True(t, ok)
	assert.NotEmpty(t, sendMessageResponse.Result.MD5OfMessageBody)
	// No FIFO Sequence
	assert.Empty(t, sendMessageResponse.Result.SequenceNumber)
}

func TestSendMessageV1_Success_FIFOQueue(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_success := models.SendMessageRequest{
		QueueUrl:    "http://localhost:4200/new-queue-1",
		MessageBody: "Test Message",
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageRequest)
		*v = sendMessageRequest_success
		return true
	}

	q := &app.Queue{
		Name:               "new-queue-1",
		MaximumMessageSize: 1024,
		IsFIFO:             true,
	}
	app.SyncQueues.Queues["new-queue-1"] = q

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageV1(r)

	// Check the queue
	assert.Equal(t, 1, len(q.Messages))
	msg := q.Messages[0]
	assert.Equal(t, "Test Message", string(msg.MessageBody))

	// Check the response
	assert.Equal(t, http.StatusOK, status)
	sendMessageResponse, ok := response.(models.SendMessageResponse)
	assert.True(t, ok)
	assert.NotEmpty(t, sendMessageResponse.Result.MD5OfMessageBody)
	// Should have FIFO Sequence
	assert.NotEmpty(t, sendMessageResponse.Result.SequenceNumber)
}

func TestSendMessageV1_Success_Deduplication(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_success := models.SendMessageRequest{
		QueueUrl:               "http://localhost:4200/new-queue-1",
		MessageBody:            "Test Message",
		MessageDeduplicationId: "1",
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageRequest)
		*v = sendMessageRequest_success
		return true
	}

	q := &app.Queue{
		Name:               "new-queue-1",
		MaximumMessageSize: 1024,
		IsFIFO:             true,
		EnableDuplicates:   true,
		Duplicates:         make(map[string]time.Time),
	}
	app.SyncQueues.Queues["new-queue-1"] = q

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := SendMessageV1(r)

	// Check the queue
	assert.Equal(t, 1, len(q.Messages))
	// Check the response
	assert.Equal(t, http.StatusOK, status)

	// Send the same message (have DeduplicationId)
	status, _ = SendMessageV1(r)
	// Response is "success"
	assert.Equal(t, http.StatusOK, status)
	// Only 1 message should be in the queue
	assert.Equal(t, 1, len(q.Messages))
}

func TestSendMessageV1_request_transformer_error(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SendMessageV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSendMessageV1_MaximumMessageSize_MessageTooBig(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_success := models.SendMessageRequest{
		QueueUrl:    "http://localhost:4200/new-queue-1",
		MessageBody: "Test Message",
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageRequest)
		*v = sendMessageRequest_success
		return true
	}

	q := &app.Queue{
		Name:               "new-queue-1",
		MaximumMessageSize: 1,
	}
	app.SyncQueues.Queues["new-queue-1"] = q

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageV1(r)

	// Check the response
	assert.Equal(t, http.StatusBadRequest, status)
	errorResponse, ok := response.(models.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, "MessageTooBig", errorResponse.Result.Type)
}

func TestSendMessageV1_POST_QueueNonExistant(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_success := models.SendMessageRequest{
		QueueUrl:    "http://localhost:4200/new-queue-1",
		MessageBody: "Test Message",
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageRequest)
		*v = sendMessageRequest_success
		return true
	}

	// No test queue is added to app.SyncQueues

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageV1(r)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusBadRequest, status)

	// Check the response body is what we expect.
	errorResponse, ok := response.(models.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, "Not Found", errorResponse.Result.Type)
}
