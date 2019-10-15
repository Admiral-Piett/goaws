package conf

import (
	"testing"

	"github.com/p4tin/goaws/app"
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
	if numQueues != 3 {
		t.Errorf("Expected three queues to be in the environment but got %d\n", numQueues)
	}
	numQueues = len(app.SyncQueues.Queues)
	if numQueues != 5 {
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

	receiveWaitTime = app.SyncQueues.Queues["local-queue2"].ReceiveWaitTimeSecs
	if receiveWaitTime != 20 {
		t.Errorf("Expected local-queue2 Queue to be configured with ReceiveMessageWaitTimeSeconds: 20 but got %d\n", receiveWaitTime)
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
