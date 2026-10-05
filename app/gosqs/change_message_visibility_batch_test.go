package gosqs

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestChangeMessageVisibilityBatchV1_success_all_found(t *testing.T) {
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT

	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	q := &models.Queue{
		Name: "testing",
		Messages: []models.SqsMessage{
			{
				MessageBody:   "test%20message%20body%201",
				ReceiptHandle: "test1",
			},
			{
				MessageBody:   "test%20message%20body%202",
				ReceiptHandle: "test2",
			},
			{
				MessageBody:   "test%20message%20body%203",
				ReceiptHandle: "test3",
			},
		},
	}
	models.SyncQueues.Queues["testing"] = q

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ChangeMessageVisibilityBatchRequest)
		*v = models.ChangeMessageVisibilityBatchRequest{
			Entries: []models.ChangeMessageVisibilityBatchRequestEntry{
				{Id: "cmv-test-1", ReceiptHandle: "test1", VisibilityTimeout: 30},
				{Id: "cmv-test-2", ReceiptHandle: "test2", VisibilityTimeout: 60},
				{Id: "cmv-test-3", ReceiptHandle: "test3", VisibilityTimeout: 90},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "testing"),
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := ChangeMessageVisibilityBatchV1(r)
	resp := response.(models.ChangeMessageVisibilityBatchResponse)

	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, resp.Result.Successful, 3)
	assert.Empty(t, resp.Result.Failed)
}

func TestChangeMessageVisibilityBatchV1_success_some_not_found(t *testing.T) {
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT

	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	q := &models.Queue{
		Name: "testing",
		Messages: []models.SqsMessage{
			{
				MessageBody:   "test%20message%20body%201",
				ReceiptHandle: "test1",
			},
			{
				MessageBody:   "test%20message%20body%203",
				ReceiptHandle: "test3",
			},
		},
	}
	models.SyncQueues.Queues["testing"] = q

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ChangeMessageVisibilityBatchRequest)
		*v = models.ChangeMessageVisibilityBatchRequest{
			Entries: []models.ChangeMessageVisibilityBatchRequestEntry{
				{Id: "cmv-test-1", ReceiptHandle: "test1", VisibilityTimeout: 30},
				{Id: "cmv-test-2", ReceiptHandle: "test2", VisibilityTimeout: 60},
				{Id: "cmv-test-3", ReceiptHandle: "test3", VisibilityTimeout: 90},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "testing"),
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := ChangeMessageVisibilityBatchV1(r)
	resp := response.(models.ChangeMessageVisibilityBatchResponse)

	assert.Equal(t, http.StatusOK, status)
	assert.Len(t, resp.Result.Successful, 2)
	assert.Len(t, resp.Result.Failed, 1)
	assert.Equal(t, "cmv-test-2", resp.Result.Failed[0].Id)
	assert.Equal(t, "Message not in flight", resp.Result.Failed[0].Message)
	assert.True(t, resp.Result.Failed[0].SenderFault)
}

func TestChangeMessageVisibilityBatchV1_error_queue_not_found(t *testing.T) {
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT

	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ChangeMessageVisibilityBatchRequest)
		*v = models.ChangeMessageVisibilityBatchRequest{
			Entries: []models.ChangeMessageVisibilityBatchRequestEntry{
				{Id: "cmv-test-1", ReceiptHandle: "test1", VisibilityTimeout: 30},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "not-exist-queue"),
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := ChangeMessageVisibilityBatchV1(r)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestChangeMessageVisibilityBatchV1_error_empty_batch(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ChangeMessageVisibilityBatchRequest)
		*v = models.ChangeMessageVisibilityBatchRequest{
			Entries:  make([]models.ChangeMessageVisibilityBatchRequestEntry, 0),
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := ChangeMessageVisibilityBatchV1(r)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestChangeMessageVisibilityBatchV1_error_too_many_entries(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ChangeMessageVisibilityBatchRequest)
		entries := make([]models.ChangeMessageVisibilityBatchRequestEntry, 11)
		for i := range entries {
			entries[i] = models.ChangeMessageVisibilityBatchRequestEntry{
				Id:                fmt.Sprintf("test-%d", i+1),
				ReceiptHandle:     fmt.Sprintf("handle-%d", i+1),
				VisibilityTimeout: 30,
			}
		}
		*v = models.ChangeMessageVisibilityBatchRequest{
			Entries:  entries,
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := ChangeMessageVisibilityBatchV1(r)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestChangeMessageVisibilityBatchV1_error_ids_not_distinct(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ChangeMessageVisibilityBatchRequest)
		*v = models.ChangeMessageVisibilityBatchRequest{
			Entries: []models.ChangeMessageVisibilityBatchRequestEntry{
				{Id: "duplicate-id", ReceiptHandle: "test1", VisibilityTimeout: 30},
				{Id: "duplicate-id", ReceiptHandle: "test2", VisibilityTimeout: 60},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := ChangeMessageVisibilityBatchV1(r)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestChangeMessageVisibilityBatchV1_error_visibility_timeout_too_large(t *testing.T) {
	models.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT

	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	q := &models.Queue{
		Name:     "testing",
		Messages: []models.SqsMessage{},
	}
	models.SyncQueues.Queues["testing"] = q

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.ChangeMessageVisibilityBatchRequest)
		*v = models.ChangeMessageVisibilityBatchRequest{
			Entries: []models.ChangeMessageVisibilityBatchRequestEntry{
				{Id: "cmv-test-1", ReceiptHandle: "test1", VisibilityTimeout: 43201},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "testing"),
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := ChangeMessageVisibilityBatchV1(r)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestChangeMessageVisibilityBatchV1_error_transformer(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		models.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := ChangeMessageVisibilityBatchV1(r)
	assert.Equal(t, http.StatusBadRequest, status)
}
