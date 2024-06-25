package smoke_tests

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/mitchellh/copystructure"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/stretchr/testify/assert"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
	sf "github.com/Admiral-Piett/goaws/smoke_tests/fixtures"

	"github.com/gavv/httpexpect/v2"
)

func Test_CreateQueueV1_json_no_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	sqsClient := sqs.NewFromConfig(sdkConfig)
	sdkResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
	})

	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *sdkResponse.QueueUrl)

	r := e.POST("/").
		WithForm(sf.ListQueuesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp2 := models.ListQueuesResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Result:   models.ListQueuesResult{QueueUrls: []string{fmt.Sprintf("%s/new-queue-1", af.BASE_URL)}},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}
	r2 := models.ListQueuesResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Equal(t, exp2, r2)

	r = e.POST("/").
		WithForm(sf.GetQueueAttributesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	r3 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r3)
	assert.Equal(t, sf.BASE_GET_QUEUE_ATTRIBUTES_RESPONSE, r3)
}

func Test_CreateQueueV1_json_with_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	redriveQueue := "redrive-queue"

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)

	sqsClient := sqs.NewFromConfig(sdkConfig)
	sdkResponse, _ := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &redriveQueue,
	})

	sdkResponse, err := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: &af.QueueName,
		Attributes: map[string]string{
			"DelaySeconds":           "1",
			"MaximumMessageSize":     "2",
			"MessageRetentionPeriod": "3",
			//"Policy":                        "{\"this-is\": \"the-policy\"}",
			"ReceiveMessageWaitTimeSeconds": "4",
			"VisibilityTimeout":             "5",
			"RedrivePolicy":                 fmt.Sprintf(`{"maxReceiveCount":"100","deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue),
			//"RedriveAllowPolicy":            "{\"this-is\": \"the-redrive-allow-policy\"}",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("%s/new-queue-1", af.BASE_URL), *sdkResponse.QueueUrl)

	r := e.POST("/").
		WithForm(sf.ListQueuesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	r2 := models.ListQueuesResponse{}
	xml.Unmarshal([]byte(r), &r2)

	assert.Equal(t, models.BASE_XMLNS, r2.Xmlns)
	assert.Equal(t, models.BASE_RESPONSE_METADATA, r2.Metadata)
	assert.Equal(t, 2, len(r2.Result.QueueUrls))
	assert.Contains(t, r2.Result.QueueUrls, fmt.Sprintf("%s/%s", af.BASE_URL, redriveQueue))
	assert.Contains(t, r2.Result.QueueUrls, fmt.Sprintf("%s/new-queue-1", af.BASE_URL))

	r = e.POST("/").
		WithForm(sf.GetQueueAttributesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	dupe, _ := copystructure.Copy(sf.BASE_GET_QUEUE_ATTRIBUTES_RESPONSE)
	exp3, _ := dupe.(models.GetQueueAttributesResponse)
	exp3.Result.Attrs[0].Value = "1"
	exp3.Result.Attrs[1].Value = "2"
	exp3.Result.Attrs[2].Value = "3"
	exp3.Result.Attrs[3].Value = "4"
	exp3.Result.Attrs[4].Value = "5"
	exp3.Result.Attrs[9].Value = fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName)
	exp3.Result.Attrs = append(exp3.Result.Attrs, models.Attribute{
		Name:  "RedrivePolicy",
		Value: fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, redriveQueue),
	})
	r3 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r3)
	assert.Equal(t, exp3, r3)
}

func Test_CreateQueueV1_json_with_attributes_as_ints(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	r := e.POST("/").
		WithHeaders(map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "AmazonSQS.CreateQueue",
		}).
		WithJSON(sf.CreateQueueV1RequestBodyJSON).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp1 := models.CreateQueueResult{QueueUrl: fmt.Sprintf("%s/new-queue-1", af.BASE_URL)}

	r1 := models.CreateQueueResult{}
	json.Unmarshal([]byte(r), &r1)
	assert.Equal(t, exp1, r1)

	r = e.POST("/").
		WithForm(sf.ListQueuesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp2 := models.ListQueuesResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Result:   models.ListQueuesResult{QueueUrls: []string{fmt.Sprintf("%s/new-queue-1", af.BASE_URL)}},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}
	r2 := models.ListQueuesResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Equal(t, exp2, r2)

	r = e.POST("/").
		WithForm(sf.GetQueueAttributesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	dupe, _ := copystructure.Copy(sf.BASE_GET_QUEUE_ATTRIBUTES_RESPONSE)
	exp3, _ := dupe.(models.GetQueueAttributesResponse)
	exp3.Result.Attrs[0].Value = "1"
	exp3.Result.Attrs[1].Value = "2"
	exp3.Result.Attrs[2].Value = "3"
	exp3.Result.Attrs[3].Value = "4"
	exp3.Result.Attrs[4].Value = "5"
	exp3.Result.Attrs[9].Value = fmt.Sprintf("%s:%s", af.BASE_SQS_ARN, af.QueueName)

	r3 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r3)
	assert.Equal(t, exp3, r3)
}

func Test_CreateQueueV1_json_with_attributes_ints_as_strings(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	e.POST("/").
		WithHeaders(map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "AmazonSQS.CreateQueue",
		}).
		WithJSON(sf.CreateQueueV1RequestBodyJSON).
		Expect().
		Status(http.StatusOK)

	cqr := struct {
		Version    string `json:"Version"`
		QueueName  string `json:"QueueName"`
		Attributes struct {
			DelaySeconds           string `json:"DelaySeconds"`
			MaximumMessageSize     string `json:"MaximumMessageSize"`
			MessageRetentionPeriod string `json:"MessageRetentionPeriod"`
			//Policy                        string `json:"Policy"`
			ReceiveMessageWaitTimeSeconds string `json:"ReceiveMessageWaitTimeSeconds"`
			RedrivePolicy                 struct {
				MaxReceiveCount     string `json:"maxReceiveCount"`
				DeadLetterTargetArn string `json:"deadLetterTargetArn"`
			} `json:"RedrivePolicy"`
			VisibilityTimeout string `json:"VisibilityTimeout"`
		} `json:"Attributes"`
	}{
		Version:   "2012-11-05",
		QueueName: "new-string-queue",
		Attributes: struct {
			DelaySeconds           string `json:"DelaySeconds"`
			MaximumMessageSize     string `json:"MaximumMessageSize"`
			MessageRetentionPeriod string `json:"MessageRetentionPeriod"`
			//Policy                        string `json:"Policy"`
			ReceiveMessageWaitTimeSeconds string `json:"ReceiveMessageWaitTimeSeconds"`
			RedrivePolicy                 struct {
				MaxReceiveCount     string `json:"maxReceiveCount"`
				DeadLetterTargetArn string `json:"deadLetterTargetArn"`
			} `json:"RedrivePolicy"`
			VisibilityTimeout string `json:"VisibilityTimeout"`
		}{
			DelaySeconds:           "1",
			MaximumMessageSize:     "2",
			MessageRetentionPeriod: "3",
			//Policy:                        "",
			ReceiveMessageWaitTimeSeconds: "0",
			RedrivePolicy: struct {
				MaxReceiveCount     string `json:"maxReceiveCount"`
				DeadLetterTargetArn string `json:"deadLetterTargetArn"`
			}{
				MaxReceiveCount:     "100",
				DeadLetterTargetArn: fmt.Sprintf("%s:new-queue-1", af.BASE_SQS_ARN),
			},
			VisibilityTimeout: "30"},
	}
	r := e.POST("/").
		WithHeaders(map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "AmazonSQS.CreateQueue",
		}).
		WithJSON(cqr).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp1 := models.CreateQueueResult{QueueUrl: fmt.Sprintf("%s/new-string-queue", af.BASE_URL)}

	r1 := models.CreateQueueResult{}
	json.Unmarshal([]byte(r), &r1)
	assert.Equal(t, exp1, r1)

	gqar := struct {
		Action     string `xml:"Action"`
		Attribute1 string `xml:"AttributeName.1"`
		QueueUrl   string `xml:"QueueUrl"`
	}{
		Action:     "GetQueueAttributes",
		Attribute1: "All",
		QueueUrl:   fmt.Sprintf("%s/new-string-queue", af.BASE_URL),
	}
	r = e.POST("/").
		WithForm(gqar).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	dupe, _ := copystructure.Copy(sf.BASE_GET_QUEUE_ATTRIBUTES_RESPONSE)
	exp3, _ := dupe.(models.GetQueueAttributesResponse)
	exp3.Result.Attrs[0].Value = "1"
	exp3.Result.Attrs[1].Value = "2"
	exp3.Result.Attrs[2].Value = "3"
	exp3.Result.Attrs[3].Value = "0"
	exp3.Result.Attrs[4].Value = "30"
	exp3.Result.Attrs[9].Value = fmt.Sprintf("%s:new-string-queue", af.BASE_SQS_ARN)
	exp3.Result.Attrs = append(exp3.Result.Attrs, models.Attribute{
		Name:  "RedrivePolicy",
		Value: fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, af.QueueName),
	})
	r3 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r3)
	assert.Equal(t, exp3, r3)
}

func Test_CreateQueueV1_xml_no_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	r := e.POST("/").
		WithForm(sf.CreateQueueV1RequestXML_NoAttributes).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp1 := models.CreateQueueResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Result:   models.CreateQueueResult{QueueUrl: fmt.Sprintf("%s/new-queue-1", af.BASE_URL)},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}

	r1 := models.CreateQueueResponse{}
	xml.Unmarshal([]byte(r), &r1)
	assert.Equal(t, exp1, r1)

	r = e.POST("/").
		WithForm(sf.ListQueuesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp2 := models.ListQueuesResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Result:   models.ListQueuesResult{QueueUrls: []string{fmt.Sprintf("%s/new-queue-1", af.BASE_URL)}},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}
	r2 := models.ListQueuesResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Equal(t, exp2, r2)

	r = e.POST("/").
		WithForm(sf.GetQueueAttributesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	r3 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r3)
	assert.Equal(t, sf.BASE_GET_QUEUE_ATTRIBUTES_RESPONSE, r3)
}

func Test_CreateQueueV1_xml_with_attributes(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	e.POST("/").
		WithHeaders(map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "AmazonSQS.CreateQueue",
		}).
		WithJSON(sf.CreateQueueV1RequestBodyJSON).
		Expect().
		Status(http.StatusOK)

	request := struct {
		Action    string `xml:"Action"`
		Version   string `xml:"Version"`
		QueueName string `xml:"QueueName"`
	}{
		Action:    "CreateQueue",
		Version:   "2012-11-05",
		QueueName: "new-queue-2",
	}
	r := e.POST("/").
		WithForm(request).
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
		WithFormField("Attribute.7.Value", fmt.Sprintf("{\"maxReceiveCount\": 100, \"deadLetterTargetArn\":\"%s:new-queue-1\"}", af.BASE_SQS_ARN)).
		WithFormField("Attribute.8.Name", "RedriveAllowPolicy").
		WithFormField("Attribute.8.Value", "{\"this-is\": \"the-redrive-allow-policy\"}").
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp1 := models.CreateQueueResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Result:   models.CreateQueueResult{QueueUrl: fmt.Sprintf("%s/new-queue-2", af.BASE_URL)},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}

	r1 := models.CreateQueueResponse{}
	xml.Unmarshal([]byte(r), &r1)
	assert.Equal(t, exp1, r1)

	gqar := struct {
		Action     string `xml:"Action"`
		Attribute1 string `xml:"AttributeName.1"`
		QueueUrl   string `xml:"QueueUrl"`
	}{
		Action:     "GetQueueAttributes",
		Attribute1: "All",
		QueueUrl:   fmt.Sprintf("%s/new-queue-2", af.BASE_URL),
	}
	r = e.POST("/").
		WithForm(gqar).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	dupe, _ := copystructure.Copy(sf.BASE_GET_QUEUE_ATTRIBUTES_RESPONSE)
	exp3, _ := dupe.(models.GetQueueAttributesResponse)
	exp3.Result.Attrs[0].Value = "1"
	exp3.Result.Attrs[1].Value = "2"
	exp3.Result.Attrs[2].Value = "3"
	exp3.Result.Attrs[3].Value = "4"
	exp3.Result.Attrs[4].Value = "5"
	exp3.Result.Attrs[9].Value = fmt.Sprintf("%s:new-queue-2", af.BASE_SQS_ARN)
	exp3.Result.Attrs = append(exp3.Result.Attrs, models.Attribute{
		Name:  "RedrivePolicy",
		Value: fmt.Sprintf(`{"maxReceiveCount":"100", "deadLetterTargetArn":"%s:%s"}`, af.BASE_SQS_ARN, af.QueueName),
	})
	r3 := models.GetQueueAttributesResponse{}
	xml.Unmarshal([]byte(r), &r3)
	assert.Equal(t, exp3, r3)
}
