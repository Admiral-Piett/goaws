package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Admiral-Piett/goaws/app"
)

func TestConfig_NoQueuesOrTopics(t *testing.T) {
	env := "NoQueuesOrTopics"
	port := LoadYamlConfig("./mock-data/mock-config.yaml", env)
	if port[0] != "4100" {
		t.Errorf("Expected port number 4200 but got %s\n", port)
	}

	numQueues := len(envs[env].Queues)
	if numQueues != 0 {
		t.Errorf("Expected zero queues to be in the environment but got %d\n", numQueues)
	}
	numQueues = len(app.SyncQueues.Queues)
	if numQueues != 0 {
		t.Errorf("Expected zero queues to be in the sqs topics but got %d\n", numQueues)
	}

	numTopics := len(envs[env].Topics)
	if numTopics != 0 {
		t.Errorf("Expected zero topics to be in the environment but got %d\n", numTopics)
	}
	numTopics = len(app.SyncTopics.Topics)
	if numTopics != 0 {
		t.Errorf("Expected zero topics to be in the sns topics but got %d\n", numTopics)
	}
}

func TestConfig_CreateQueuesTopicsAndSubscriptions(t *testing.T) {
	env := "Local"
	port := LoadYamlConfig("./mock-data/mock-config.yaml", env)
	if port[0] != "4100" {
		t.Errorf("Expected port number 4100 but got %s\n", port)
	}

	numQueues := len(envs[env].Queues)
	if numQueues != 4 {
		t.Errorf("Expected three queues to be in the environment but got %d\n", numQueues)
	}
	numQueues = len(app.SyncQueues.Queues)
	if numQueues != 6 {
		t.Errorf("Expected five queues to be in the sqs topics but got %d\n", numQueues)
	}

	numTopics := len(envs[env].Topics)
	if numTopics != 2 {
		t.Errorf("Expected two topics to be in the environment but got %d\n", numTopics)
	}
	numTopics = len(app.SyncTopics.Topics)
	if numTopics != 2 {
		t.Errorf("Expected two topics to be in the sns topics but got %d\n", numTopics)
	}
}

func TestConfig_QueueAttributes(t *testing.T) {
	var emptyQueue *app.Queue
	env := "Local"
	port := LoadYamlConfig("./mock-data/mock-config.yaml", env)
	if port[0] != "4100" {
		t.Errorf("Expected port number 4100 but got %s\n", port)
	}

	assert.Equal(t, 10, app.SyncQueues.Queues["local-queue1"].ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 10, app.SyncQueues.Queues["local-queue1"].VisibilityTimeout)
	assert.Equal(t, 1024, app.SyncQueues.Queues["local-queue1"].MaximumMessageSize)
	assert.Equal(t, emptyQueue, app.SyncQueues.Queues["local-queue1"].DeadLetterQueue)
	assert.Equal(t, 0, app.SyncQueues.Queues["local-queue1"].MaxReceiveCount)
	assert.Equal(t, 445600, app.SyncQueues.Queues["local-queue1"].MessageRetentionPeriod)
	assert.Equal(t, 100, app.SyncQueues.Queues["local-queue3"].MaxReceiveCount)

	assert.Equal(t, "local-queue3-dlq", app.SyncQueues.Queues["local-queue3"].DeadLetterQueue.Name)
	assert.Equal(t, 128, app.SyncQueues.Queues["local-queue2"].MaximumMessageSize)
	assert.Equal(t, 150, app.SyncQueues.Queues["local-queue2"].VisibilityTimeout)
	assert.Equal(t, 245600, app.SyncQueues.Queues["local-queue2"].MessageRetentionPeriod)
}

func TestConfig_NoQueueAttributeDefaults(t *testing.T) {
	env := "NoQueueAttributeDefaults"
	LoadYamlConfig("./mock-data/mock-config.yaml", env)

	receiveWaitTime := app.SyncQueues.Queues["local-queue1"].ReceiveMessageWaitTimeSeconds
	if receiveWaitTime != 0 {
		t.Errorf("Expected local-queue1 Queue to be configured with ReceiveMessageWaitTimeSeconds: 0 but got %d\n", receiveWaitTime)
	}
	timeoutSecs := app.SyncQueues.Queues["local-queue1"].VisibilityTimeout
	if timeoutSecs != 30 {
		t.Errorf("Expected local-queue1 Queue to be configured with VisibilityTimeout: 30 but got %d\n", timeoutSecs)
	}

	receiveWaitTime = app.SyncQueues.Queues["local-queue2"].ReceiveMessageWaitTimeSeconds
	if receiveWaitTime != 20 {
		t.Errorf("Expected local-queue2 Queue to be configured with ReceiveMessageWaitTimeSeconds: 20 but got %d\n", receiveWaitTime)
	}

	messageRetentionPeriod := app.SyncQueues.Queues["local-queue1"].MessageRetentionPeriod
	if messageRetentionPeriod != 345600 {
		t.Errorf("Expected local-queue2 Queue to be configured with VisibilityTimeout: 150 but got %d\n", timeoutSecs)
	}
}

func TestConfig_invalid_config_resorts_to_default_queue_attributes(t *testing.T) {
	env := "missing"
	port := LoadYamlConfig("./mock-data/mock-config.yaml", env)
	if port[0] != "4100" {
		t.Errorf("Expected port number 4100 but got %s\n", port)
	}

	assert.Equal(t, 262144, app.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize)
	assert.Equal(t, 345600, app.CurrentEnvironment.QueueAttributeDefaults.MessageRetentionPeriod)
	assert.Equal(t, 0, app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 30, app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout)
}

func TestConfig_LoadYamlConfig_finds_default_config(t *testing.T) {
	expectedQueues := []string{
		"local-queue1",
		"local-queue2",
		"local-queue3",
		"local-queue3-dlq",
		"local-queue4",
	}
	expectedTopics := []string{
		"local-topic1",
		"local-topic2",
		"local-topic3",
		"local-topic4",
	}

	env := "Local"
	LoadYamlConfig("", env)

	queues := app.SyncQueues.Queues
	topics := app.SyncTopics.Topics
	for _, expectedName := range expectedQueues {
		_, ok := queues[expectedName]
		assert.True(t, ok)
	}
	for _, expectedName := range expectedTopics {
		_, ok := topics[expectedName]
		assert.True(t, ok)
	}
}
