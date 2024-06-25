package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app/models"
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

func Test_GetQueueAttributes_json_all(t *testing.T) {
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
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName:  &af.QueueName,
		Attributes: attributes,
	})

	sdkResponse, err := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl:       &af.QueueUrl,
		AttributeNames: []types.QueueAttributeName{"All"},
	})

	dupe, _ := copystructure.Copy(attributes)
	expectedAttributes, _ := dupe.(map[string]string)
	expectedAttributes["RedrivePolicy"] = fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue)
	expectedAttributes["ApproximateNumberOfMessages"] = "0"
	expectedAttributes["ApproximateNumberOfMessagesNotVisible"] = "0"
	expectedAttributes["CreatedTimestamp"] = "0000000000"
	expectedAttributes["LastModifiedTimestamp"] = "0000000000"
	expectedAttributes["QueueArn"] = fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName)
	assert.Nil(t, err)
	assert.Equal(t, expectedAttributes, sdkResponse.Attributes)
}

func Test_GetQueueAttributes_json_specific_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
		Attributes: map[string]string{
			"DelaySeconds":           "1",
			"MaximumMessageSize":     "2",
			"MessageRetentionPeriod": "3",
		},
	})

	sdkResponse, err := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl:       &af.QueueUrl,
		AttributeNames: []types.QueueAttributeName{"DelaySeconds"},
	})

	expectedAttributes := map[string]string{
		"DelaySeconds": "1",
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedAttributes, sdkResponse.Attributes)
}

func Test_GetQueueAttributes_json_missing_attribute_name_returns_all(t *testing.T) {
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
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName:  &af.QueueName,
		Attributes: attributes,
	})

	sdkResponse, err := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: &af.QueueUrl,
	})

	dupe, _ := copystructure.Copy(attributes)
	expectedAttributes, _ := dupe.(map[string]string)
	expectedAttributes["RedrivePolicy"] = fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue)
	expectedAttributes["ApproximateNumberOfMessages"] = "0"
	expectedAttributes["ApproximateNumberOfMessagesNotVisible"] = "0"
	expectedAttributes["CreatedTimestamp"] = "0000000000"
	expectedAttributes["LastModifiedTimestamp"] = "0000000000"
	expectedAttributes["QueueArn"] = fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName)
	assert.Nil(t, err)
	assert.Equal(t, expectedAttributes, sdkResponse.Attributes)
}

func Test_GetQueueAttributes_xml_all(t *testing.T) {
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
	attributes := map[string]string{
		"DelaySeconds":                  "1",
		"MaximumMessageSize":            "2",
		"MessageRetentionPeriod":        "3",
		"ReceiveMessageWaitTimeSeconds": "4",
		"VisibilityTimeout":             "5",
		"RedrivePolicy":                 fmt.Sprintf(`{"maxReceiveCount":"100","deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue),
	}
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName:  &af.QueueName,
		Attributes: attributes,
	})

	r := e.POST("/").
		WithForm(sf.GetQueueAttributesRequestBodyXML).
		WithFormField("AttributeName.1", "All").
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	dupe, _ := copystructure.Copy(sf.BASE_GET_QUEUE_ATTRIBUTES_RESPONSE)
	expectedResponse, _ := dupe.(models.GetQueueAttributesResponse)
	expectedResponse.Result.Attrs[0].Value = "1"
	expectedResponse.Result.Attrs[1].Value = "2"
	expectedResponse.Result.Attrs[2].Value = "3"
	expectedResponse.Result.Attrs[3].Value = "4"
	expectedResponse.Result.Attrs[4].Value = "5"
	expectedResponse.Result.Attrs = append(expectedResponse.Result.Attrs, models.Attribute{
		Name:  "RedrivePolicy",
		Value: fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue),
	})

	r1 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r1)
	assert.Equal(t, expectedResponse, r1)
}

func Test_GetQueueAttributes_xml_select_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	attributes := map[string]string{
		"DelaySeconds":           "1",
		"MaximumMessageSize":     "2",
		"MessageRetentionPeriod": "3",
		//"Policy":                        "{\"this-is\": \"the-policy\"}",
		"ReceiveMessageWaitTimeSeconds": "4",
		"VisibilityTimeout":             "5",
		//"RedriveAllowPolicy":            "{\"this-is\": \"the-redrive-allow-policy\"}",
	}
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName:  &af.QueueName,
		Attributes: attributes,
	})

	body := struct {
		Action   string `xml:"Action"`
		Version  string `xml:"Version"`
		QueueUrl string `xml:"QueueUrl"`
	}{
		Action:   "GetQueueAttributes",
		Version:  "2012-11-05",
		QueueUrl: af.QueueUrl,
	}

	r := e.POST("/").
		WithForm(body).
		WithFormField("AttributeName.1", "DelaySeconds").
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	expectedResponse := models.GetQueueAttributesResponse{
		Xmlns: models.BASE_XMLNS,
		Result: models.GetQueueAttributesResult{
			Attrs: []models.Attribute{
				{
					Name:  "DelaySeconds",
					Value: "1",
				},
			},
		},
		Metadata: models.BASE_RESPONSE_METADATA,
	}

	r1 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r1)
	assert.Equal(t, expectedResponse, r1)
}

func Test_GetQueueAttributes_xml_missing_attribute_name_returns_all(t *testing.T) {
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
	attributes := map[string]string{
		"DelaySeconds":           "1",
		"MaximumMessageSize":     "2",
		"MessageRetentionPeriod": "3",
		//"Policy":                        "{\"this-is\": \"the-policy\"}",
		"ReceiveMessageWaitTimeSeconds": "4",
		"VisibilityTimeout":             "5",
		"RedrivePolicy":                 fmt.Sprintf(`{"maxReceiveCount":"100","deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue),
	}
	sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName:  &af.QueueName,
		Attributes: attributes,
	})

	body := struct {
		Action   string `xml:"Action"`
		Version  string `xml:"Version"`
		QueueUrl string `xml:"QueueUrl"`
	}{
		Action:   "GetQueueAttributes",
		Version:  "2012-11-05",
		QueueUrl: af.QueueUrl,
	}

	r := e.POST("/").
		WithForm(body).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	dupe, _ := copystructure.Copy(sf.BASE_GET_QUEUE_ATTRIBUTES_RESPONSE)
	expectedResponse, _ := dupe.(models.GetQueueAttributesResponse)
	expectedResponse.Result.Attrs[0].Value = "1"
	expectedResponse.Result.Attrs[1].Value = "2"
	expectedResponse.Result.Attrs[2].Value = "3"
	expectedResponse.Result.Attrs[3].Value = "4"
	expectedResponse.Result.Attrs[4].Value = "5"
	expectedResponse.Result.Attrs = append(expectedResponse.Result.Attrs, models.Attribute{
		Name:  "RedrivePolicy",
		Value: fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue),
	})

	r1 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r1)
	assert.Equal(t, expectedResponse, r1)
}
