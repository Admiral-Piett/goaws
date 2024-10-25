package gosns

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

func Test_PublishBatchV1_success_all_entries_successful(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		return fmt.Sprintf("messageId-%d", callCount), nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"),
			PublishBatchRequestEntries: models.PublishBatchRequestEntries{
				Member: []*models.PublishBatchRequestEntry{
					{
						ID:      "1",
						Message: "message1",
					},
					{
						ID:      "2",
						Message: "message2",
					},
				},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := PublishBatchV1(r)

	response, _ := res.(models.PublishBatchResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	assert.NotEqual(t, "", response.Metadata.RequestId)

	expectedSuccessEntries := models.PublishBatchSuccessful{
		SuccessEntries: []models.PublishBatchResultEntry{
			{
				Id:        "1",
				MessageId: "messageId-1",
			},
			{
				Id:        "2",
				MessageId: "messageId-2",
			},
		},
	}

	expectedFailedEntries := models.PublishBatchFailed{
		ErrorEntries: []models.BatchResultErrorEntry{},
	}

	assert.Equal(t, expectedFailedEntries, response.Result.Failed)
	assert.Equal(t, expectedSuccessEntries, response.Result.Successful)
}

func Test_PublishBatchV1_success_some_entries_fail(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		if callCount%2 == 0 {
			return "", fmt.Errorf("ValidationError")
		}
		return fmt.Sprintf("messageId-%d", callCount), nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"),
			PublishBatchRequestEntries: models.PublishBatchRequestEntries{
				Member: []*models.PublishBatchRequestEntry{
					{
						ID:      "1",
						Message: "message1",
					},
					{
						ID:      "2",
						Message: "message2",
					},
					{
						ID:      "3",
						Message: "message3",
					},
				},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, res := PublishBatchV1(r)

	response, _ := res.(models.PublishBatchResponse)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, models.BASE_XMLNS, response.Xmlns)
	assert.NotEqual(t, "", response.Metadata.RequestId)

	expectedSuccessEntries := models.PublishBatchSuccessful{
		SuccessEntries: []models.PublishBatchResultEntry{
			{
				Id:        "1",
				MessageId: "messageId-1",
			},
			{
				Id:        "3",
				MessageId: "messageId-3",
			},
		},
	}

	e := models.SnsErrors["ValidationError"]
	expectedFailedEntries := models.PublishBatchFailed{
		ErrorEntries: []models.BatchResultErrorEntry{
			{
				Id:          "2",
				Code:        e.Code,
				Message:     e.Message,
				SenderFault: true,
			},
		},
	}

	assert.Equal(t, expectedFailedEntries, response.Result.Failed)
	assert.Equal(t, expectedSuccessEntries, response.Result.Successful)
}

func Test_PublishBatchV1_error_invalid_request(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		return "", nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, 0, callCount)
}

func Test_PublishBatchV1_error_missing_topic_arn(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		return "", nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			PublishBatchRequestEntries: models.PublishBatchRequestEntries{
				Member: []*models.PublishBatchRequestEntry{
					{
						ID:      "1",
						Message: "message1",
					},
					{
						ID:      "2",
						Message: "message2",
					},
				},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, 0, callCount)
}

func Test_PublishBatchV1_error_no_requested_messages(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		return "", nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"),
			PublishBatchRequestEntries: models.PublishBatchRequestEntries{
				Member: []*models.PublishBatchRequestEntry{},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, 0, callCount)
}

func Test_PublishBatchV1_error_too_many_requested_messages(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		return "", nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"),
			PublishBatchRequestEntries: models.PublishBatchRequestEntries{
				Member: []*models.PublishBatchRequestEntry{
					{},
					{},
					{},
					{},
					{},
					{},
					{},
					{},
					{},
					{},
					{},
				},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, 0, callCount)
}

func Test_PublishBatchV1_error_message_missing_message_id(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		return "", nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"),
			PublishBatchRequestEntries: models.PublishBatchRequestEntries{
				Member: []*models.PublishBatchRequestEntry{
					{},
					{},
				},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, 0, callCount)
}

func Test_PublishBatchV1_error_duplicate_message_id(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		return "", nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "unit-topic2"),
			PublishBatchRequestEntries: models.PublishBatchRequestEntries{
				Member: []*models.PublishBatchRequestEntry{
					{ID: "1"},
					{ID: "1"},
				},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, 0, callCount)
}

func Test_PublishBatchV1_error_missing_topic(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
		publishMessageByTopicFunc = publishMessageByTopic
	}()

	callCount := 0
	publishMessageByTopicFunc = func(topic *app.Topic, message interfaces.AbstractPublishEntry) (string, error) {
		callCount++
		return "", nil
	}

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishBatchRequest)
		*v = models.PublishBatchRequest{
			TopicArn: fmt.Sprintf("%s:%s", fixtures.BASE_SNS_ARN, "garbage"),
			PublishBatchRequestEntries: models.PublishBatchRequestEntries{
				Member: []*models.PublishBatchRequestEntry{
					{ID: "1"},
					{ID: "2"},
				},
			},
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	code, _ := PublishBatchV1(r)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, 0, callCount)
}
