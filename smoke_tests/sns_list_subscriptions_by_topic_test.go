package smoke_tests

import (
	"context"
	"encoding/xml"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ListSubscriptionsByTopic_Success_Multiple_Subscriptions(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse1, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	createQueueResponse2, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: aws.String("new-queue-2"),
	})

	getQueueAttributesOutput1, _ := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: createQueueResponse1.QueueUrl,
	})

	getQueueAttributesOutput2, _ := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: createQueueResponse2.QueueUrl,
	})

	protocol := aws.String("sqs")
	topicName := "new-topic-1"
	createTopicResponse, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})

	subscribeResponse1, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              protocol,
		TopicArn:              createTopicResponse.TopicArn,
		Attributes:            map[string]string{},
		Endpoint:              aws.String(getQueueAttributesOutput1.Attributes["QueueArn"]),
		ReturnSubscriptionArn: true,
	})

	subscribeResponse2, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              protocol,
		TopicArn:              createTopicResponse.TopicArn,
		Attributes:            map[string]string{},
		Endpoint:              aws.String(getQueueAttributesOutput2.Attributes["QueueArn"]),
		ReturnSubscriptionArn: true,
	})

	listSubscriptionsByTopicOutput, _ := snsClient.ListSubscriptionsByTopic(context.TODO(), &sns.ListSubscriptionsByTopicInput{
		TopicArn: createTopicResponse.TopicArn,
	})

	assert.NotNil(t, listSubscriptionsByTopicOutput)
	assert.Len(t, listSubscriptionsByTopicOutput.Subscriptions, 2)

	subscriptionMap := make(map[string]types.Subscription, 2)
	for _, subscription := range listSubscriptionsByTopicOutput.Subscriptions {
		subscriptionMap[*subscription.SubscriptionArn] = subscription
	}

	subscription1, exists := subscriptionMap[*subscribeResponse1.SubscriptionArn]
	assert.True(t, exists)
	assert.Equal(t, createTopicResponse.TopicArn, subscription1.TopicArn)
	assert.Equal(t, *subscribeResponse1.SubscriptionArn, *subscription1.SubscriptionArn)
	assert.Equal(t, *protocol, *(subscription1.Protocol))
	assert.Equal(t, app.CurrentEnvironment.AccountID, *(subscription1.Owner))
	assert.Equal(t, getQueueAttributesOutput1.Attributes["QueueArn"], *(subscription1.Endpoint))

	subscription2, exists := subscriptionMap[*subscribeResponse2.SubscriptionArn]
	assert.True(t, exists)
	assert.Equal(t, createTopicResponse.TopicArn, subscription2.TopicArn)
	assert.Equal(t, subscribeResponse2.SubscriptionArn, subscription2.SubscriptionArn)
	assert.Equal(t, *protocol, *(subscription2.Protocol))
	assert.Equal(t, app.CurrentEnvironment.AccountID, *(subscription2.Owner))
	assert.Equal(t, getQueueAttributesOutput2.Attributes["QueueArn"], *(subscription2.Endpoint))
}

func Test_ListSubscriptionsByTopic_Json_Not_Found(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	listSubscriptionsByTopicOutput, err := snsClient.ListSubscriptionsByTopic(context.TODO(), &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String("not exist arn"),
	})

	assert.Nil(t, listSubscriptionsByTopicOutput)
	assert.Contains(t, err.Error(), "AWS.SimpleNotificationService.NonExistentTopic")
}

func Test_ListSubscriptionsByTopic_Xml_Success_Multiple_Subscriptions(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	createQueueResponse1, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	createQueueResponse2, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: aws.String("new-queue-2"),
	})

	getQueueAttributesOutput1, _ := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: createQueueResponse1.QueueUrl,
	})

	getQueueAttributesOutput2, _ := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: createQueueResponse2.QueueUrl,
	})

	protocol := aws.String("sqs")
	topicName := "new-topic-1"
	createTopicResponse, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})

	subscribeResponse1, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              protocol,
		TopicArn:              createTopicResponse.TopicArn,
		Attributes:            map[string]string{},
		Endpoint:              aws.String(getQueueAttributesOutput1.Attributes["QueueArn"]),
		ReturnSubscriptionArn: true,
	})

	subscribeResponse2, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              protocol,
		TopicArn:              createTopicResponse.TopicArn,
		Attributes:            map[string]string{},
		Endpoint:              aws.String(getQueueAttributesOutput2.Attributes["QueueArn"]),
		ReturnSubscriptionArn: true,
	})

	requestBody := struct {
		Action   string `xml:"Action"`
		TopicArn string `xml:"TopicArn"`
		Version  string `xml:"Version"`
	}{
		Action:   "ListSubscriptionsByTopic",
		TopicArn: *createTopicResponse.TopicArn,
		Version:  "2012-11-05",
	}
	e := httpexpect.Default(t, server.URL)

	r := e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	listSubscriptionsByTopicResponse := models.ListSubscriptionsByTopicResponse{}
	xml.Unmarshal([]byte(r), &listSubscriptionsByTopicResponse)

	assert.NotNil(t, listSubscriptionsByTopicResponse)
	assert.Len(t, listSubscriptionsByTopicResponse.Result.Subscriptions.Member, 2)

	expectedMember := []models.TopicMemberResult{
		{
			TopicArn:        *createTopicResponse.TopicArn,
			SubscriptionArn: *subscribeResponse1.SubscriptionArn,
			Protocol:        *protocol,
			Owner:           app.CurrentEnvironment.AccountID,
			Endpoint:        getQueueAttributesOutput1.Attributes["QueueArn"],
		},
		{
			TopicArn:        *createTopicResponse.TopicArn,
			SubscriptionArn: *subscribeResponse2.SubscriptionArn,
			Protocol:        *protocol,
			Owner:           app.CurrentEnvironment.AccountID,
			Endpoint:        getQueueAttributesOutput2.Attributes["QueueArn"],
		},
	}

	assert.ElementsMatch(t, expectedMember, listSubscriptionsByTopicResponse.Result.Subscriptions.Member)
}

func Test_ListSubscriptionsByTopic_Xml_Not_Found(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	requestBody := struct {
		Action   string `xml:"Action"`
		TopicArn string `xml:"TopicArn"`
		Version  string `xml:"Version"`
	}{
		Action:   "ListSubscriptionsByTopic",
		TopicArn: "not exist arn",
		Version:  "2012-11-05",
	}

	r := e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusBadRequest).
		Body().Raw()

	listSubscriptionsByTopicResponse := models.ListSubscriptionsByTopicResponse{}
	xml.Unmarshal([]byte(r), &listSubscriptionsByTopicResponse)
	assert.Empty(t, listSubscriptionsByTopicResponse.Result.Subscriptions.Member)

}
