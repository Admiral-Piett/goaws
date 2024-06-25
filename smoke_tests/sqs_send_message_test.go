package smoke_tests

import (
	"context"
	"encoding/xml"
	"net/http"
	"testing"
	"time"

	"github.com/Admiral-Piett/goaws/app/test"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_SendMessageV1_json_no_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sdkResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	targetQueueUrl := sdkResponse.QueueUrl

	// Assert no messages in the queue before sending message
	getQueueAttributeOutput, _ := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: targetQueueUrl,
	})
	assert.Equal(t, "0", getQueueAttributeOutput.Attributes["ApproximateNumberOfMessages"])

	// Target test: send a message
	targetMessageBody := "Test_SendMessageV1_json_no_attributes"
	sendMessageOutput, _ := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    targetQueueUrl,
		MessageBody: &targetMessageBody,
	})
	assert.NotNil(t, sendMessageOutput.MessageId)

	// Assert 1 message in the queue
	getQueueAttributeOutput, _ = sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: targetQueueUrl,
	})
	assert.Equal(t, "1", getQueueAttributeOutput.Attributes["ApproximateNumberOfMessages"])

	// Receive message and check attribute
	r3, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: targetQueueUrl,
	})
	message := r3.Messages[0]
	assert.Equal(t, targetMessageBody, string(*message.Body))
	assert.Equal(t, 0, len(message.MessageAttributes))
}

func Test_SendMessageV1_json_with_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sdkResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	targetQueueUrl := sdkResponse.QueueUrl

	// Target test: send a message
	targetMessageBody := "Test_SendMessageV1_json_with_attributes"
	attr1_dataType := "String"
	attr1_value := "attr1_value"
	attr2_dataType := "Number"
	attr2_value := "2"
	attr3_dataType := "Binary"
	attr3_value := []byte("attr3_value")
	sendMessageOutput, _ := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:     targetQueueUrl,
		MessageBody:  &targetMessageBody,
		DelaySeconds: 1,
		MessageAttributes: map[string]sqstypes.MessageAttributeValue{
			"attr1": {
				DataType:    &attr1_dataType,
				StringValue: &attr1_value,
			},
			"attr2": {
				DataType:    &attr2_dataType,
				StringValue: &attr2_value,
			},
			"attr3": {
				DataType:    &attr3_dataType,
				BinaryValue: attr3_value,
			},
		},
	})
	assert.NotNil(t, sendMessageOutput.MessageId)

	// Wait for DelaySecond
	time.Sleep(1 * time.Second)

	// Assert 1 message in the queue
	getQueueAttributeOutput, _ := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: targetQueueUrl,
	})
	assert.Equal(t, "1", getQueueAttributeOutput.Attributes["ApproximateNumberOfMessages"])

	// Receive message and check attribute
	r3, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: targetQueueUrl,
	})
	message := r3.Messages[0]
	assert.Equal(t, targetMessageBody, string(*message.Body))
	assert.Equal(t, 3, len(message.MessageAttributes))
	var attr1, attr2, attr3 models.ResultMessageAttribute
	for k, attr := range message.MessageAttributes {
		if k == "attr1" {
			attr1.Name = k
			attr1.Value = &models.ResultMessageAttributeValue{
				DataType:    *attr.DataType,
				StringValue: *attr.StringValue,
				BinaryValue: string(attr.BinaryValue),
			}
		} else if k == "attr2" {
			attr2.Name = k
			attr2.Value = &models.ResultMessageAttributeValue{
				DataType:    *attr.DataType,
				StringValue: *attr.StringValue,
				BinaryValue: string(attr.BinaryValue),
			}
		} else if k == "attr3" {
			attr3.Name = k
			attr3.Value = &models.ResultMessageAttributeValue{
				DataType:    *attr.DataType,
				StringValue: *attr.StringValue,
				BinaryValue: string(attr.BinaryValue),
			}
		}
	}
	assert.Equal(t, "attr1", attr1.Name)
	assert.Equal(t, "String", attr1.Value.DataType)
	assert.Equal(t, "attr1_value", attr1.Value.StringValue)
	assert.Equal(t, "attr2", attr2.Name)
	assert.Equal(t, "Number", attr2.Value.DataType)
	assert.Equal(t, "2", attr2.Value.StringValue)
	assert.Equal(t, "attr3", attr3.Name)
	assert.Equal(t, "Binary", attr3.Value.DataType)
	assert.Equal(t, "YXR0cjNfdmFsdWU=", attr3.Value.BinaryValue) // base64 encoded "attr3_value"
}

func Test_SendMessageV1_json_MaximumMessageSize_TooBig(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sdkResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
		Attributes: map[string]string{
			"MaximumMessageSize": "1",
		},
	})
	targetQueueUrl := sdkResponse.QueueUrl

	// Target test: send a message that is bigger than MaximumMessageSize
	targetMessageBody := "Test_SendMessageV1_json_no_attributes"
	sendMessageOutput, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    targetQueueUrl,
		MessageBody: &targetMessageBody,
	})
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "InvalidParameterValue")
	assert.Nil(t, sendMessageOutput)
}

func Test_SendMessageV1_json_QueueNotExistant(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	// Target test: send a message to a queue that is not exist
	queueUrl := "http://region.host:port/accountID/not-existant-queue"
	targetMessageBody := "Test_SendMessageV1_json_no_attributes"
	sendMessageOutput, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    &queueUrl,
		MessageBody: &targetMessageBody,
	})
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "AWS.SimpleQueueService.NonExistentQueue")
	assert.Nil(t, sendMessageOutput)
}

func Test_SendMessageV1_xml_no_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sdkResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	targetQueueUrl := sdkResponse.QueueUrl

	// Assert no messages in the queue before sending message
	getQueueAttributeOutput, _ := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: targetQueueUrl,
	})
	assert.Equal(t, "0", getQueueAttributeOutput.Attributes["ApproximateNumberOfMessages"])

	// Target test: send a message
	sendMessageXML := struct {
		Action      string `xml:"Action"`
		Version     string `xml:"Version"`
		QueueUrl    string `xml:"QueueUrl"`
		MessageBody string `xml:"MessageBody"`
	}{
		Action:      "SendMessage",
		Version:     "2012-11-05",
		QueueUrl:    *targetQueueUrl,
		MessageBody: "Test Message",
	}
	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(sendMessageXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r3 := models.SendMessageResult{}
	xml.Unmarshal([]byte(r), &r3)
	assert.NotNil(t, r3.MessageId)

	// Assert 1 message in the queue
	getQueueAttributeOutput, _ = sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: targetQueueUrl,
	})
	assert.Equal(t, "1", getQueueAttributeOutput.Attributes["ApproximateNumberOfMessages"])
}

func Test_SendMessageV1_xml_with_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sdkResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})
	targetQueueUrl := sdkResponse.QueueUrl

	// Assert no messages in the queue before sending message
	getQueueAttributeOutput, _ := sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: targetQueueUrl,
	})
	assert.Equal(t, "0", getQueueAttributeOutput.Attributes["ApproximateNumberOfMessages"])

	// Target test: send a message
	sendMessageXML := struct {
		Action       string `xml:"Action"`
		Version      string `xml:"Version"`
		QueueUrl     string `xml:"QueueUrl"`
		MessageBody  string `xml:"MessageBody"`
		DelaySeconds string `xml:"DelaySeconds"`
	}{
		Action:       "SendMessage",
		Version:      "2012-11-05",
		QueueUrl:     *targetQueueUrl,
		MessageBody:  "Test Message",
		DelaySeconds: "1",
	}
	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(sendMessageXML).
		WithFormField("MessageAttribute.1.Name", "attr1").
		WithFormField("MessageAttribute.1.Value.DataType", "String").
		WithFormField("MessageAttribute.1.Value.StringValue", "attr1_value").
		WithFormField("MessageAttribute.2.Name", "attr2").
		WithFormField("MessageAttribute.2.Value.DataType", "Number").
		WithFormField("MessageAttribute.2.Value.StringValue", "2").
		WithFormField("MessageAttribute.3.Name", "attr3").
		WithFormField("MessageAttribute.3.Value.DataType", "Binary").
		WithFormField("MessageAttribute.3.Value.BinaryValue", "YXR0cjNfdmFsdWU=").
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r3 := models.SendMessageResult{}
	xml.Unmarshal([]byte(r), &r3)
	assert.NotNil(t, r3.MessageId)

	// Wait for DelaySecond
	time.Sleep(1 * time.Second)

	// Receive message and check attribute
	receiveMessageBodyXML := struct {
		Action   string `xml:"Action"`
		Version  string `xml:"Version"`
		QueueUrl string `xml:"QueueUrl"`
	}{
		Action:   "ReceiveMessage",
		Version:  "2012-11-05",
		QueueUrl: *targetQueueUrl,
	}
	r = e.POST("/").
		WithForm(receiveMessageBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r4 := models.ReceiveMessageResponse{}
	xml.Unmarshal([]byte(r), &r4)
	message := r4.Result.Messages[0]
	assert.Equal(t, "Test Message", string(message.Body))
	assert.Equal(t, 3, len(message.MessageAttributes))
	var attr1, attr2, attr3 models.ResultMessageAttribute
	for _, attr := range message.MessageAttributes {
		if attr.Name == "attr1" {
			attr1 = *attr
		} else if attr.Name == "attr2" {
			attr2 = *attr
		} else if attr.Name == "attr3" {
			attr3 = *attr
		}
	}
	assert.Equal(t, "attr1", attr1.Name)
	assert.Equal(t, "String", attr1.Value.DataType)
	assert.Equal(t, "attr1_value", attr1.Value.StringValue)
	assert.Equal(t, "attr2", attr2.Name)
	assert.Equal(t, "Number", attr2.Value.DataType)
	assert.Equal(t, "2", attr2.Value.StringValue)
	assert.Equal(t, "attr3", attr3.Name)
	assert.Equal(t, "Binary", attr3.Value.DataType)
	assert.Equal(t, "YXR0cjNfdmFsdWU=", attr3.Value.BinaryValue) // base64 encoded "attr3_value"
}
