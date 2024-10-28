package smoke_tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/gavv/httpexpect/v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"
)

func Test_Unsubscribe_json(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	subArn := models.SyncTopics.Topics["unit-topic1"].Subscriptions[0].SubscriptionArn
	response, err := snsClient.Unsubscribe(context.TODO(), &sns.UnsubscribeInput{
		SubscriptionArn: &subArn,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	models.SyncTopics.Lock()
	defer models.SyncTopics.Unlock()

	subscriptions := models.SyncTopics.Topics["unit-topic1"].Subscriptions
	assert.Len(t, subscriptions, 0)
}

func Test_Unsubscribe_xml(t *testing.T) {
	server := generateServer()
	defaultEnv := models.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		models.ResetResources()
		models.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	subArn := models.SyncTopics.Topics["unit-topic1"].Subscriptions[0].SubscriptionArn
	requestBody := struct {
		Action          string `xml:"Action"`
		SubscriptionArn string `schema:"SubscriptionArn"`
	}{
		Action:          "Unsubscribe",
		SubscriptionArn: subArn,
	}

	e.POST("/").
		WithForm(requestBody).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	subscriptions := models.SyncTopics.Topics["unit-topic1"].Subscriptions
	assert.Len(t, subscriptions, 0)
}
