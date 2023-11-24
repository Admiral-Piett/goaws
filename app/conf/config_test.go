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
	env := "Local"
	port := LoadYamlConfig("./mock-data/mock-config.yaml", env)
	if port[0] != "4100" {
		t.Errorf("Expected port number 4100 but got %s\n", port)
	}

	receiveWaitTime := app.SyncQueues.Queues["local-queue1"].ReceiveWaitTimeSecs
	if receiveWaitTime != 10 {
		t.Errorf("Expected local-queue1 Queue to be configured with ReceiveMessageWaitTimeSeconds: 10 but got %d\n", receiveWaitTime)
	}
	timeoutSecs := app.SyncQueues.Queues["local-queue1"].TimeoutSecs
	if timeoutSecs != 10 {
		t.Errorf("Expected local-queue1 Queue to be configured with VisibilityTimeout: 10 but got %d\n", timeoutSecs)
	}
	maximumMessageSize := app.SyncQueues.Queues["local-queue1"].MaximumMessageSize
	if maximumMessageSize != 1024 {
		t.Errorf("Expected local-queue1 Queue to be configured with MaximumMessageSize: 1024 but got %d\n", maximumMessageSize)
	}

	if app.SyncQueues.Queues["local-queue1"].DeadLetterQueue != nil {
		t.Errorf("Expected local-queue1 Queue to be configured without redrive policy\n")
	}
	if app.SyncQueues.Queues["local-queue1"].MaxReceiveCount != 0 {
		t.Errorf("Expected local-queue1 Queue to be configured without redrive policy and therefore MaxReceiveCount: 0 \n")
	}

	maxReceiveCount := app.SyncQueues.Queues["local-queue3"].MaxReceiveCount
	if maxReceiveCount != 100 {
		t.Errorf("Expected local-queue2 Queue to be configured with MaxReceiveCount: 3 from RedrivePolicy but got %d\n", maxReceiveCount)
	}
	dlq := app.SyncQueues.Queues["local-queue3"].DeadLetterQueue
	if dlq == nil {
		t.Errorf("Expected local-queue3 to have one dead letter queue to redrive to\n")
	}
	if dlq.Name != "local-queue3-dlq" {
		t.Errorf("Expected local-queue3 to have dead letter queue local-queue3-dlq but got %s\n", dlq.Name)
	}
	maximumMessageSize = app.SyncQueues.Queues["local-queue2"].MaximumMessageSize
	if maximumMessageSize != 128 {
		t.Errorf("Expected local-queue2 Queue to be configured with MaximumMessageSize: 128 but got %d\n", maximumMessageSize)
	}

	timeoutSecs = app.SyncQueues.Queues["local-queue2"].TimeoutSecs
	if timeoutSecs != 150 {
		t.Errorf("Expected local-queue2 Queue to be configured with VisibilityTimeout: 150 but got %d\n", timeoutSecs)
	}
}

func TestConfig_NoQueueAttributeDefaults(t *testing.T) {
	env := "NoQueueAttributeDefaults"
	LoadYamlConfig("./mock-data/mock-config.yaml", env)

	receiveWaitTime := app.SyncQueues.Queues["local-queue1"].ReceiveWaitTimeSecs
	if receiveWaitTime != 0 {
		t.Errorf("Expected local-queue1 Queue to be configured with ReceiveMessageWaitTimeSeconds: 0 but got %d\n", receiveWaitTime)
	}
	timeoutSecs := app.SyncQueues.Queues["local-queue1"].TimeoutSecs
	if timeoutSecs != 30 {
		t.Errorf("Expected local-queue1 Queue to be configured with VisibilityTimeout: 30 but got %d\n", timeoutSecs)
	}

	receiveWaitTime = app.SyncQueues.Queues["local-queue2"].ReceiveWaitTimeSecs
	if receiveWaitTime != 20 {
		t.Errorf("Expected local-queue2 Queue to be configured with ReceiveMessageWaitTimeSeconds: 20 but got %d\n", receiveWaitTime)
	}
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
