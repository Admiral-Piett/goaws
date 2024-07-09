package smoke_tests

import (
	"context"
	"net/http"
	"testing"

	"encoding/xml"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/stretchr/testify/assert"

	sf "github.com/Admiral-Piett/goaws/smoke_tests/fixtures"

	"github.com/gavv/httpexpect/v2"
)

func Test_List_Topics_json_no_topics(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	sdkResponse, err := snsClient.ListTopics(context.TODO(), &sns.ListTopicsInput{})

	assert.Nil(t, err)
	assert.Len(t, sdkResponse.Topics, 0)
}

func Test_List_Topics_json_multiple_topics(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	topicName1 := "topic-1"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName1),
	})

	topicName2 := "topic-2"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName2),
	})

	sdkResponse, err := snsClient.ListTopics(context.TODO(), &sns.ListTopicsInput{})

	assert.Nil(t, err)
	assert.Len(t, sdkResponse.Topics, 2)
	assert.NotEqual(t, sdkResponse.Topics[0].TopicArn, sdkResponse.Topics[1].TopicArn)
}

func Test_List_Topics_xml_no_topics(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	e := httpexpect.Default(t, server.URL)

	r := e.POST("/").
		WithForm(sf.ListTopicsRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	listTopicsResponseObject := models.ListTopicsResponse{}
	xml.Unmarshal([]byte(r), &listTopicsResponseObject)

	assert.Equal(t, "http://queue.amazonaws.com/doc/2012-11-05/", listTopicsResponseObject.Xmlns)
	assert.Len(t, listTopicsResponseObject.Result.Topics.Member, 0)
}

func Test_ListTopics_xml_multiple_topics(t *testing.T) {
	server := generateServer()

	defer func() {
		server.Close()
		test.ResetResources()
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	topicName1 := "topic-1"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName1),
	})

	topicName2 := "topic-2"
	snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(topicName2),
	})

	e := httpexpect.Default(t, server.URL)

	r := e.POST("/").
		WithForm(sf.ListTopicsRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	listTopicsResponseObject := models.ListTopicsResponse{}
	xml.Unmarshal([]byte(r), &listTopicsResponseObject)

	assert.Equal(t, "http://queue.amazonaws.com/doc/2012-11-05/", listTopicsResponseObject.Xmlns)
	assert.Len(t, listTopicsResponseObject.Result.Topics.Member, 2)
	assert.NotEqual(t, listTopicsResponseObject.Result.Topics.Member[0].TopicArn, listTopicsResponseObject.Result.Topics.Member[1].TopicArn)
}
