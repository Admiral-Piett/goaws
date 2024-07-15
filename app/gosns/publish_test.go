package gosns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Admiral-Piett/goaws/app/fixtures"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

func TestPublishV1_success_sqs(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			TopicArn: topicArn,
			Message:  message,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := PublishV1(r)

	assert.Equal(t, http.StatusOK, status)
	_, ok := response.(models.PublishResponse)
	assert.True(t, ok)

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 1)
	assert.Equal(t, message, string(messages[0].MessageBody))
}

func TestPublishV1_success_http(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic-http"].Arn

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			TopicArn: topicArn,
			Message:  message,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := PublishV1(r)

	assert.Equal(t, http.StatusOK, status)
	_, ok := response.(models.PublishResponse)
	assert.True(t, ok)
}

func TestPublishV1_success_https(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic-http"].Arn

	app.SyncTopics.Lock()
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	app.SyncTopics.Unlock()

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			TopicArn: topicArn,
			Message:  message,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := PublishV1(r)

	assert.Equal(t, http.StatusOK, status)
	_, ok := response.(models.PublishResponse)
	assert.True(t, ok)
}

func TestPublishV1_success_with_optional_fields(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			TopicArn: topicArn,
			Message:  message,
			MessageAttributes: map[string]models.MessageAttributeValue{
				"test": models.MessageAttributeValue{
					DataType:    "string",
					StringValue: "value",
				},
			},
			MessageDeduplicationId: "dedupe-id",
			MessageGroupId:         "group-id",
			MessageStructure:       "json",
			PhoneNumber:            "phone-number",
			Subject:                "subject",
			TargetArn:              "target-arn",
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, response := PublishV1(r)

	assert.Equal(t, http.StatusOK, status)
	_, ok := response.(models.PublishResponse)
	assert.True(t, ok)

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 1)
	assert.Equal(t, message, string(messages[0].MessageBody))
}

func TestPublishV1_request_transformer_error(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		return false
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := PublishV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}

func TestPublishV1_request_missing_topic_arn(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			Message: message,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := PublishV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}

func TestPublishV1_request_missing_message(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			TopicArn: topicArn,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := PublishV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}

func TestPublishV1_request_invalid_topic(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		utils.REQUEST_TRANSFORMER = utils.TransformRequest
	}()

	message := "{\"IAm\": \"aMessage\"}"
	utils.REQUEST_TRANSFORMER = func(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
		v := resultingStruct.(*models.PublishRequest)
		*v = models.PublishRequest{
			TopicArn: "garbage",
			Message:  message,
		}
		return true
	}

	_, r := test.GenerateRequestInfo("POST", "/", nil, true)
	status, _ := PublishV1(r)

	assert.Equal(t, http.StatusBadRequest, status)
}

func Test_publishSQS_success_raw(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	message := "{\"IAm\": \"aMessage\"}"
	sub := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0]
	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}
	err := publishSQS(sub, "unit-topic1", &request)

	assert.Nil(t, err)

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 1)
	assert.Equal(t, message, string(messages[0].MessageBody))
}

// Most other scenarios should be tested in the functions above, if reasonably possible
func Test_publishSQS_success_json(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	message := "{\"IAm\": \"aMessage\"}"

	app.SyncTopics.Lock()
	sub := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0]
	sub.Raw = false
	app.SyncTopics.Unlock()
	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}
	err := publishSQS(sub, "unit-topic1", &request)

	assert.Nil(t, err)

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 1)

	body := string(messages[0].MessageBody)
	assert.Contains(t, body, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body, "Type")
	assert.Contains(t, body, "MessageId")
	assert.Contains(t, body, "TopicArn")
	assert.Contains(t, body, "Signature")
	assert.Contains(t, body, "SigningCertURL")
	assert.Contains(t, body, "UnsubscribeURL")
	assert.Contains(t, body, "SubscribeURL")
	assert.Contains(t, body, "MessageAttributes")
}

func Test_publishSQS_filter_policy_not_satisfied_by_attributes(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"

	app.SyncTopics.Lock()
	sub := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0]
	sub.FilterPolicy = &app.FilterPolicy{"foo": []string{"bar"}}
	app.SyncTopics.Unlock()

	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
		MessageAttributes: map[string]models.MessageAttributeValue{
			"invalid": models.MessageAttributeValue{
				DataType:    "string",
				StringValue: "garbage",
			},
		},
	}
	err := publishSQS(sub, "unit-topic1", &request)

	assert.Nil(t, err)
}

func Test_publishSQS_missing_queue_returns_nil(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn

	message := "{\"IAm\": \"aMessage\"}"

	app.SyncTopics.Lock()
	sub := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0]
	sub.EndPoint = "garbage"
	app.SyncTopics.Unlock()

	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}
	err := publishSQS(sub, "unit-topic1", &request)

	assert.Nil(t, err)
}

// Most other scenarios should be tested in the functions above, if reasonably possible
func Test_publishHTTP_success(t *testing.T) {
	called := false
	subscribedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
	}))

	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
		subscribedServer.Close()
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"

	app.SyncTopics.Lock()
	sub := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0]
	sub.EndPoint = subscribedServer.URL
	app.SyncTopics.Unlock()

	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}

	publishHTTP(sub, &request)

	assert.True(t, called)
}

func Test_publishHTTP_callEndpoint_failure(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
	}()

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"

	app.SyncTopics.Lock()
	sub := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0]
	app.SyncTopics.Unlock()

	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}

	publishHTTP(sub, &request)
	// swallows all errors
}

func Test_createMessageBody_success_json(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
	}()

	message := "{\"default\": \"message\"}"
	subject := "I'm the subject"
	attrs := map[string]app.MessageAttributeValue{
		"test": app.MessageAttributeValue{
			Name:     "MyAttr",
			DataType: "string",
			Value:    "value",
		},
	}

	sub := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0]

	result, err := createMessageBody(sub, message, subject, "json", attrs)

	assert.Nil(t, err)

	msg := &app.SNSMessage{}
	json.Unmarshal(result, msg)

	assert.Equal(t, "Notification", msg.Type)
	assert.Equal(t, "", msg.Token)
	assert.Equal(t, fmt.Sprintf("%s:unit-topic1", fixtures.BASE_SNS_ARN), msg.TopicArn)
	assert.Equal(t, "I'm the subject", msg.Subject)
	assert.Equal(t, "message", msg.Message)
	assert.Equal(t, "1", msg.SignatureVersion)
	assert.Contains(t, msg.SigningCertURL, "http://host:port/SimpleNotificationService/")
	assert.Contains(t, msg.UnsubscribeURL, "http://host:port/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:region:accountID:unit-topic1:")
	assert.Equal(t, msg.MessageAttributes, map[string]app.MsgAttr{"test": app.MsgAttr{Type: "string", Value: "value"}})
}

func Test_createMessageBody_success_raw(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		test.ResetApp()
	}()

	message := "{\"default\": \"message\"}"
	subject := "I'm the subject"
	attrs := map[string]app.MessageAttributeValue{
		"test": app.MessageAttributeValue{
			Name:     "MyAttr",
			DataType: "string",
			Value:    "value",
		},
	}

	sub := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0]

	result, err := createMessageBody(sub, message, subject, "not-json", attrs)

	assert.Nil(t, err)

	msg := &app.SNSMessage{}
	json.Unmarshal(result, msg)

	assert.Equal(t, "Notification", msg.Type)
	assert.Equal(t, "", msg.Token)
	assert.Equal(t, fmt.Sprintf("%s:unit-topic1", fixtures.BASE_SNS_ARN), msg.TopicArn)
	assert.Equal(t, "I'm the subject", msg.Subject)
	assert.Equal(t, message, msg.Message)
	assert.Equal(t, "1", msg.SignatureVersion)
	assert.Contains(t, msg.SigningCertURL, "http://host:port/SimpleNotificationService/")
	assert.Contains(t, msg.UnsubscribeURL, "http://host:port/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:region:accountID:unit-topic1:")
	assert.Equal(t, msg.MessageAttributes, map[string]app.MsgAttr{"test": app.MsgAttr{Type: "string", Value: "value"}})
}

func Test_formatAttributes_success(t *testing.T) {
	attrs := map[string]app.MessageAttributeValue{
		"test1": app.MessageAttributeValue{
			Name:     "MyAttr",
			DataType: "string",
			Value:    "value1",
		},
		"test2": app.MessageAttributeValue{
			Name:     "MyAttr",
			DataType: "string",
			Value:    "value2",
		},
	}
	expected := map[string]app.MsgAttr{
		"test1": app.MsgAttr{Type: "string", Value: "value1"},
		"test2": app.MsgAttr{Type: "string", Value: "value2"},
	}

	result := formatAttributes(attrs)

	assert.Equal(t, expected, result)
}
