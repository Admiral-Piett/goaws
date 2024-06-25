package smoke_tests

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_GetQueueUrlV1_json_success_retrieve_queue_url(t *testing.T) {

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
	})

	getQueueUrlOutput, _ := sqsClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: &af.QueueName,
	})
	assert.Contains(t, string(*getQueueUrlOutput.QueueUrl), fmt.Sprintf("%s/%s", af.BASE_URL, af.QueueName))
}

func Test_GetQueueUrlV1_json_error_not_found_queue(t *testing.T) {

	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	getQueueUrlOutput, err := sqsClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: &af.QueueName})

	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "AWS.SimpleQueueService.NonExistentQueue")
	assert.Contains(t, err.Error(), "The specified queue does not exist for this wsdl version.")
	assert.Nil(t, getQueueUrlOutput)
}

func Test_GetQueueUrlV1_xml_success_retrieve_queue_url(t *testing.T) {
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
	})

	e := httpexpect.Default(t, server.URL)

	getQueueUrlRequestBodyXML := struct {
		Action                 string `xml:"Action"`
		QueueName              string `xml:"QueueName"`
		QueueOwnerAWSAccountId string `xml:"QueueOwnerAWSAccountId"`
		Version                string `xml:"Version"`
	}{
		Action:                 "GetQueueUrl",
		QueueName:              af.QueueName,
		QueueOwnerAWSAccountId: "hogehoge",
		Version:                "2012-11-05",
	}

	r := e.POST("/").
		WithForm(getQueueUrlRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	r1 := models.GetQueueUrlResponse{}
	xml.Unmarshal([]byte(r), &r1)

	assert.Contains(t, string(r1.Result.QueueUrl), fmt.Sprintf("%s/%s", af.BASE_URL, af.QueueName))
}

func Test_GetQueueUrlV1_xml_error_not_found_queue(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	getQueueUrlRequestBodyXML := struct {
		Action                 string `xml:"Action"`
		QueueName              string `xml:"QueueName"`
		QueueOwnerAWSAccountId string `xml:"QueueOwnerAWSAccountId"`
		Version                string `xml:"Version"`
	}{
		Action:                 "GetQueueUrl",
		QueueName:              af.QueueName,
		QueueOwnerAWSAccountId: "hogehoge",
		Version:                "2012-11-05",
	}

	r := e.POST("/").
		WithForm(getQueueUrlRequestBodyXML).
		Expect().
		Status(http.StatusBadRequest).
		Body().Raw()

	r1 := models.ErrorResponse{}
	xml.Unmarshal([]byte(r), &r1)

	assert.Contains(t, r1.Result.Type, "Not Found")
	assert.Contains(t, r1.Result.Code, "AWS.SimpleQueueService.NonExistentQueue")
	assert.Contains(t, r1.Result.Message, "The specified queue does not exist for this wsdl version.")
}
