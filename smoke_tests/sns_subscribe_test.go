package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/gavv/httpexpect/v2"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
)

func Test_Subscribe_json(t *testing.T) {
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

	response, err := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2")),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2")),
		ReturnSubscriptionArn: true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	app.SyncTopics.Lock()
	defer app.SyncTopics.Unlock()

	subscriptions := app.SyncTopics.Topics["unit-topic2"].Subscriptions
	assert.Len(t, subscriptions, 1)

	expectedFilterPolicy := app.FilterPolicy(nil)
	assert.Equal(t, fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2"), subscriptions[0].EndPoint)
	assert.Equal(t, &expectedFilterPolicy, subscriptions[0].FilterPolicy)
	assert.Equal(t, "sqs", subscriptions[0].Protocol)
	assert.False(t, subscriptions[0].Raw)
	assert.Contains(t, subscriptions[0].SubscriptionArn, fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2"))
	assert.Equal(t, response.SubscriptionArn, &subscriptions[0].SubscriptionArn)
}

func Test_Subscribe_json_with_duplicate_subscription(t *testing.T) {
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

	snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2")),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2")),
		ReturnSubscriptionArn: true,
	})

	response, err := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2")),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2")),
		ReturnSubscriptionArn: true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	app.SyncTopics.Lock()
	defer app.SyncTopics.Unlock()

	subscriptions := app.SyncTopics.Topics["unit-topic2"].Subscriptions
	assert.Len(t, subscriptions, 1)

	expectedFilterPolicy := app.FilterPolicy(nil)
	assert.Equal(t, fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2"), subscriptions[0].EndPoint)
	assert.Equal(t, &expectedFilterPolicy, subscriptions[0].FilterPolicy)
	assert.Equal(t, "sqs", subscriptions[0].Protocol)
	assert.False(t, subscriptions[0].Raw)
	assert.Contains(t, subscriptions[0].SubscriptionArn, fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2"))
	assert.Equal(t, response.SubscriptionArn, &subscriptions[0].SubscriptionArn)
	assert.Equal(t, subscriptions[0].TopicArn, fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2"))
}

func Test_Subscribe_json_with_additional_fields(t *testing.T) {
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

	response, err := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		TopicArn: aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2")),
		Attributes: map[string]string{
			"FilterPolicy":       "{\"filter\": [\"policy\"]}",
			"RawMessageDelivery": "true",
		},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2")),
		ReturnSubscriptionArn: true,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	app.SyncTopics.Lock()
	defer app.SyncTopics.Unlock()

	subscriptions := app.SyncTopics.Topics["unit-topic2"].Subscriptions
	assert.Len(t, subscriptions, 1)

	expectedFilterPolicy := app.FilterPolicy{"filter": []string{"policy"}}
	assert.Equal(t, fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2"), subscriptions[0].EndPoint)
	assert.Equal(t, &expectedFilterPolicy, subscriptions[0].FilterPolicy)
	assert.Equal(t, "sqs", subscriptions[0].Protocol)
	assert.True(t, subscriptions[0].Raw)
	assert.Contains(t, subscriptions[0].SubscriptionArn, fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2"))
	assert.Equal(t, response.SubscriptionArn, &subscriptions[0].SubscriptionArn)
	assert.Equal(t, subscriptions[0].TopicArn, fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2"))
}

func Test_Subscribe_xml(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	requestBody := struct {
		Action   string `schema:"Subscribe"`
		TopicArn string `schema:"TopicArn"`
		Endpoint string `schema:"Endpoint"`
		Protocol string `schema:"Protocol"`
	}{
		Action:   "Subscribe",
		TopicArn: fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2"),
		Endpoint: fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2"),
		Protocol: "sqs",
	}

	r := e.POST("/").
		WithForm(requestBody).
		WithFormField("Attributes.entry.1.key", "RawMessageDelivery").
		WithFormField("Attributes.entry.1.value", "true").
		WithFormField("Attributes.entry.2.key", "FilterPolicy").
		WithFormField("Attributes.entry.2.value", "{\"filter\": [\"policy\"]}").
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	response := models.SubscribeResponse{}
	xml.Unmarshal([]byte(r), &response)
	subscriptions := app.SyncTopics.Topics["unit-topic2"].Subscriptions
	assert.Len(t, subscriptions, 1)

	expectedFilterPolicy := app.FilterPolicy{"filter": []string{"policy"}}
	assert.Equal(t, fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, "unit-queue2"), subscriptions[0].EndPoint)
	assert.Equal(t, &expectedFilterPolicy, subscriptions[0].FilterPolicy)
	assert.Equal(t, "sqs", subscriptions[0].Protocol)
	assert.True(t, subscriptions[0].Raw)
	assert.Contains(t, subscriptions[0].SubscriptionArn, fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2"))
	assert.Equal(t, response.Result.SubscriptionArn, subscriptions[0].SubscriptionArn)
	assert.Equal(t, subscriptions[0].TopicArn, fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, "unit-topic2"))
}
