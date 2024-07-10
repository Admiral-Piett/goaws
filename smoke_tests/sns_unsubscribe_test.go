package smoke_tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/gavv/httpexpect/v2"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/assert"
)

func Test_Unsubscribe_json(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
	sdkConfig.BaseEndpoint = aws.String(server.URL)
	snsClient := sns.NewFromConfig(sdkConfig)

	subArn := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0].SubscriptionArn
	response, err := snsClient.Unsubscribe(context.TODO(), &sns.UnsubscribeInput{
		SubscriptionArn: &subArn,
	})

	assert.Nil(t, err)
	assert.NotNil(t, response)

	app.SyncTopics.Lock()
	defer app.SyncTopics.Unlock()

	subscriptions := app.SyncTopics.Topics["unit-topic1"].Subscriptions
	assert.Len(t, subscriptions, 0)
}

func Test_Unsubscribe_xml(t *testing.T) {
	server := generateServer()
	defaultEnv := app.CurrentEnvironment
	conf.LoadYamlConfig("../app/conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		server.Close()
		test.ResetResources()
		app.CurrentEnvironment = defaultEnv
	}()

	e := httpexpect.Default(t, server.URL)

	subArn := app.SyncTopics.Topics["unit-topic1"].Subscriptions[0].SubscriptionArn
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

	subscriptions := app.SyncTopics.Topics["unit-topic1"].Subscriptions
	assert.Len(t, subscriptions, 0)
}
