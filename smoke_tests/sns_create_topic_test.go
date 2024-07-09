package smoke_tests

import (
	"context"
	"encoding/xml"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func Test_CreateTopicV1_json_success(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	// Target test
	topicName := "new-topic-1"
	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sdkResponse, err := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})

	// Should success
	assert.Contains(t, *sdkResponse.TopicArn, topicName)
	assert.Nil(t, err)

	// Get created topic
	listTopicsXML := struct {
		Action  string `xml:"Action"`
		Version string `xml:"Version"`
	}{
		Action:  "ListTopics",
		Version: "2012-11-05",
	}
	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(listTopicsXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r2 := models.ListTopicsResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Equal(t, 1, len(r2.Result.Topics.Member))
	assert.Contains(t, r2.Result.Topics.Member[0].TopicArn, topicName)
}

func Test_CreateTopicV1_json_existant_topic(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	// Prepare existant topic
	topicName := "new-topic-1"
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sdkResponse, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})
	assert.Contains(t, *sdkResponse.TopicArn, topicName)
	assert.Nil(t, err)

	// Target test: create topic with same name
	sdkResponse, err = snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})

	// Should success
	assert.Contains(t, *sdkResponse.TopicArn, topicName)
	assert.Nil(t, err)

	// Topic should not be duplicated
	listTopicsXML := struct {
		Action  string `xml:"Action"`
		Version string `xml:"Version"`
	}{
		Action:  "ListTopics",
		Version: "2012-11-05",
	}
	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(listTopicsXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r2 := models.ListTopicsResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Equal(t, 1, len(r2.Result.Topics.Member))
	assert.Contains(t, r2.Result.Topics.Member[0].TopicArn, topicName)
}

func Test_CreateTopicV1_json_add_multiple_topics(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	// Prepare existant topic
	topicName := "new-topic-1"
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)
	sdkResponse, _ := snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName,
	})
	assert.Contains(t, *sdkResponse.TopicArn, topicName)
	assert.Nil(t, err)

	// Target test: create topic with different name
	topicName2 := "new-topic-2"
	sdkResponse, err = snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: &topicName2,
	})

	// Should success
	assert.Contains(t, *sdkResponse.TopicArn, topicName2)
	assert.Nil(t, err)

	// Number of topic should be 2
	listTopicsXML := struct {
		Action  string `xml:"Action"`
		Version string `xml:"Version"`
	}{
		Action:  "ListTopics",
		Version: "2012-11-05",
	}
	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(listTopicsXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r2 := models.ListTopicsResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Equal(t, 2, len(r2.Result.Topics.Member))
}

func Test_CreateTopicV1_xml_success(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	// Target test
	topicName := "new-topic-1"
	createTopicsXML := struct {
		Action  string `xml:"Action"`
		Version string `xml:"Version"`
		Name    string `xml:"Name"`
	}{
		Action:  "CreateTopic",
		Version: "2012-11-05",
		Name:    topicName,
	}
	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(createTopicsXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r2 := models.CreateTopicResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Contains(t, r2.Result.TopicArn, topicName)

	// Get created topic
	listTopicsXML := struct {
		Action  string `xml:"Action"`
		Version string `xml:"Version"`
	}{
		Action:  "ListTopics",
		Version: "2012-11-05",
	}
	r = e.POST("/").
		WithForm(listTopicsXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r3 := models.ListTopicsResponse{}
	xml.Unmarshal([]byte(r), &r3)
	assert.Equal(t, 1, len(r3.Result.Topics.Member))
	assert.Contains(t, r3.Result.Topics.Member[0].TopicArn, topicName)
}

func Test_CreateTopicV1_xml_existant_topic(t *testing.T) {
	server := generateServer()
	defer func() {
		server.Close()
		test.ResetResources()
	}()

	topicName := "new-topic-1"

	// Prepare existant topic
	createTopicsXML := struct {
		Action  string `xml:"Action"`
		Version string `xml:"Version"`
		Name    string `xml:"Name"`
	}{
		Action:  "CreateTopic",
		Version: "2012-11-05",
		Name:    topicName,
	}
	e := httpexpect.Default(t, server.URL)
	r := e.POST("/").
		WithForm(createTopicsXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r2 := models.CreateTopicResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Contains(t, r2.Result.TopicArn, topicName)

	// Target test: create topic with same name
	r = e.POST("/").
		WithForm(createTopicsXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r2 = models.CreateTopicResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Contains(t, r2.Result.TopicArn, topicName)

	// Topic should not be duplicated
	listTopicsXML := struct {
		Action  string `xml:"Action"`
		Version string `xml:"Version"`
	}{
		Action:  "ListTopics",
		Version: "2012-11-05",
	}
	r = e.POST("/").
		WithForm(listTopicsXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	r3 := models.ListTopicsResponse{}
	xml.Unmarshal([]byte(r), &r3)
	assert.Equal(t, 1, len(r3.Result.Topics.Member))
	assert.Contains(t, r3.Result.Topics.Member[0].TopicArn, topicName)
}
