package smoke_tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Admiral-Piett/goaws/app/conf"
	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/gavv/httpexpect/v2"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
)

func Test_Publish_sqs_json_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &message,
		Subject:  &subject,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 1)
	assert.Equal(t, message, string(messages[0].MessageBody))
}

func Test_Publish_Sqs_With_Message_Attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	snsClient := sns.NewFromConfig(sdkConfig)

	createQueueResult, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	topicName := aws.String("unit-topic2")

	createTopicResult, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: topicName,
	})

	snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              createTopicResult.TopicArn,
		Attributes:            map[string]string{},
		Endpoint:              createQueueResult.QueueUrl,
		ReturnSubscriptionArn: true,
	})
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	attributes := map[string]types.MessageAttributeValue{
		"someKey": {
			BinaryValue: []byte(message),
			DataType:    aws.String("Binary"),
		},
	}

	publishResponse, publishErr := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn:          createTopicResult.TopicArn,
		Message:           &message,
		Subject:           &subject,
		MessageAttributes: attributes,
	})

	receiveMessageResponse, receiveErr := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: createQueueResult.QueueUrl,
	})

	type Message struct {
		Message string `json:"Message"`
		Subject string `json:"Subject"`
	}

	var receiveMessage Message

	assert.Nil(t, publishErr)
	assert.NotNil(t, publishResponse)

	assert.Nil(t, receiveErr)
	assert.NotNil(t, receiveMessageResponse)

	body := *receiveMessageResponse.Messages[0].Body
	json.Unmarshal([]byte(body), &receiveMessage)
	assert.Equal(t, message, receiveMessage.Message)
	assert.Equal(t, subject, receiveMessage.Subject)
}

func Test_Publish_sqs_json_not_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	topicArn := app.SyncTopics.Topics["unit-topic3"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &message,
		Subject:  &subject,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	messages := app.SyncQueues.Queues["subscribed-queue3"].Messages
	assert.Len(t, messages, 1)

	body := string(messages[0].MessageBody)
	assert.Contains(t, body, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body, "Type")
	assert.Contains(t, body, "MessageId")
	assert.Contains(t, body, "TopicArn")
	assert.Contains(t, body, subject)
	assert.Contains(t, body, "Signature")
	assert.Contains(t, body, "SigningCertURL")
	assert.Contains(t, body, "UnsubscribeURL")
	assert.Contains(t, body, "SubscribeURL")
	assert.Contains(t, body, "MessageAttributes")
}

func Test_Publish_http_json(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")

	called := false
	httpMessage := ""
	subscribedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)
		httpMessage = buf.String()

		called = true
		w.WriteHeader(200)
	}))

	defer func() {
		server.Close()
		subscribedServer.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	app.SyncTopics.Lock()
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	app.SyncTopics.Unlock()

	topicArn := app.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	response, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &message,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_https_json_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")

	called := false
	httpMessage := ""
	subscribedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)
		httpMessage = buf.String()

		called = true
		w.WriteHeader(200)
	}))

	defer func() {
		server.Close()
		subscribedServer.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	app.SyncTopics.Lock()
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	app.SyncTopics.Unlock()

	topicArn := app.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	response, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &message,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_https_json_not_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")

	called := false
	httpMessage := ""
	subscribedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)
		httpMessage = buf.String()

		called = true
		w.WriteHeader(200)
	}))

	defer func() {
		server.Close()
		subscribedServer.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	app.SyncTopics.Lock()
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Raw = false
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	app.SyncTopics.Unlock()

	topicArn := app.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &message,
		Subject:  &subject,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	assert.True(t, called)
	assert.Contains(t, httpMessage, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, httpMessage, "Type")
	assert.Contains(t, httpMessage, "MessageId")
	assert.Contains(t, httpMessage, "TopicArn")
	assert.Contains(t, httpMessage, subject)
	assert.Contains(t, httpMessage, "Signature")
	assert.Contains(t, httpMessage, "SigningCertURL")
	assert.Contains(t, httpMessage, "UnsubscribeURL")
	assert.Contains(t, httpMessage, "SubscribeURL")
	assert.Contains(t, httpMessage, "MessageAttributes")
}

func Test_Publish_sqs_xml_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action   string `schema:"Action"`
		TopicArn string `schema:"TopicArn"`
		Message  string `schema:"Message"`
		Subject  string `schema:"Subject"`
	}{
		Action:   "Publish",
		TopicArn: topicArn,
		Message:  message,
		Subject:  subject,
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 1)
	assert.Equal(t, message, string(messages[0].MessageBody))
}

func Test_Publish_sqs_xml_not_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	topicArn := app.SyncTopics.Topics["unit-topic3"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action   string `schema:"Action"`
		TopicArn string `schema:"TopicArn"`
		Message  string `schema:"Message"`
		Subject  string `schema:"Subject"`
	}{
		Action:   "Publish",
		TopicArn: topicArn,
		Message:  message,
		Subject:  subject,
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	messages := app.SyncQueues.Queues["subscribed-queue3"].Messages
	assert.Len(t, messages, 1)

	body := string(messages[0].MessageBody)
	assert.Contains(t, body, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body, "Type")
	assert.Contains(t, body, "MessageId")
	assert.Contains(t, body, "TopicArn")
	assert.Contains(t, body, subject)
	assert.Contains(t, body, "Signature")
	assert.Contains(t, body, "SigningCertURL")
	assert.Contains(t, body, "UnsubscribeURL")
	assert.Contains(t, body, "SubscribeURL")
	assert.Contains(t, body, "MessageAttributes")
}

func Test_Publish_http_xml(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")

	called := false
	httpMessage := ""
	subscribedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)
		httpMessage = buf.String()

		called = true
		w.WriteHeader(200)
	}))

	defer func() {
		server.Close()
		subscribedServer.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	app.SyncTopics.Lock()
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	app.SyncTopics.Unlock()

	topicArn := app.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action   string `schema:"Action"`
		TopicArn string `schema:"TopicArn"`
		Message  string `schema:"Message"`
		Subject  string `schema:"Subject"`
	}{
		Action:   "Publish",
		TopicArn: topicArn,
		Message:  message,
		Subject:  subject,
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_https_xml_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")

	called := false
	httpMessage := ""
	subscribedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)
		httpMessage = buf.String()

		called = true
		w.WriteHeader(200)
	}))

	defer func() {
		server.Close()
		subscribedServer.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	app.SyncTopics.Lock()
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	app.SyncTopics.Unlock()

	topicArn := app.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action   string `schema:"Action"`
		TopicArn string `schema:"TopicArn"`
		Message  string `schema:"Message"`
		Subject  string `schema:"Subject"`
	}{
		Action:   "Publish",
		TopicArn: topicArn,
		Message:  message,
		Subject:  subject,
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_https_xml_not_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")

	called := false
	httpMessage := ""
	subscribedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		io.Copy(buf, r.Body)
		httpMessage = buf.String()

		called = true
		w.WriteHeader(200)
	}))

	defer func() {
		server.Close()
		subscribedServer.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	app.SyncTopics.Lock()
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Raw = false
	app.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	app.SyncTopics.Unlock()

	topicArn := app.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action   string `schema:"Action"`
		TopicArn string `schema:"TopicArn"`
		Message  string `schema:"Message"`
		Subject  string `schema:"Subject"`
	}{
		Action:   "Publish",
		TopicArn: topicArn,
		Message:  message,
		Subject:  subject,
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	assert.True(t, called)
	assert.Contains(t, httpMessage, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, httpMessage, "Type")
	assert.Contains(t, httpMessage, "MessageId")
	assert.Contains(t, httpMessage, "TopicArn")
	assert.Contains(t, httpMessage, subject)
	assert.Contains(t, httpMessage, "Signature")
	assert.Contains(t, httpMessage, "SigningCertURL")
	assert.Contains(t, httpMessage, "UnsubscribeURL")
	assert.Contains(t, httpMessage, "SubscribeURL")
	assert.Contains(t, httpMessage, "MessageAttributes")
}
