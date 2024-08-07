package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_GetSubscriptionAttributes_json_error_subscription_not_found(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	arn := "hogehoge"
	input := sns.GetSubscriptionAttributesInput{
		SubscriptionArn: &arn,
	}

	_, err := snsClient.GetSubscriptionAttributes(context.TODO(), &input)

	assert.Contains(t, err.Error(), strconv.Itoa(http.StatusNotFound))
	assert.Contains(t, err.Error(), "AWS.SimpleNotificationService.NonExistentSubscription")
	assert.Contains(t, err.Error(), "The specified subscription does not exist for this wsdl version.")
}

func Test_GetSubscriptionAttributes_json_success(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	topicName := "new-topic-1"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})

	response, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName)),
		ReturnSubscriptionArn: true,
	})

	getSubscriptionAttributesOutput, err := snsClient.GetSubscriptionAttributes(context.TODO(), &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: response.SubscriptionArn,
	})

	assert.Contains(t, getSubscriptionAttributesOutput.Attributes["Protocol"], "sqs")
	assert.Contains(t, getSubscriptionAttributesOutput.Attributes["TopicArn"], fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName))
	assert.Contains(t, getSubscriptionAttributesOutput.Attributes["Endpoint"], fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName))
	assert.Nil(t, err)
}

func Test_GetSubscriptionAttributes_xml_error_no_subscriptions(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	topicName := "new-topic-1"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})

	snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName)),
		ReturnSubscriptionArn: true,
	})

	e := httpexpect.Default(t, server.URL)

	getSubscriptionAttributesXML := struct {
		Action          string `xml:"Action"`
		Version         string `xml:"Version"`
		SubscriptionArn string `xml:"SubscriptionArn"`
	}{
		Action:          "GetSubscriptionAttributes",
		Version:         "2012-11-05",
		SubscriptionArn: "not-exist-arn",
	}

	r := e.POST("/").
		WithForm(getSubscriptionAttributesXML).
		Expect().
		Status(http.StatusNotFound).
		Body().Raw()

	getSubscriptionAttributesResponse := models.GetSubscriptionAttributesResponse{}
	xml.Unmarshal([]byte(r), &getSubscriptionAttributesResponse)

	assert.Nil(t, getSubscriptionAttributesResponse.Result.Attributes.Entries)

}

func Test_GetSubscriptionAttributes_xml_success(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	topicName := "new-topic-1"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})

	response, _ := snsClient.Subscribe(context.TODO(), &sns.SubscribeInput{
		Protocol:              aws.String("sqs"),
		TopicArn:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName)),
		Attributes:            map[string]string{},
		Endpoint:              aws.String(fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName)),
		ReturnSubscriptionArn: true,
	})

	e := httpexpect.Default(t, server.URL)

	getSubscriptionAttributesXML := struct {
		Action          string `xml:"Action"`
		Version         string `xml:"Version"`
		SubscriptionArn string `xml:"SubscriptionArn"`
	}{
		Action:          "GetSubscriptionAttributes",
		Version:         "2012-11-05",
		SubscriptionArn: *response.SubscriptionArn,
	}

	r := e.POST("/").
		WithForm(getSubscriptionAttributesXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	getSubscriptionAttributesResponse := models.GetSubscriptionAttributesResponse{}
	xml.Unmarshal([]byte(r), &getSubscriptionAttributesResponse)

	expectedAttributes := []models.SubscriptionAttributeEntry{
		{
			Key:   "Owner",
			Value: app.CurrentEnvironment.AccountID,
		},
		{
			Key:   "RawMessageDelivery",
			Value: "false",
		},
		{
			Key:   "TopicArn",
			Value: fmt.Sprintf("%s:%s", af.BASE_SNS_ARN, topicName),
		},
		{
			Key:   "Endpoint",
			Value: fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName),
		},
		{
			Key:   "PendingConfirmation",
			Value: "false",
		},
		{
			Key:   "ConfirmationWasAuthenticated",
			Value: "true",
		}, {
			Key:   "SubscriptionArn",
			Value: *response.SubscriptionArn,
		}, {
			Key:   "Protocol",
			Value: "sqs",
		},
		{
			Key:   "FilterPolicy",
			Value: "null",
		},
	}

	assert.ElementsMatch(t, expectedAttributes, getSubscriptionAttributesResponse.Result.Attributes.Entries)
}
