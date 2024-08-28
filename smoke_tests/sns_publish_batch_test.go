package smoke_tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Publish_batch_sqs_json_raw(t *testing.T) {
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
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Message: &message,
				Subject: &subject,
			},
			{
				Message: &message,
				Subject: &subject,
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	require.Len(t, messages, 2)
	assert.Equal(t, message, string(messages[0].MessageBody))
	assert.Equal(t, message, string(messages[1].MessageBody))
}

func Test_Publish_batch_sqs_json_not_raw(t *testing.T) {
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
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Message: &message,
				Subject: &subject,
			},
			{
				Message: &message,
				Subject: &subject,
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	messages := app.SyncQueues.Queues["subscribed-queue3"].Messages
	assert.Len(t, messages, 2)

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

func Test_Publish_batch_http_json(t *testing.T) {
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
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,

		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Message: &message,
			},
			{
				Message: &message,
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_batch_https_json_raw(t *testing.T) {
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
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,

		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Message: &message,
			},
			{
				Message: &message,
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_batch_https_json_not_raw(t *testing.T) {
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
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Message: &message,
				Subject: &subject,
			},
			{
				Message: &message,
				Subject: &subject,
			},
		},
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

func Test_Publish_batch_sqs_xml_raw(t *testing.T) {
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
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		}{
			{
				Message: message,
				Subject: subject,
			},
			{
				Message: message,
				Subject: subject,
			},
		},
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	messages := app.SyncQueues.Queues["subscribed-queue1"].Messages
	assert.Len(t, messages, 2)
	assert.Equal(t, message, string(messages[0].MessageBody))
	assert.Equal(t, message, string(messages[1].MessageBody))
}

func Test_Publish_batch_sqs_xml_not_raw(t *testing.T) {
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
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		}{
			{
				Message: message,
				Subject: subject,
			},
			{
				Message: message,
				Subject: subject,
			},
		},
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	messages := app.SyncQueues.Queues["subscribed-queue3"].Messages
	assert.Len(t, messages, 2)

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

func Test_Publish_batch_http_xml(t *testing.T) {
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
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		}{
			{
				Message: message,
				Subject: subject,
			},
			{
				Message: message,
				Subject: subject,
			},
		},
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_batch_https_xml_raw(t *testing.T) {
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
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		}{
			{
				Message: message,
				Subject: subject,
			},
			{
				Message: message,
				Subject: subject,
			},
		},
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_batch_https_xml_not_raw(t *testing.T) {
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
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: []struct {
			Message string `schema:"Message"`
			Subject string `schema:"Subject"`
		}{
			{
				Message: message,
				Subject: subject,
			},
			{
				Message: message,
				Subject: subject,
			},
		},
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
