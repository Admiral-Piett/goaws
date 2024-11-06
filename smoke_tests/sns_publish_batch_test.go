package smoke_tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/gavv/httpexpect/v2"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_Publish_batch_sqs_json_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	topicArn := models.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Failed, 0)
	assert.Len(t, response.Successful, 2)

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &models.SyncQueues.Queues["subscribed-queue1"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Equal(t, 2, len(receivedMessage.Messages))
	assert.Nil(t, err)

	assert.Equal(t, message, *receivedMessage.Messages[0].Body)
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", *receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[0].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[0].MessageId)
	assert.NotNil(t, receivedMessage.Messages[0].ReceiptHandle)

	assert.Equal(t, message, *receivedMessage.Messages[1].Body)
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[1].MD5OfBody)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", *receivedMessage.Messages[1].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[1].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[1].MessageId)
	assert.NotNil(t, receivedMessage.Messages[1].ReceiptHandle)
}

func Test_Publish_batch_sqs_failure_json_duplicate_ids(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	topicArn := models.SyncTopics.Topics["unit-topic1"].Arn
	subject := "I am a subject"
	dupeId := uuid.NewString()
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      aws.String(dupeId),
				Message: aws.String("{\"IAm\": \"theFirstMessage\"}"),
				Subject: aws.String(subject),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String("{\"IAm\": \"theSecondMessage\"}"),
				Subject: aws.String(subject),
			},
			{
				Id:      aws.String(dupeId),
				Message: aws.String("{\"IAm\": \"theThirdMessage\"}"),
				Subject: aws.String(subject),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String("{\"IAm\": \"theForthMessage\"}"),
				Subject: aws.String(subject),
			},
		},
	})

	assert.Error(t, err)
	assert.Nil(t, response)

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &models.SyncQueues.Queues["subscribed-queue1"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Equal(t, 0, len(receivedMessage.Messages))
}

func Test_Publish_batch_sqs_json_not_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	topicArn := models.SyncTopics.Topics["unit-topic3"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Failed, 0)
	assert.Len(t, response.Successful, 2)

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &models.SyncQueues.Queues["subscribed-queue3"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Equal(t, 2, len(receivedMessage.Messages))
	assert.Nil(t, err)

	assert.NotEqual(t, "", *receivedMessage.Messages[0].MD5OfBody)
	assert.Nil(t, receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[0].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[0].MessageId)
	assert.NotNil(t, receivedMessage.Messages[0].ReceiptHandle)

	assert.NotEqual(t, "", *receivedMessage.Messages[0].MD5OfBody)
	assert.Nil(t, receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[1].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[1].MessageId)
	assert.NotNil(t, receivedMessage.Messages[1].ReceiptHandle)

	body0 := *receivedMessage.Messages[0].Body
	assert.Contains(t, body0, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body0, "Type")
	assert.Contains(t, body0, "MessageId")
	assert.Contains(t, body0, "TopicArn")
	assert.Contains(t, body0, subject)
	assert.Contains(t, body0, "Signature")
	assert.Contains(t, body0, "SigningCertURL")
	assert.Contains(t, body0, "UnsubscribeURL")

	body1 := *receivedMessage.Messages[1].Body
	assert.Contains(t, body1, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body1, "Type")
	assert.Contains(t, body1, "MessageId")
	assert.Contains(t, body1, "TopicArn")
	assert.Contains(t, body1, subject)
	assert.Contains(t, body1, "Signature")
	assert.Contains(t, body1, "SigningCertURL")
	assert.Contains(t, body1, "UnsubscribeURL")
}

func Test_Publish_batch_sqs_json_raw_with_optional_fields(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	topicArn := models.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
				MessageAttributes: map[string]types.MessageAttributeValue{
					"test": {
						DataType:    aws.String("String"),
						StringValue: aws.String("string-value"),
					},
				},
				MessageDeduplicationId: aws.String("dedupe-id"),
				MessageGroupId:         aws.String("group-id"),
				MessageStructure:       aws.String("json"),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Failed, 0)
	assert.Len(t, response.Successful, 2)

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &models.SyncQueues.Queues["subscribed-queue1"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Len(t, receivedMessage.Messages, 2)

	assert.Equal(t, message, *receivedMessage.Messages[0].Body)
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "45a48b32ccd821cc81a8c28fbac4cd97", *receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Equal(t, *receivedMessage.Messages[0].MessageAttributes["test"].DataType, "String")
	assert.Equal(t, *receivedMessage.Messages[0].MessageAttributes["test"].StringValue, "string-value")
	assert.NotNil(t, receivedMessage.Messages[0].MessageId)
	assert.NotNil(t, receivedMessage.Messages[0].ReceiptHandle)

	assert.Equal(t, message, *receivedMessage.Messages[1].Body)
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[1].MD5OfBody)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", *receivedMessage.Messages[1].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[1].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[1].MessageId)
	assert.NotNil(t, receivedMessage.Messages[1].ReceiptHandle)
}

func Test_Publish_batch_sqs_json_not_raw_with_optional_fields(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	topicArn := models.SyncTopics.Topics["unit-topic3"].Arn
	message := "{\"default\": \"json-structured-message\", \"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
				MessageAttributes: map[string]types.MessageAttributeValue{
					"test": {
						DataType:    aws.String("String"),
						StringValue: aws.String("string-value"),
					},
				},
				MessageDeduplicationId: aws.String("dedupe-id"),
				MessageGroupId:         aws.String("group-id"),
				MessageStructure:       aws.String("json"),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Failed, 0)
	assert.Len(t, response.Successful, 2)

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &models.SyncQueues.Queues["subscribed-queue3"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Len(t, receivedMessage.Messages, 2)

	assert.NotEqual(t, "", *receivedMessage.Messages[0].MD5OfBody)
	assert.Nil(t, receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[0].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[0].MessageId)
	assert.NotNil(t, receivedMessage.Messages[0].ReceiptHandle)

	assert.NotEqual(t, "", *receivedMessage.Messages[1].MD5OfBody)
	assert.Nil(t, receivedMessage.Messages[1].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[1].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[1].MessageId)
	assert.NotNil(t, receivedMessage.Messages[1].ReceiptHandle)

	body0 := *receivedMessage.Messages[0].Body
	assert.Contains(t, body0, "\"Message\":\"json-structured-message\"")
	assert.Contains(t, body0, "Type")
	assert.Contains(t, body0, "MessageId")
	assert.Contains(t, body0, "TopicArn")
	assert.Contains(t, body0, subject)
	assert.Contains(t, body0, "Signature")
	assert.Contains(t, body0, "SigningCertURL")
	assert.Contains(t, body0, "UnsubscribeURL")

	body1 := *receivedMessage.Messages[1].Body
	assert.Contains(t, body1, "\"Message\":\"{\\\"default\\\": \\\"json-structured-message\\\", \\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body1, "Type")
	assert.Contains(t, body1, "MessageId")
	assert.Contains(t, body1, "TopicArn")
	assert.Contains(t, body1, subject)
	assert.Contains(t, body1, "Signature")
	assert.Contains(t, body1, "SigningCertURL")
	assert.Contains(t, body1, "UnsubscribeURL")
}

func Test_Publish_batch_http_json(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
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
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	models.SyncTopics.Lock()
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	models.SyncTopics.Unlock()

	topicArn := models.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,

		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Failed, 0)
	assert.Len(t, response.Successful, 2)

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_batch_https_json_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
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
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	models.SyncTopics.Lock()
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	models.SyncTopics.Unlock()

	topicArn := models.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,

		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Failed, 0)
	assert.Len(t, response.Successful, 2)

	assert.True(t, called)
	assert.Equal(t, "\"{\\\"IAm\\\": \\\"aMessage\\\"}\"", httpMessage)
}

func Test_Publish_batch_https_json_not_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
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
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	models.SyncTopics.Lock()
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Raw = false
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	models.SyncTopics.Unlock()

	topicArn := models.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.PublishBatch(context.TODO(), &sns.PublishBatchInput{
		TopicArn: &topicArn,
		PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
			},
			{
				Id:      aws.String(uuid.NewString()),
				Message: aws.String(message),
				Subject: aws.String(subject),
			},
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Failed, 0)
	assert.Len(t, response.Successful, 2)

	assert.True(t, called)
	assert.Contains(t, httpMessage, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, httpMessage, "Type")
	assert.Contains(t, httpMessage, "MessageId")
	assert.Contains(t, httpMessage, "TopicArn")
	assert.Contains(t, httpMessage, subject)
	assert.Contains(t, httpMessage, "Signature")
	assert.Contains(t, httpMessage, "SigningCertURL")
	assert.Contains(t, httpMessage, "UnsubscribeURL")
}

func Test_Publish_batch_sqs_xml_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	e := httpexpect.Default(t, server.URL)

	topicArn := models.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		}{
			Member: []struct {
				Id      string "schema:\"Id\""
				Message string "schema:\"Message\""
				Subject string "schema:\"Subject\""
			}{
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
			},
		},
	}
	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &models.SyncQueues.Queues["subscribed-queue1"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Equal(t, 2, len(receivedMessage.Messages))
	assert.Nil(t, err)

	assert.Equal(t, message, *receivedMessage.Messages[0].Body)
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", *receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[0].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[0].MessageId)
	assert.NotNil(t, receivedMessage.Messages[0].ReceiptHandle)

	assert.Equal(t, message, *receivedMessage.Messages[1].Body)
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[1].MD5OfBody)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", *receivedMessage.Messages[1].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[1].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[1].MessageId)
	assert.NotNil(t, receivedMessage.Messages[1].ReceiptHandle)
}

func Test_Publish_batch_sqs_xml_not_raw(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	e := httpexpect.Default(t, server.URL)

	topicArn := models.SyncTopics.Topics["unit-topic3"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		}{
			Member: []struct {
				Id      string "schema:\"Id\""
				Message string "schema:\"Message\""
				Subject string "schema:\"Subject\""
			}{
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
			},
		},
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &models.SyncQueues.Queues["subscribed-queue3"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Equal(t, 2, len(receivedMessage.Messages))
	assert.Nil(t, err)

	assert.NotEqual(t, "", *receivedMessage.Messages[0].MD5OfBody)
	assert.Nil(t, receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[0].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[0].MessageId)
	assert.NotNil(t, receivedMessage.Messages[0].ReceiptHandle)

	assert.NotEqual(t, "", *receivedMessage.Messages[1].MD5OfBody)
	assert.Nil(t, receivedMessage.Messages[1].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[1].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[1].MessageId)
	assert.NotNil(t, receivedMessage.Messages[1].ReceiptHandle)

	body0 := *receivedMessage.Messages[0].Body
	assert.Contains(t, body0, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body0, "Type")
	assert.Contains(t, body0, "MessageId")
	assert.Contains(t, body0, "TopicArn")
	assert.Contains(t, body0, subject)
	assert.Contains(t, body0, "Signature")
	assert.Contains(t, body0, "SigningCertURL")
	assert.Contains(t, body0, "UnsubscribeURL")

	body1 := *receivedMessage.Messages[1].Body
	assert.Contains(t, body1, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body1, "Type")
	assert.Contains(t, body1, "MessageId")
	assert.Contains(t, body1, "TopicArn")
	assert.Contains(t, body1, subject)
	assert.Contains(t, body1, "Signature")
	assert.Contains(t, body1, "SigningCertURL")
	assert.Contains(t, body1, "UnsubscribeURL")
}

func Test_Publish_batch_http_xml(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
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
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	models.SyncTopics.Lock()
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	models.SyncTopics.Unlock()

	topicArn := models.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		}{
			Member: []struct {
				Id      string "schema:\"Id\""
				Message string "schema:\"Message\""
				Subject string "schema:\"Subject\""
			}{
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
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
	defaultEnv := models.CurrentEnvironment
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
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	models.SyncTopics.Lock()
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	models.SyncTopics.Unlock()

	topicArn := models.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		}{
			Member: []struct {
				Id      string "schema:\"Id\""
				Message string "schema:\"Message\""
				Subject string "schema:\"Subject\""
			}{
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
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
	defaultEnv := models.CurrentEnvironment
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
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	models.SyncTopics.Lock()
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Protocol = "https"
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].Raw = false
	models.SyncTopics.Topics["unit-topic-http"].Subscriptions[0].EndPoint = subscribedServer.URL
	models.SyncTopics.Unlock()

	topicArn := models.SyncTopics.Topics["unit-topic-http"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"

	requestBody := struct {
		Action                     string `schema:"Action"`
		TopicArn                   string `schema:"TopicArn"`
		PublishBatchRequestEntries struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		} `schema:"PublishBatchRequestEntries"`
	}{
		Action:   "PublishBatch",
		TopicArn: topicArn,
		PublishBatchRequestEntries: struct {
			Member []struct {
				Id      string `schema:"Id"`
				Message string `schema:"Message"`
				Subject string `schema:"Subject"`
			} `schema:"member"`
		}{
			Member: []struct {
				Id      string "schema:\"Id\""
				Message string "schema:\"Message\""
				Subject string "schema:\"Subject\""
			}{
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
				{
					Id:      uuid.NewString(),
					Message: message,
					Subject: subject,
				},
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
}
