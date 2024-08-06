package smoke_tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"encoding/xml"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/stretchr/testify/assert"

	af "github.com/Admiral-Piett/goaws/app/fixtures"

	"github.com/gavv/httpexpect/v2"
)

func Test_List_Subscriptions_json_no_subscriptions(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	sdkResponse, err := snsClient.ListSubscriptions(context.TODO(), &sns.ListSubscriptionsInput{})

	assert.Nil(t, err)
	assert.Len(t, sdkResponse.Subscriptions, 0)
}

func Test_List_Subscriptions_json_multiple_subscriptions(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "NoQueueAttributeDefaults")
	defer func() {
		server.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	// add new topics to subscribe to
	topicName := "new-topic-1"
	createTopicResponse, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})
	assert.Contains(t, *createTopicResponse.TopicArn, topicName)

	topicName2 := "new-topic-2"
	createTopicResponse2, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName2,
	})
	assert.Contains(t, *createTopicResponse2.TopicArn, topicName2)

	// subscribe to new topics
	subscribeResponse, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue1")),
		ReturnSubscriptionArn: true,
	})

	subscribeResponse2, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName2)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue1")),
		ReturnSubscriptionArn: true,
	})

	assert.NotNil(t, subscribeResponse)
	assert.NotNil(t, subscribeResponse2)

	app.SyncTopics.Lock()
	defer app.SyncTopics.Unlock()

	// check listed subscriptions
	sdkResponse, err := snsClient.ListSubscriptions(context.TODO(), &sns.ListSubscriptionsInput{})
	assert.Nil(t, err)
	assert.Len(t, sdkResponse.Subscriptions, 2)

	assert.NotEqual(t, sdkResponse.Subscriptions[0], sdkResponse.Subscriptions[1])

	assert.Equal(t, *sdkResponse.Subscriptions[0].TopicArn, *createTopicResponse.TopicArn)
	assert.Equal(t, *sdkResponse.Subscriptions[0].SubscriptionArn, *subscribeResponse.SubscriptionArn)

	assert.Equal(t, *sdkResponse.Subscriptions[1].TopicArn, *createTopicResponse2.TopicArn)
	assert.Equal(t, *sdkResponse.Subscriptions[1].SubscriptionArn, *subscribeResponse2.SubscriptionArn)
}

func Test_List_Subscriptions_xml_no_subscriptions(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	r := e.POST("/").
		WithForm(struct {
			Action  string `xml:"Action"`
			Version string `xml:"Version"`
		}{
			Action:  "ListSubscriptions",
			Version: "2012-11-05",
		}).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	listSubscriptionsResponseObject := models.ListSubscriptionsResponse{}
	xml.Unmarshal([]byte(r), &listSubscriptionsResponseObject)

	assert.Equal(t, "http://queue.amazonaws.com/doc/2012-11-05/", listSubscriptionsResponseObject.Xmlns)
	assert.Len(t, listSubscriptionsResponseObject.Result.Subscriptions.Member, 0)
}

func Test_List_Subscriptions_xml_multiple_subscriptions(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	// add new topics to subscribe to
	topicName := "new-topic-1"
	createTopicResponse, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})
	assert.Contains(t, *createTopicResponse.TopicArn, topicName)

	topicName2 := "new-topic-2"
	createTopicResponse2, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName2,
	})
	assert.Contains(t, *createTopicResponse2.TopicArn, topicName2)

	// subscribe to new topics
	subscribeResponse, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue1")),
		ReturnSubscriptionArn: true,
	})

	subscribeResponse2, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName2)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue1")),
		ReturnSubscriptionArn: true,
	})
	assert.NotNil(t, subscribeResponse)
	assert.NotNil(t, subscribeResponse2)

	e := httpexpect.Default(t, server.URL)

	// check listed subscriptions
	r := e.POST("/").
		WithForm(struct {
			Action  string `xml:"Action"`
			Version string `xml:"Version"`
		}{
			Action:  "ListSubscriptions",
			Version: "2012-11-05",
		}).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	listSubscriptionsResponseObject := models.ListSubscriptionsResponse{}
	xml.Unmarshal([]byte(r), &listSubscriptionsResponseObject)

	assert.Equal(t, "http://queue.amazonaws.com/doc/2012-11-05/", listSubscriptionsResponseObject.Xmlns)
	assert.Len(t, listSubscriptionsResponseObject.Result.Subscriptions.Member, 2)
	assert.NotEqual(t, listSubscriptionsResponseObject.Result.Subscriptions.Member[0].TopicArn, listSubscriptionsResponseObject.Result.Subscriptions.Member[1].TopicArn)

	assert.Equal(t, listSubscriptionsResponseObject.Result.Subscriptions.Member[0].TopicArn, *createTopicResponse.TopicArn)
	assert.Equal(t, listSubscriptionsResponseObject.Result.Subscriptions.Member[0].SubscriptionArn, *subscribeResponse.SubscriptionArn)

	assert.Equal(t, listSubscriptionsResponseObject.Result.Subscriptions.Member[1].TopicArn, *createTopicResponse2.TopicArn)
	assert.Equal(t, listSubscriptionsResponseObject.Result.Subscriptions.Member[1].SubscriptionArn, *subscribeResponse2.SubscriptionArn)
}
