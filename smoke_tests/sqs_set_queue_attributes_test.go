package smoke_tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	sf "github.com/Admiral-Piett/goaws/smoke_tests/fixtures"
	"github.com/gavv/httpexpect/v2"

	"github.com/mitchellh/copystructure"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/stretchr/testify/assert"
)

func Test_SetQueueAttributes_json_multiple(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	redriveQueue := "redrive-queue"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &redriveQueue,
	})
	queueName := "unit-queue1"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName,
	})
	attributes := map[string]string{
		"DelaySeconds":           "1",
		"MaximumMessageSize":     "2",
		"MessageRetentionPeriod": "3",
		//"Policy":                        "{\"this-is\": \"the-policy\"}",
		"ReceiveMessageWaitTimeSeconds": "4",
		"VisibilityTimeout":             "5",
		"RedrivePolicy":                 fmt.Sprintf(`{"maxReceiveCount":"100","deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue),
		//"RedriveAllowPolicy":            "{\"this-is\": \"the-redrive-allow-policy\"}",
	}

	queueUrl := fmt.Sprintf("%s/%s", af.BASE_URL, queueName)
	_, err := sqsClient.SetQueueAttributes(context.TODO(), &sqs.SetQueueAttributesInput{
		QueueUrl:   &queueUrl,
		Attributes: attributes,
	})

	assert.Nil(t, err)

	sdkResponse, err := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl:       &queueUrl,
		AttributeNames: []types.QueueAttributeName{"All"},
	})

	dupe, _ := copystructure.Copy(attributes)
	expectedAttributes, _ := dupe.(map[string]string)
	expectedAttributes["RedrivePolicy"] = fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue)
	expectedAttributes["ApproximateNumberOfMessages"] = "0"
	expectedAttributes["ApproximateNumberOfMessagesNotVisible"] = "0"
	expectedAttributes["CreatedTimestamp"] = "0000000000"
	expectedAttributes["LastModifiedTimestamp"] = "0000000000"
	expectedAttributes["QueueArn"] = fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, queueName)
	assert.Nil(t, err)
	assert.Equal(t, expectedAttributes, sdkResponse.Attributes)
}

func Test_SetQueueAttributes_json_single(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	queueName := "unit-queue1"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &queueName,
	})
	attributes := map[string]string{
		"DelaySeconds": "1",
	}

	queueUrl := fmt.Sprintf("%s/%s", af.BASE_URL, queueName)
	_, err := sqsClient.SetQueueAttributes(context.TODO(), &sqs.SetQueueAttributesInput{
		QueueUrl:   &queueUrl,
		Attributes: attributes,
	})

	assert.Nil(t, err)

	sdkResponse, err := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl:       &queueUrl,
		AttributeNames: []types.QueueAttributeName{"All"},
	})

	dupe, _ := copystructure.Copy(attributes)
	expectedAttributes, _ := dupe.(map[string]string)
	expectedAttributes["ReceiveMessageWaitTimeSeconds"] = "0"
	expectedAttributes["VisibilityTimeout"] = "0"
	expectedAttributes["MaximumMessageSize"] = "0"
	expectedAttributes["MessageRetentionPeriod"] = "0"
	expectedAttributes["ApproximateNumberOfMessages"] = "0"
	expectedAttributes["ApproximateNumberOfMessagesNotVisible"] = "0"
	expectedAttributes["CreatedTimestamp"] = "0000000000"
	expectedAttributes["LastModifiedTimestamp"] = "0000000000"
	expectedAttributes["QueueArn"] = fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, queueName)
	assert.Nil(t, err)
	assert.Equal(t, expectedAttributes, sdkResponse.Attributes)
}

func Test_SetQueueAttributes_xml_all(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	redriveQueue := "redrive-queue"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &redriveQueue,
	})
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	e.POST("/").
		WithForm(sf.SetQueueAttributesRequestBodyXML).
		WithFormField("Attribute.1.Name", "VisibilityTimeout").
		WithFormField("Attribute.1.Value", "5").
		WithFormField("Attribute.2.Name", "MaximumMessageSize").
		WithFormField("Attribute.2.Value", "2").
		WithFormField("Attribute.3.Name", "DelaySeconds").
		WithFormField("Attribute.3.Value", "1").
		WithFormField("Attribute.4.Name", "MessageRetentionPeriod").
		WithFormField("Attribute.4.Value", "3").
		WithFormField("Attribute.5.Name", "Policy").
		WithFormField("Attribute.5.Value", "{\"this-is\": \"the-policy\"}").
		WithFormField("Attribute.6.Name", "ReceiveMessageWaitTimeSeconds").
		WithFormField("Attribute.6.Value", "4").
		WithFormField("Attribute.7.Name", "RedrivePolicy").
		WithFormField("Attribute.7.Value", fmt.Sprintf("{\"maxReceiveCount\": 100, \"deadLetterTargetArn\":\"%s:%s\"}", af.BASE_SQS_ARN, redriveQueue)).
		WithFormField("Attribute.8.Name", "RedriveAllowPolicy").
		WithFormField("Attribute.8.Value", "{\"this-is\": \"the-redrive-allow-policy\"}").
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	sdkResponse, err := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl:       &af.QueueUrl,
		AttributeNames: []types.QueueAttributeName{"All"},
	})

	expectedAttributes := map[string]string{
		"DelaySeconds":           "1",
		"MaximumMessageSize":     "2",
		"MessageRetentionPeriod": "3",
		//"Policy":                        "{\"this-is\": \"the-policy\"}",
		"ReceiveMessageWaitTimeSeconds": "4",
		"VisibilityTimeout":             "5",
		"RedrivePolicy":                 fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue),
		//"RedriveAllowPolicy":            "{\"this-is\": \"the-redrive-allow-policy\"}",
		"ApproximateNumberOfMessages":           "0",
		"ApproximateNumberOfMessagesNotVisible": "0",
		"CreatedTimestamp":                      "0000000000",
		"LastModifiedTimestamp":                 "0000000000",
		"QueueArn":                              fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName),
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedAttributes, sdkResponse.Attributes)
}

func Test_SetQueueAttributes_xml_single(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	redriveQueue := "redrive-queue"
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &redriveQueue,
	})
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	e.POST("/").
		WithForm(sf.SetQueueAttributesRequestBodyXML).
		WithFormField("Attribute.1.Name", "VisibilityTimeout").
		WithFormField("Attribute.1.Value", "5").
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	sdkResponse, err := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl:       &af.QueueUrl,
		AttributeNames: []types.QueueAttributeName{"All"},
	})

	expectedAttributes := map[string]string{
		"DelaySeconds":                          "0",
		"MaximumMessageSize":                    "0",
		"MessageRetentionPeriod":                "0",
		"ReceiveMessageWaitTimeSeconds":         "0",
		"VisibilityTimeout":                     "5",
		"ApproximateNumberOfMessages":           "0",
		"ApproximateNumberOfMessagesNotVisible": "0",
		"CreatedTimestamp":                      "0000000000",
		"LastModifiedTimestamp":                 "0000000000",
		"QueueArn":                              fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName),
	}
	assert.Nil(t, err)
	assert.Equal(t, expectedAttributes, sdkResponse.Attributes)
}
