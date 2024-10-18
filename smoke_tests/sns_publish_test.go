package smoke_tests

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/aws/aws-sdk-go-v2/service/sns/types"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Admiral-Piett/goaws/app/conf"
	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/gavv/httpexpect/v2"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
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
	sqsClient := sqs.NewFromConfig(sdkConfig)

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

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: &app.SyncQueues.Queues["subscribed-queue1"].URL,
	})

	assert.Len(t, receivedMessage.Messages, 1)
	assert.Equal(t, 0, len(receivedMessage.Messages[0].MessageAttributes))
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", *receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Equal(t, message, *receivedMessage.Messages[0].Body)
}

func Test_Publish_Sqs_With_Message_Attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()
	addBytesToHash := func(hasher hash.Hash, arr []byte) {
		bs := make([]byte, 4)
		binary.BigEndian.PutUint32(bs, uint32(len(arr)))
		hasher.Write(bs)
		hasher.Write(arr)
	}

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

	subscribeResult, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              createTopicResult.TopicArn,
		Attributes:            map[string]string{},
		Endpoint:              createQueueResult.QueueUrl,
		ReturnSubscriptionArn: true,
	})

	snsClient.SetSubscriptionAttributes(context.TODO(), &sns.SetSubscriptionAttributesInput{
		SubscriptionArn: subscribeResult.SubscriptionArn,
		AttributeName:   aws.String("RawMessageDelivery"),
		AttributeValue:  aws.String("true"),
	})

	message := "{\"IAm\": \"aMessage\"}"
	binaryValueString := "binary-value"
	binaryValue := []byte(binaryValueString)
	subject := "I am a subject"

	attributes := map[string]types.MessageAttributeValue{
		"someKey": {
			BinaryValue: binaryValue,
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
		QueueUrl:            createQueueResult.QueueUrl,
		MaxNumberOfMessages: 10,
	})

	assert.Nil(t, publishErr)
	assert.NotNil(t, publishResponse)

	assert.Nil(t, receiveErr)
	assert.NotNil(t, receiveMessageResponse)

	body := *receiveMessageResponse.Messages[0].Body
	assert.Equal(t, message, body)

	messageAttribute := receiveMessageResponse.Messages[0].MessageAttributes["someKey"]
	assert.Equal(t, "Binary", *messageAttribute.DataType)
	assert.Equal(t, binaryValue, messageAttribute.BinaryValue)

	// check client md5hash and received md5hash
	hasher := md5.New()
	addBytesToHash(hasher, []byte("someKey"))
	addBytesToHash(hasher, []byte(*messageAttribute.DataType))
	hasher.Write([]byte{2})
	addBytesToHash(hasher, messageAttribute.BinaryValue)
	clientHash := hex.EncodeToString(hasher.Sum(nil))

	assert.Equal(t, clientHash, *receiveMessageResponse.Messages[0].MD5OfMessageAttributes)

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

	sqsClient := sqs.NewFromConfig(sdkConfig)

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

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: &app.SyncQueues.Queues["subscribed-queue3"].URL,
	})

	assert.Len(t, receivedMessage.Messages, 1)
	assert.Equal(t, 0, len(receivedMessage.Messages[0].MessageAttributes))
	assert.NotEqual(t, "", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "", *receivedMessage.Messages[0].MD5OfMessageAttributes)

	body := *receivedMessage.Messages[0].Body
	assert.Contains(t, body, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body, "Type")
	assert.Contains(t, body, "MessageId")
	assert.Contains(t, body, "TopicArn")
	assert.Contains(t, body, subject)
	assert.Contains(t, body, "Signature")
	assert.Contains(t, body, "SigningCertURL")
	assert.Contains(t, body, "UnsubscribeURL")
}

func Test_Publish_sqs_json_raw_optional_fields(t *testing.T) {
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

	sqsClient := sqs.NewFromConfig(sdkConfig)

	topicArn := app.SyncTopics.Topics["unit-topic1"].Arn
	message := "{\"IAm\": \"aMessage\"}"
	subject := "I am a subject"
	response, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &message,
		Subject:  &subject,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"test": {
				DataType:    aws.String("String"),
				StringValue: aws.String("value"),
			},
		},
		MessageDeduplicationId: aws.String("dedupe-id"),
		MessageGroupId:         aws.String("group-id"),
		MessageStructure:       aws.String("json"),
		PhoneNumber:            aws.String("phone-number"),
		TargetArn:              aws.String("target-arn"),
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: &app.SyncQueues.Queues["subscribed-queue1"].URL,
	})

	assert.Len(t, receivedMessage.Messages, 1)
	assert.Equal(t, message, *receivedMessage.Messages[0].Body)
	assert.Len(t, receivedMessage.Messages[0].MessageAttributes, 1)
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "9c35d992dee4528f7d20c274d61e16f5", *receivedMessage.Messages[0].MD5OfMessageAttributes)

	assert.Equal(t, "String", *receivedMessage.Messages[0].MessageAttributes["test"].DataType)
	assert.Equal(t, "value", *receivedMessage.Messages[0].MessageAttributes["test"].StringValue)
	assert.Equal(t, []byte{}, receivedMessage.Messages[0].MessageAttributes["test"].BinaryValue)
	assert.Nil(t, receivedMessage.Messages[0].MessageAttributes["test"].BinaryListValues)
	assert.Nil(t, receivedMessage.Messages[0].MessageAttributes["test"].StringListValues)
}

func Test_Publish_sqs_json_not_raw_optional_fields(t *testing.T) {
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

	sqsClient := sqs.NewFromConfig(sdkConfig)

	topicArn := app.SyncTopics.Topics["unit-topic3"].Arn
	message := "{\"default\": \"message-value\", \"http\": \"different-protocol-multi-protocol-not-supported-atm\", \"IAm\": \"toBeIgnored\"}"
	subject := "I am a subject"
	response, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &message,
		Subject:  &subject,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"test": {
				DataType:    aws.String("String"),
				StringValue: aws.String("value"),
			},
		},
		MessageDeduplicationId: aws.String("dedupe-id"),
		MessageGroupId:         aws.String("group-id"),
		MessageStructure:       aws.String("json"),
		PhoneNumber:            aws.String("phone-number"),
		TargetArn:              aws.String("target-arn"),
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: &app.SyncQueues.Queues["subscribed-queue3"].URL,
	})

	assert.Len(t, receivedMessage.Messages, 1)
	assert.NotEqual(t, "", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "", *receivedMessage.Messages[0].MD5OfMessageAttributes)

	body := *receivedMessage.Messages[0].Body
	assert.Contains(t, body, "\"Message\":\"message-value\"")
	assert.Contains(t, body, "Type")
	assert.Contains(t, body, "MessageId")
	assert.Contains(t, body, "TopicArn")
	assert.Contains(t, body, subject)
	assert.Contains(t, body, "Signature")
	assert.Contains(t, body, "SigningCertURL")
	assert.Contains(t, body, "UnsubscribeURL")
	assert.Contains(t, body, "MessageAttributes")

	assert.Contains(t, body, "\"MessageAttributes\":{\"test\":{\"DataType\":\"String\",\"StringValue\":\"value\"}}}")
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

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

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

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &app.SyncQueues.Queues["subscribed-queue1"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Equal(t, 1, len(receivedMessage.Messages))
	assert.Nil(t, err)

	assert.Equal(t, message, *receivedMessage.Messages[0].Body)
	assert.Equal(t, "649b2c548f103e499304eda4d6d4c5a2", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", *receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[0].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[0].MessageId)
	assert.NotNil(t, receivedMessage.Messages[0].ReceiptHandle)
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

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

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

	receivedMessage, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            &app.SyncQueues.Queues["subscribed-queue3"].URL,
		MaxNumberOfMessages: 3,
	})

	assert.Equal(t, 1, len(receivedMessage.Messages))
	assert.Nil(t, err)

	assert.NotEqual(t, "", *receivedMessage.Messages[0].MD5OfBody)
	assert.Equal(t, "", *receivedMessage.Messages[0].MD5OfMessageAttributes)
	assert.Len(t, receivedMessage.Messages[0].MessageAttributes, 0)
	assert.NotNil(t, receivedMessage.Messages[0].MessageId)
	assert.NotNil(t, receivedMessage.Messages[0].ReceiptHandle)

	body0 := *receivedMessage.Messages[0].Body
	assert.Contains(t, body0, "\"Message\":\"{\\\"IAm\\\": \\\"aMessage\\\"}\"")
	assert.Contains(t, body0, "Type")
	assert.Contains(t, body0, "MessageId")
	assert.Contains(t, body0, "TopicArn")
	assert.Contains(t, body0, subject)
	assert.Contains(t, body0, "Signature")
	assert.Contains(t, body0, "SigningCertURL")
	assert.Contains(t, body0, "UnsubscribeURL")
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
}
