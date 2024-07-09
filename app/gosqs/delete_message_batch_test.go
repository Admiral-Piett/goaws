package gosqs

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestDeleteMessageBatchV1_success_all_message(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	q := &app.Queue{
		Name: "testing",
		Messages: []app.Message{
			{
				MessageBody:   []byte("test%20message%20body%201"),
				ReceiptHandle: "test1",
			},
			{
				MessageBody:   []byte("test%20message%20body%202"),
				ReceiptHandle: "test2",
			},
			{
				MessageBody:   []byte("test%20message%20body%203"),
				ReceiptHandle: "test3",
			},
		},
	}
	app.SyncQueues.Queues["testing"] = q

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteMessageBatchRequest)
		*v = models.DeleteMessageBatchRequest{
			Entries: []models.DeleteMessageBatchRequestEntry{
				{
					Id:            "delete-test-1",
					ReceiptHandle: "test1",
				},
				{
					Id:            "delete-test-2",
					ReceiptHandle: "test2",
				},
				{
					Id:            "delete-test-3",
					ReceiptHandle: "test3",
				},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "testing"),
		}
		return true
	}
	_, request2 := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)

	status, response2 := DeleteMessageBatchV1(request2)
	deleteMessageBatchResponse := response2.(models.DeleteMessageBatchResponse)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, "delete-test-1", deleteMessageBatchResponse.Result.Successful[0].Id)
	assert.Equal(t, "delete-test-2", deleteMessageBatchResponse.Result.Successful[1].Id)
	assert.Equal(t, "delete-test-3", deleteMessageBatchResponse.Result.Successful[2].Id)
	assert.Empty(t, deleteMessageBatchResponse.Result.Failed)
	assert.Empty(t, app.SyncQueues.Queues["testing"].Messages)
}
func TestDeleteMessageBatchV1_success_not_found_message(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	q := &app.Queue{
		Name: "testing",
		Messages: []app.Message{
			{
				MessageBody:   []byte("test%20message%20body%201"),
				ReceiptHandle: "test1",
			},
			{
				MessageBody:   []byte("test%20message%20body%203"),
				ReceiptHandle: "test3",
			},
		},
	}
	app.SyncQueues.Queues["testing"] = q

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteMessageBatchRequest)
		*v = models.DeleteMessageBatchRequest{
			Entries: []models.DeleteMessageBatchRequestEntry{
				{
					Id:            "delete-test-1",
					ReceiptHandle: "test1",
				},
				{
					Id:            "delete-test-2",
					ReceiptHandle: "test2",
				},
				{
					Id:            "delete-test-3",
					ReceiptHandle: "test3",
				},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "testing"),
		}
		return true
	}
	_, request := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)

	status, response := DeleteMessageBatchV1(request)
	deleteMessageBatchResponse := response.(models.DeleteMessageBatchResponse)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, "delete-test-1", deleteMessageBatchResponse.Result.Successful[0].Id)
	assert.Equal(t, "delete-test-3", deleteMessageBatchResponse.Result.Successful[1].Id)
	assert.Equal(t, "1", deleteMessageBatchResponse.Result.Failed[0].Code)
	assert.Equal(t, "delete-test-2", deleteMessageBatchResponse.Result.Failed[0].Id)
	assert.Equal(t, "Message not found", deleteMessageBatchResponse.Result.Failed[0].Message)
	assert.True(t, deleteMessageBatchResponse.Result.Failed[0].SenderFault)
	assert.Empty(t, app.SyncQueues.Queues["testing"].Messages)
}

func TestDeleteMessageBatchV1_error_not_found_queue(t *testing.T) {
	app.CurrentEnvironment = fixtures.LOCAL_ENVIRONMENT

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteMessageBatchRequest)
		*v = models.DeleteMessageBatchRequest{
			Entries: []models.DeleteMessageBatchRequestEntry{
				{
					Id:            "delete-test-1",
					ReceiptHandle: "test1",
				},
				{
					Id:            "delete-test-2",
					ReceiptHandle: "test2",
				},
				{
					Id:            "delete-test-3",
					ReceiptHandle: "test3",
				},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "not-exist-queue"),
		}
		return true
	}
	_, r := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)

	status, _ := DeleteMessageBatchV1(r)
	assert.Equal(t, status, http.StatusBadRequest)

}

func TestDeleteMessageBatchV1_error_no_entry(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteMessageBatchRequest)
		*v = models.DeleteMessageBatchRequest{
			Entries:  make([]models.DeleteMessageBatchRequestEntry, 0),
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}
	_, r := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)

	status, _ := DeleteMessageBatchV1(r)
	assert.Equal(t, status, http.StatusBadRequest)
}

func TestDeleteMessageBatchV1_error_too_many_entries(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteMessageBatchRequest)
		*v = models.DeleteMessageBatchRequest{
			Entries: []models.DeleteMessageBatchRequestEntry{
				{
					Id:            "test-1",
					ReceiptHandle: "test-1",
				},
				{
					Id:            "test-2",
					ReceiptHandle: "test-2",
				},
				{
					Id:            "test-3",
					ReceiptHandle: "test-3",
				},
				{
					Id:            "test-4",
					ReceiptHandle: "test-4",
				},
				{
					Id:            "test-5",
					ReceiptHandle: "test-5",
				},
				{
					Id:            "test-6",
					ReceiptHandle: "test-6",
				},
				{
					Id:            "test-7",
					ReceiptHandle: "test-7",
				},
				{
					Id:            "test-8",
					ReceiptHandle: "test-8",
				},
				{
					Id:            "test-9",
					ReceiptHandle: "test-9",
				},
				{
					Id:            "test-10",
					ReceiptHandle: "test-10",
				},
				{
					Id:            "test-11",
					ReceiptHandle: "test-11",
				},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}
	_, r := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)

	status, _ := DeleteMessageBatchV1(r)
	assert.Equal(t, status, http.StatusBadRequest)
}

func TestDeleteMessageBatchV1_Error_IdNotDistinct(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")

	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.DeleteMessageBatchRequest)
		*v = models.DeleteMessageBatchRequest{
			Entries: []models.DeleteMessageBatchRequestEntry{
				{
					Id:            "delete-test-1",
					ReceiptHandle: "test1",
				},
				{
					Id:            "delete-test-1",
					ReceiptHandle: "test2",
				},
			},
			QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
		}
		return true
	}
	_, r := test.GenerateRequestInfo(
		"POST",
		"/",
		nil,
		true)

	status, _ := DeleteMessageBatchV1(r)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestDeleteMessageBatchV1_Error_transformer(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := DeleteMessageBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, status)

}
