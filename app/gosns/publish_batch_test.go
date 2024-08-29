package gosns

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishBatchV1_request_transformer_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestBatchPublishV1_empty_topicArn_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		message := "{\"IAm\": \"aMessage\"}"
		e := &models.PublishBatchRequestEntry{ID: "1", Message: message}
		*v = models.PublishBatchRequest{
			TopicArn: "",
			PublishBatchRequestEntries: struct {
				Member []*models.PublishBatchRequestEntry `json:"member" schema:"member"`
			}{Member: []*models.PublishBatchRequestEntry{e}},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestBatchPublishV1_nonexistent_topicArn_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		message := "{\"IAm\": \"aMessage\"}"
		e := &models.PublishBatchRequestEntry{ID: "1", Message: message}
		*v = models.PublishBatchRequest{
			TopicArn: "non-existing",
			PublishBatchRequestEntries: struct {
				Member []*models.PublishBatchRequestEntry `json:"member" schema:"member"`
			}{Member: []*models.PublishBatchRequestEntry{e}},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestBatchPublishV1_zero_entries_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: topicArn,
			PublishBatchRequestEntries: struct {
				Member []*models.PublishBatchRequestEntry `json:"member" schema:"member"`
			}{Member: []*models.PublishBatchRequestEntry{}},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestBatchPublishV1_too_many_entries_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		entries := make([]*models.PublishBatchRequestEntry, 0, 11)
		for i := 0; i < 11; i++ {
			message := fmt.Sprintf("{\"IAm\": \"aMessage-%d\"}", i)
			e := &models.PublishBatchRequestEntry{ID: fmt.Sprintf("%d", i), Message: message}
			entries = append(entries, e)
		}
		*v = models.PublishBatchRequest{
			TopicArn: topicArn,
			PublishBatchRequestEntries: struct {
				Member []*models.PublishBatchRequestEntry `json:"member" schema:"member"`
			}{Member: entries},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestBatchPublishV1_success_sqs(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		e := &models.PublishBatchRequestEntry{ID: "1", Message: message}
		*v = models.PublishBatchRequest{
			TopicArn: topicArn,
			PublishBatchRequestEntries: struct {
				Member []*models.PublishBatchRequestEntry `json:"member" schema:"member"`
			}{Member: []*models.PublishBatchRequestEntry{e}},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := PublishBatchV1(r)

	require.Equal(t, http.StatusOK, code)
	_, ok := response.(models.PublishBatchResponse)
	assert.True(t, ok)

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	require.Len(t, messages, 1)
	assert.Equal(t, message, string(messages[0].MessageBody))
}

func TestBatchPublishV1_duplicate_id_errors(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		e := &models.PublishBatchRequestEntry{ID: "1", Message: message}
		*v = models.PublishBatchRequest{
			TopicArn: topicArn,
			PublishBatchRequestEntries: struct {
				Member []*models.PublishBatchRequestEntry `json:"member" schema:"member"`
			}{Member: []*models.PublishBatchRequestEntry{e, e}},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestBatchPublishV1_multiple_success_sqs(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	entries := make([]*models.PublishBatchRequestEntry, 0, 10)
	expectedMessages := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		message := fmt.Sprintf("{\"IAm\": \"aMessage-%d\"}", i)
		e := &models.PublishBatchRequestEntry{ID: fmt.Sprintf("%d", i), Message: message}
		entries = append(entries, e)
		expectedMessages = append(expectedMessages, message)
	}
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: topicArn,
			PublishBatchRequestEntries: struct {
				Member []*models.PublishBatchRequestEntry `json:"member" schema:"member"`
			}{Member: entries},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, response := PublishBatchV1(r)

	require.Equal(t, http.StatusOK, code)
	_, ok := response.(models.PublishBatchResponse)
	assert.True(t, ok)

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	gotMessages := []string{}
	for _, m := range messages {
		gotMessages = append(gotMessages, string(m.MessageBody))
	}
	assert.ElementsMatch(t, expectedMessages, gotMessages)
}
