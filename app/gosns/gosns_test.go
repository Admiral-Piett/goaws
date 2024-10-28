package gosns

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/stretchr/testify/assert"
)

const (
	messageStructureJSON = "json"
)

func Test_publishSQS_success_raw_true(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
	}()

	message := "{\"IAm\": \"aMessage\"}"
	topic := models.SyncTopics.Topics["unit-topic1"]
	sub := topic.Subscriptions[0]
	topicArn := topic.Arn
	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}
	err := publishSQS(sub, topic, &request)

	assert.Nil(t, err)

	messages := models.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 1)
	assert.Equal(t, message, string(messages[0].MessageBody))
}

func Test_publishSQS_success_json_raw_false(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
	}()

	message := "{\"IAm\": \"aMessage\"}"

	models.SyncTopics.Lock()
	topic := models.SyncTopics.Topics["unit-topic1"]
	sub := topic.Subscriptions[0]
	topicArn := topic.Arn
	sub.Raw = false
	models.SyncTopics.Unlock()
	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}
	err := publishSQS(sub, topic, &request)

	assert.Nil(t, err)

	messages := models.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 1)

	body := string(messages[0].MessageBody)
	assert.Contains(t, body, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body, "Type")
	assert.Contains(t, body, "MessageId")
	assert.Contains(t, body, "TopicArn")
	assert.Contains(t, body, "Signature")
	assert.Contains(t, body, "SigningCertURL")
	assert.Contains(t, body, "UnsubscribeURL")
}

func Test_publishSQS_filter_policy_not_satisfied_by_attributes(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
	}()

	topicArn := models.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"

	models.SyncTopics.Lock()
	topic := models.SyncTopics.Topics["unit-topic1"]
	sub := topic.Subscriptions[0]
	sub.FilterPolicy = &models.FilterPolicy{"foo": []string{"bar"}}
	models.SyncTopics.Unlock()

	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
		MessageAttributes: map[string]models.MessageAttributeValue{
			"invalid": models.MessageAttributeValue{
				DataType:    "String",
				StringValue: "garbage",
			},
		},
	}
	err := publishSQS(sub, topic, &request)

	assert.Nil(t, err)
}

func Test_publishSQS_missing_queue_returns_nil(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
	}()

	message := "{\"IAm\": \"aMessage\"}"

	models.SyncTopics.Lock()
	topic := models.SyncTopics.Topics["unit-topic1"]
	sub := topic.Subscriptions[0]
	topicArn := topic.Arn
	sub.EndPoint = "garbage"
	models.SyncTopics.Unlock()

	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}
	err := publishSQS(sub, topic, &request)

	assert.Nil(t, err)
}

func Test_publishHTTP_success(t *testing.T) {
	called := false
	subscribedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
	}))

	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
		subscribedServer.Close()
	}()

	message := "{\"IAm\": \"aMessage\"}"

	models.SyncTopics.Lock()
	topic := models.SyncTopics.Topics["unit-topic1"]
	sub := topic.Subscriptions[0]
	topicArn := topic.Arn
	sub.EndPoint = subscribedServer.URL
	models.SyncTopics.Unlock()

	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}

	publishHTTP(sub, topicArn, &request)

	assert.True(t, called)
}

func Test_publishHTTP_callEndpoint_failure(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
	}()

	message := "{\"IAm\": \"aMessage\"}"

	models.SyncTopics.Lock()
	topic := models.SyncTopics.Topics["unit-topic1"]
	sub := topic.Subscriptions[0]
	topicArn := topic.Arn
	models.SyncTopics.Unlock()

	request := models.PublishRequest{
		TopicArn: topicArn,
		Message:  message,
	}

	publishHTTP(sub, topicArn, &request)
	// swallows all errors
}

func TestCreateMessageBody_success_NoMessageAttributes(t *testing.T) {
	message := "message text"
	subject := "subject"
	subs := &models.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	msg := &models.PublishRequest{
		Message: message,
		Subject: subject,
	}

	result, err := createMessageBody(subs, msg, map[string]models.SqsMessageAttributeValue{})
	assert.Nil(t, err)

	unmarshalledMessage := &models.SNSMessage{}
	json.Unmarshal(result, unmarshalledMessage)

	assert.Equal(t, "Notification", unmarshalledMessage.Type)
	assert.Equal(t, "", unmarshalledMessage.Token)
	assert.Equal(t, "topic-arn", unmarshalledMessage.TopicArn)
	assert.Equal(t, subject, unmarshalledMessage.Subject)
	assert.Equal(t, message, unmarshalledMessage.Message)
	assert.Equal(t, "1", unmarshalledMessage.SignatureVersion)
	assert.Contains(t, unmarshalledMessage.SigningCertURL, "/SimpleNotificationService/")
	assert.Contains(t, unmarshalledMessage.UnsubscribeURL, "/?Action=Unsubscribe&SubscriptionArn=subs-arn")
	assert.Nil(t, unmarshalledMessage.MessageAttributes)
}

func TestCreateMessageBody_success_WithMessageAttributes(t *testing.T) {
	message := "message text"
	subject := "subject"
	subs := &models.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	attributes := map[string]models.SqsMessageAttributeValue{
		"test": {
			DataType: "String",
			ValueKey: "StringValue",
			Value:    "test",
		},
	}

	msg := &models.PublishRequest{
		Message: message,
		Subject: subject,
	}
	snsMessage, err := createMessageBody(subs, msg, attributes)

	assert.Nil(t, err)
	assert.Contains(t, string(snsMessage), "\"MessageAttributes\":{\"test\":{\"DataType\":\"String\",\"StringValue\":\"test\"}}")
}

func TestCreateMessageBody_JSONMessageStructure_UsesDefaultMessageIfNoMatchingProtocolKeyProvided(t *testing.T) {
	subs := &models.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	message := `{"default": "default message text", "http": "HTTP message text"}`
	subject := "subject"
	msg := &models.PublishRequest{
		Message:          message,
		Subject:          subject,
		MessageStructure: messageStructureJSON,
	}

	snsMessage, err := createMessageBody(subs, msg, nil)
	assert.Nil(t, err)
	assert.Contains(t, string(snsMessage), "\"Message\":\"default message text\"")
}

func TestCreateMessageBody_JSONMessageStructure_MissingDefaultKey(t *testing.T) {
	subs := &models.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	message := `{"sqs": "message text"}`
	subject := "subject"
	msg := &models.PublishRequest{
		Message:          message,
		Subject:          subject,
		MessageStructure: messageStructureJSON,
	}

	snsMessage, err := createMessageBody(subs, msg, nil)

	assert.Error(t, err)
	assert.Nil(t, snsMessage)
}

func TestCreateMessageBody_JSONMessageStructure_SelectsProtocolSpecificMessageIfAvailable(t *testing.T) {
	subs := &models.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	message := `{"default": "default message text", "sqs": "sqs message text"}`
	subject := "subject"
	msg := &models.PublishRequest{
		Message:          message,
		Subject:          subject,
		MessageStructure: messageStructureJSON,
	}

	snsMessage, err := createMessageBody(subs, msg, nil)

	assert.Nil(t, err)
	assert.Contains(t, string(snsMessage), "\"Message\":\"sqs message text\"")
}

func TestCreateMessageBody_NonJsonMessageStructure_MessageContainingJson(t *testing.T) {
	subs := &models.Subscription{
		Protocol:        "sns",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	message := `{"default": "default message text", "sqs": "sqs message text"}`
	subject := "subject"
	msg := &models.PublishRequest{
		Message: message,
		Subject: subject,
	}

	snsMessage, err := createMessageBody(subs, msg, nil)
	assert.Nil(t, err)
	assert.Contains(t, string(snsMessage), "\"Message\":\"{\\\"default\\\": \\\"default message text\\\", \\\"sqs\\\": \\\"sqs message text\\\"}\"")
}

func Test_formatAttributes_success(t *testing.T) {
	attrs := map[string]models.SqsMessageAttributeValue{
		"test1": models.SqsMessageAttributeValue{
			Name:     "MyAttr",
			DataType: "String",
			Value:    "value1",
		},
		"test2": models.SqsMessageAttributeValue{
			Name:     "MyAttr",
			DataType: "String",
			Value:    "value2",
		},
	}
	expected := map[string]models.MessageAttributeValue{
		"test1": {DataType: "String", StringValue: "value1"},
		"test2": {DataType: "String", StringValue: "value2"},
	}

	result := formatAttributes(attrs)

	assert.Equal(t, expected, result)
}

func Test_publishMessageByTopic_sqs_success(t *testing.T) {
	defer func() {
		publishSqsMessageFunc = publishSQS
	}()

	calledWith := [][]interface{}{}
	publishSqsMessageFunc = func(subscription *models.Subscription, topic *models.Topic, entry interfaces.AbstractPublishEntry) error {
		calledWith = append(calledWith, []interface{}{subscription, topic, entry})
		return nil
	}
	subscription := &models.Subscription{Protocol: "sqs"}
	topic := &models.Topic{
		Subscriptions: []*models.Subscription{subscription},
	}
	entry := &models.PublishBatchRequestEntry{}

	msgId, err := publishMessageByTopic(topic, entry)

	assert.NotEqual(t, "", msgId)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(calledWith))
	assert.Equal(t, []interface{}{subscription, topic, entry}, calledWith[0])
}

func Test_publishMessageByTopic_http_success(t *testing.T) {
	defer func() {
		publishHttpMessageFunc = publishHTTP
	}()

	topicArn := "my-topic-arn"

	calledWith := [][]interface{}{}
	publishHttpMessageFunc = func(subscription *models.Subscription, topicArn string, entry interfaces.AbstractPublishEntry) {
		calledWith = append(calledWith, []interface{}{subscription, topicArn, entry})
	}
	subscription := &models.Subscription{Protocol: "http"}
	topic := &models.Topic{
		Arn:           topicArn,
		Subscriptions: []*models.Subscription{subscription},
	}
	entry := &models.PublishBatchRequestEntry{}

	msgId, err := publishMessageByTopic(topic, entry)

	assert.NotEqual(t, "", msgId)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(calledWith))
	assert.Equal(t, []interface{}{subscription, topicArn, entry}, calledWith[0])
}

func Test_publishMessageByTopic_https_success(t *testing.T) {
	defer func() {
		publishHttpMessageFunc = publishHTTP
	}()

	topicArn := "my-topic-arn"

	calledWith := [][]interface{}{}
	publishHttpMessageFunc = func(subscription *models.Subscription, topicArn string, entry interfaces.AbstractPublishEntry) {
		calledWith = append(calledWith, []interface{}{subscription, topicArn, entry})
	}
	subscription := &models.Subscription{Protocol: "https"}
	topic := &models.Topic{
		Arn:           topicArn,
		Subscriptions: []*models.Subscription{subscription},
	}
	entry := &models.PublishBatchRequestEntry{}

	msgId, err := publishMessageByTopic(topic, entry)

	assert.NotEqual(t, "", msgId)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(calledWith))
	assert.Equal(t, []interface{}{subscription, topicArn, entry}, calledWith[0])
}

func Test_publishMessageByTopic_success_no_subscriptions(t *testing.T) {
	defer func() {
		publishSqsMessageFunc = publishSQS
	}()

	called := false
	publishSqsMessageFunc = func(subscription *models.Subscription, topic *models.Topic, entry interfaces.AbstractPublishEntry) error {
		called = true
		return nil
	}
	topic := &models.Topic{}
	entry := &models.PublishBatchRequestEntry{}

	msgId, err := publishMessageByTopic(topic, entry)

	assert.NotEqual(t, "", msgId)
	assert.Nil(t, err)

	assert.False(t, called)
}

func Test_publishMessageByTopic_failure_sqs_publish_failure(t *testing.T) {
	defer func() {
		publishSqsMessageFunc = publishSQS
	}()

	called := false
	publishSqsMessageFunc = func(subscription *models.Subscription, topic *models.Topic, entry interfaces.AbstractPublishEntry) error {
		called = true
		return fmt.Errorf("boom")
	}
	subscription := &models.Subscription{Protocol: "sqs"}
	topic := &models.Topic{
		Subscriptions: []*models.Subscription{subscription},
	}
	entry := &models.PublishBatchRequestEntry{}

	msgId, err := publishMessageByTopic(topic, entry)

	assert.NotEqual(t, "", msgId)
	assert.Error(t, err)

	assert.True(t, called)
}
