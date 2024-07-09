package smoke_tests

import (
	"context"
	"net/http"
	"testing"

	"encoding/xml"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/stretchr/testify/assert"

	"github.com/gavv/httpexpect/v2"
)

func Test_Delete_Topic_json_success(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	topicName1 := "topic-1"
	resp, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName1),
	})

	topicName2 := "topic-2"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName2),
	})

	_, err := snsClient.DeleteTopic(context.TODO(), &sns.DeleteTopicInput{
		TopicArn: resp.TopicArn,
	})

	assert.Nil(t, err)

	app.SyncQueues.Lock()

	defer app.SyncQueues.Unlock()

	topics := app.SyncTopics.Topics
	assert.Len(t, topics, 1)

	_, ok := topics[topicName1]
	assert.False(t, ok)

	_, ok = topics[topicName2]
	assert.True(t, ok)
}

func Test_Delete_Topic_json_NotFound(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	_, err := snsClient.DeleteTopic(context.TODO(), &sns.DeleteTopicInput{
		TopicArn: aws.String("asdf"),
	})

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "SimpleNotificationService.NonExistentTopic")

	app.SyncQueues.Lock()
	defer app.SyncQueues.Unlock()

	topics := app.SyncTopics.Topics
	assert.Len(t, topics, 0)
}

func Test_Delete_Topic_xml_success(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	topicName1 := "topic-1"
	resp, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName1),
	})

	topicName2 := "topic-2"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName2),
	})

	requestBody := struct {
		Action   string `schema:"DeleteTopic"`
		TopicArn string `schema:"TopicArn"`
	}{
		Action:   "DeleteTopic",
		TopicArn: *resp.TopicArn,
	}

	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	deleteTopicResponseObject := models.DeleteTopicResponse{}
	xml.Unmarshal([]byte(r), &deleteTopicResponseObject)

	assert.Equal(t, "http://queue.amazonaws.com/doc/2012-11-05/", deleteTopicResponseObject.Xmlns)
	app.SyncQueues.Lock()

	defer app.SyncQueues.Unlock()

	topics := app.SyncTopics.Topics
	assert.Len(t, topics, 1)

	_, ok := topics[topicName1]
	assert.False(t, ok)

	_, ok = topics[topicName2]
	assert.True(t, ok)

}

func Test_Delete_Topic_xml_NotFound(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	requestBody := struct {
		Action   string `schema:"DeleteTopic"`
		TopicArn string `schema:"TopicArn"`
	}{
		Action:   "DeleteTopic",
		TopicArn: "asdf",
	}

	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusBadRequest).
		Body().Raw()

	deleteTopicErrorResponseObject := models.ErrorResponse{}
	xml.Unmarshal([]byte(r), &deleteTopicErrorResponseObject)

	assert.Equal(t, deleteTopicErrorResponseObject.Result.Type, "Not Found")
}
