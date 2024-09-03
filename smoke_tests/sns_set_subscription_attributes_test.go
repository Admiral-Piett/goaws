package smoke_tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_SetSubscriptionAttributes_json_success(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	// Create a subscription
	queueName := "new-queue-1"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName,
	})
	topicName := "new-topic-1"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})
	subResp, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, queueName)),
		ReturnSubscriptionArn: true,
	})

	// Check initial attribute
	getResp, err := snsClient.GetSubscriptionAttributes(context.TODO(), &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: subResp.SubscriptionArn,
	})
	assert.Equal(t, "false", getResp.Attributes["RawMessageDelivery"])
	assert.Nil(t, err)

	// Target test: Set attribute
	attrName := "RawMessageDelivery"
	attrValue := "true"
	_, err = snsClient.SetSubscriptionAttributes(context.TODO(), &sns.SetSubscriptionAttributesInput{
		SubscriptionArn: subResp.SubscriptionArn,
		AttributeName:   &attrName,
		AttributeValue:  &attrValue,
	})
	assert.Nil(t, err)

	// Assert the attribute has been updated
	getResp, err = snsClient.GetSubscriptionAttributes(context.TODO(), &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: subResp.SubscriptionArn,
	})
	assert.Equal(t, "true", getResp.Attributes["RawMessageDelivery"])
	assert.Nil(t, err)
}

func Test_SetSubscriptionAttributes_json_error_SubscriptionNotExistence(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	// Target test: Set attribute
	subscriptionArn := "not existence sub"
	attrName := "RawMessageDelivery"
	attrValue := "true"
	response, err := snsClient.SetSubscriptionAttributes(context.TODO(), &sns.SetSubscriptionAttributesInput{
		SubscriptionArn: &subscriptionArn,
		AttributeName:   &attrName,
		AttributeValue:  &attrValue,
	})
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "AWS.SimpleNotificationService.NonExistentSubscription")
	assert.Nil(t, response)
}

func Test_SetSubscriptionAttributes_xml_success(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	// Create a subscription
	queueName := "new-queue-1"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName,
	})
	topicName := "new-topic-1"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})
	subResp, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, queueName)),
		ReturnSubscriptionArn: true,
	})

	// Check initial attribute
	getResp, err := snsClient.GetSubscriptionAttributes(context.TODO(), &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: subResp.SubscriptionArn,
	})
	assert.Equal(t, "false", getResp.Attributes["RawMessageDelivery"])
	assert.Nil(t, err)

	// Target test: Set attribute
	setSubscriptionAttributesXML := struct {
		Action          string `xml:"Action"`
		Version         string `xml:"Version"`
		SubscriptionArn string `xml:"SubscriptionArn"`
		AttributeName   string `xml:"AttributeName"`
		AttributeValue  string `xml:"AttributeValue"`
	}{
		Action:          "SetSubscriptionAttributes",
		Version:         "2012-11-05",
		SubscriptionArn: *subResp.SubscriptionArn,
		AttributeName:   "RawMessageDelivery",
		AttributeValue:  "true",
	}
	e := httpexpect.Default(t, server.URL)
	e.POST("/").
		WithForm(setSubscriptionAttributesXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	// Assert the attribute has been updated
	getResp, err = snsClient.GetSubscriptionAttributes(context.TODO(), &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: subResp.SubscriptionArn,
	})
	assert.Equal(t, "true", getResp.Attributes["RawMessageDelivery"])
	assert.Nil(t, err)
}
