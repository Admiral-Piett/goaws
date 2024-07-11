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

func TestSendMessageBatchV1_Success(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_success := models.SendMessageBatchRequest{
		Entries: []models.SendMessageBatchRequestEntry{
			{
				Id:          "test-msg-with-non-attribute",
				MessageBody: "test%20message%20body%201",
			},
			{
				Id:          "test-msg-with-single-attirbute",
				MessageBody: "test%20message%20body%202",
				MessageAttributes: map[string]models.MessageAttributeValue{
					"my-attribute-name": {
						BinaryValue: "base64-encoded-value",
						DataType:    "hogehoge",
						StringValue: "my-attribute-string-value",
					},
				},
			},
			{
				Id:          "test-msg-with-multi-attirbute",
				MessageBody: "test%20message%20body%203",
				MessageAttributes: map[string]models.MessageAttributeValue{
					"my-attribute-name-1": {
						BinaryValue: "base64-encoded-value-1",
						DataType:    "hogehoge",
						StringValue: "my-attribute-string-value-1",
					},
					"my-attribute-name-2": {
						BinaryValue: "base64-encoded-value-2",
						DataType:    "hogehoge",
						StringValue: "my-attribute-string-value-2",
					},
				},
			},
		},
		QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageBatchRequest)
		*v = sendMessageRequest_success
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageBatchV1(r)
	sendMessageBatchResponse, ok := response.(models.SendMessageBatchResponse)

	assert.Equal(t, http.StatusOK, status)
	assert.True(t, ok)

	resultEntry := sendMessageBatchResponse.Result.Entry
	assert.Equal(t, 3, len(resultEntry))
	assert.Contains(t, resultEntry[0].Id, "test-msg-with-non-attribute")
	assert.Contains(t, resultEntry[1].Id, "test-msg-with-single-attirbute")
	assert.Contains(t, resultEntry[2].Id, "test-msg-with-multi-attirbute")
	assert.Empty(t, resultEntry[0].SequenceNumber)
	assert.Empty(t, resultEntry[1].SequenceNumber)
	assert.Empty(t, resultEntry[2].SequenceNumber)

}

func TestSendMessageBatchV1_Success_Fifo_Queue(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_success := models.SendMessageBatchRequest{
		Entries: []models.SendMessageBatchRequestEntry{
			{
				Id:          "test_msg_001",
				MessageBody: "test%20message%20body%201",
			},
			{
				Id:          "test_msg_002",
				MessageBody: "test%20message%20body%202",
			},
			{
				Id:          "test_msg_003",
				MessageBody: "test%20message%20body%203",
			},
		},
		QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "fifo-queue-1"),
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageBatchRequest)
		*v = sendMessageRequest_success
		return true
	}

	q := &app.Queue{
		Name:               "fifo-queue-1",
		MaximumMessageSize: 1024,
		IsFIFO:             true,
	}
	app.SyncQueues.Queues["fifo-queue-1"] = q

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageBatchV1(r)
	sendMessageBatchResponse, ok := response.(models.SendMessageBatchResponse)

	assert.Equal(t, http.StatusOK, status)
	assert.True(t, ok)

	resultEntry := sendMessageBatchResponse.Result.Entry
	assert.Equal(t, 3, len(resultEntry))
	assert.Contains(t, resultEntry[0].Id, "test_msg_001")
	assert.NotEmpty(t, resultEntry[0].SequenceNumber)
	assert.Contains(t, resultEntry[1].Id, "test_msg_002")
	assert.NotEmpty(t, resultEntry[1].SequenceNumber)
	assert.Contains(t, resultEntry[2].Id, "test_msg_003")
	assert.NotEmpty(t, resultEntry[2].SequenceNumber)
}

func TestSendMessageBatchV1_Error_QueueNotFound(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_not_found_queue := models.SendMessageBatchRequest{
		Entries: []models.SendMessageBatchRequestEntry{
			{
				Id:          "test_msg_001",
				MessageBody: "test%20message%20body%201",
			},
			{
				Id:          "test_msg_002",
				MessageBody: "test%20message%20body%202",
			},
			{
				Id:          "test_msg_003",
				MessageBody: "test%20message%20body%203",
			},
		},
		QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "not-exist-queue1"),
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageBatchRequest)
		*v = sendMessageRequest_not_found_queue
		return true
	}

	expected := models.ErrorResult{
		Type:    "Not Found",
		Code:    "AWS.SimpleQueueService.NonExistentQueue",
		Message: "The specified queue does not exist for this wsdl version.",
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageBatchV1(r)
	errorResult := response.GetResult().(models.ErrorResult)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, expected, errorResult)
}

func TestSendMessageBatchV1_Error_NoEntry(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_no_entry := models.SendMessageBatchRequest{
		QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageBatchRequest)
		*v = sendMessageRequest_no_entry
		return true
	}

	expected := models.ErrorResult{
		Type:    "EmptyBatchRequest",
		Code:    "AWS.SimpleQueueService.EmptyBatchRequest",
		Message: "The batch request doesn't contain any entries.",
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageBatchV1(r)
	errorResult := response.GetResult().(models.ErrorResult)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, expected, errorResult)
}

func TestSendMessageBatchV1_Error_IdNotDistinct(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_no_entry := models.SendMessageBatchRequest{
		Entries: []models.SendMessageBatchRequestEntry{
			{
				Id:          "test_msg_001",
				MessageBody: "test%20message%20body%201",
			},
			{
				Id:          "test_msg_001",
				MessageBody: "test%20message%20body%202",
			},
			{
				Id:          "test_msg_001",
				MessageBody: "test%20message%20body%203",
			},
		},
		QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageBatchRequest)
		*v = sendMessageRequest_no_entry
		return true
	}

	expected := models.ErrorResult{
		Type:    "BatchEntryIdsNotDistinct",
		Code:    "AWS.SimpleQueueService.BatchEntryIdsNotDistinct",
		Message: "Two or more batch entries in the request have the same Id.",
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageBatchV1(r)
	errorResult := response.GetResult().(models.ErrorResult)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, expected, errorResult)
}

func TestSendMessageBatchV1_Error_TooManyEntries(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	sendMessageRequest_success := models.SendMessageBatchRequest{
		Entries: []models.SendMessageBatchRequestEntry{
			{
				Id:          "test_msg_001",
				MessageBody: "test%20message%20body%201",
			},
			{
				Id:          "test_msg_002",
				MessageBody: "test%20message%20body%202",
			},
			{
				Id:          "test_msg_003",
				MessageBody: "test%20message%20body%203",
			},
			{
				Id:          "test_msg_004",
				MessageBody: "test%20message%20body%204",
			},
			{
				Id:          "test_msg_005",
				MessageBody: "test%20message%20body%205",
			},
			{
				Id:          "test_msg_006",
				MessageBody: "test%20message%20body%206",
			},
			{
				Id:          "test_msg_007",
				MessageBody: "test%20message%20body%207",
			},
			{
				Id:          "test_msg_008",
				MessageBody: "test%20message%20body%208",
			},
			{
				Id:          "test_msg_009",
				MessageBody: "test%20message%20body%209",
			},
			{
				Id:          "test_msg_010",
				MessageBody: "test%20message%20body%210",
			},
			{
				Id:          "test_msg_011",
				MessageBody: "test%20message%20body%211",
			},
		},
		QueueUrl: fmt.Sprintf("%s/%s", fixtures.BASE_URL, "unit-queue1"),
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.SendMessageBatchRequest)
		*v = sendMessageRequest_success
		return true
	}

	expected := models.ErrorResult{
		Type:    "TooManyEntriesInBatchRequest",
		Code:    "AWS.SimpleQueueService.TooManyEntriesInBatchRequest",
		Message: "Maximum number of entries per request are 10.",
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := SendMessageBatchV1(r)
	errorResult := response.GetResult().(models.ErrorResult)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, expected, errorResult)

}

func TestSendMessageBatchV1_Error_transformer(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := SendMessageBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)

}
