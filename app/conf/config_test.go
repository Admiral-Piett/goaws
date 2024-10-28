package conf

import (
	"testing"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/stretchr/testify/assert"
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
	numQueues = len(models.SyncQueues.Queues)
	if numQueues != 0 {
		t.Errorf("Expected zero queues to be in the sqs topics but got %d\n", numQueues)
	}

	numTopics := len(envs[env].Topics)
	if numTopics != 0 {
		t.Errorf("Expected zero topics to be in the environment but got %d\n", numTopics)
	}
	numTopics = len(models.SyncTopics.Topics)
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
	numQueues = len(models.SyncQueues.Queues)
	if numQueues != 6 {
		t.Errorf("Expected five queues to be in the sqs topics but got %d\n", numQueues)
	}

	numTopics := len(envs[env].Topics)
	if numTopics != 2 {
		t.Errorf("Expected two topics to be in the environment but got %d\n", numTopics)
	}
	numTopics = len(models.SyncTopics.Topics)
	if numTopics != 2 {
		t.Errorf("Expected two topics to be in the sns topics but got %d\n", numTopics)
	}
}

func TestConfig_QueueAttributes(t *testing.T) {
	var emptyQueue *models.Queue
	env := "Local"
	port := LoadYamlConfig("./mock-data/mock-config.yaml", env)
	if port[0] != "4100" {
		t.Errorf("Expected port number 4100 but got %s\n", port)
	}

	assert.Equal(t, 10, models.SyncQueues.Queues["local-queue1"].ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 10, models.SyncQueues.Queues["local-queue1"].VisibilityTimeout)
	assert.Equal(t, 1024, models.SyncQueues.Queues["local-queue1"].MaximumMessageSize)
	assert.Equal(t, emptyQueue, models.SyncQueues.Queues["local-queue1"].DeadLetterQueue)
	assert.Equal(t, 0, models.SyncQueues.Queues["local-queue1"].MaxReceiveCount)
	assert.Equal(t, 345600, models.SyncQueues.Queues["local-queue1"].MessageRetentionPeriod)
	assert.Equal(t, 100, models.SyncQueues.Queues["local-queue3"].MaxReceiveCount)

	assert.Equal(t, "local-queue3-dlq", models.SyncQueues.Queues["local-queue3"].DeadLetterQueue.Name)
	assert.Equal(t, 128, models.SyncQueues.Queues["local-queue2"].MaximumMessageSize)
	assert.Equal(t, 150, models.SyncQueues.Queues["local-queue2"].VisibilityTimeout)
	assert.Equal(t, 245600, models.SyncQueues.Queues["local-queue2"].MessageRetentionPeriod)
}

func TestConfig_NoQueueAttributeDefaults(t *testing.T) {
	env := "NoQueueAttributeDefaults"
	LoadYamlConfig("./mock-data/mock-config.yaml", env)

	receiveWaitTime := models.SyncQueues.Queues["local-queue1"].ReceiveMessageWaitTimeSeconds
	if receiveWaitTime != 0 {
		t.Errorf("Expected local-queue1 Queue to be configured with ReceiveMessageWaitTimeSeconds: 0 but got %d\n", receiveWaitTime)
	}
	timeoutSecs := models.SyncQueues.Queues["local-queue1"].VisibilityTimeout
	if timeoutSecs != 30 {
		t.Errorf("Expected local-queue1 Queue to be configured with VisibilityTimeout: 30 but got %d\n", timeoutSecs)
	}

	receiveWaitTime = models.SyncQueues.Queues["local-queue2"].ReceiveMessageWaitTimeSeconds
	if receiveWaitTime != 20 {
		t.Errorf("Expected local-queue2 Queue to be configured with ReceiveMessageWaitTimeSeconds: 20 but got %d\n", receiveWaitTime)
	}

	messageRetentionPeriod := models.SyncQueues.Queues["local-queue1"].MessageRetentionPeriod
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

	assert.Equal(t, 262144, models.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize)
	assert.Equal(t, 345600, models.CurrentEnvironment.QueueAttributeDefaults.MessageRetentionPeriod)
	assert.Equal(t, 0, models.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, 30, models.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout)
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

	queues := models.SyncQueues.Queues
	topics := models.SyncTopics.Topics
	for _, expectedName := range expectedQueues {
		_, ok := queues[expectedName]
		assert.True(t, ok)
	}
	for _, expectedName := range expectedTopics {
		_, ok := topics[expectedName]
		assert.True(t, ok)
	}
}

func TestConfig_LoadYamlConfig_missing_config_loads_nothing(t *testing.T) {
	models.CurrentEnvironment = models.Environment{}
	ports := LoadYamlConfig("/garbage", "Local")

	assert.Equal(t, []string{"4100"}, ports)
	assert.Equal(t, models.CurrentEnvironment, models.Environment{})
}

func TestConfig_LoadYamlConfig_invalid_config_loads_nothing(t *testing.T) {
	models.CurrentEnvironment = models.Environment{}
	ports := LoadYamlConfig("../common/common.go", "Local")

	assert.Equal(t, []string{"4100"}, ports)
	assert.Equal(t, models.CurrentEnvironment, models.Environment{})
}
