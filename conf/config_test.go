package conf

import (
	"testing"
	"github.com/p4tin/goaws/gosns"
	"github.com/p4tin/goaws/gosqs"
)

func TestConfig_NoQueuesOrTopics(t *testing.T) {
	env := "NoQueuesOrTopics"
	port :=LoadYamlConfig("./mock-data/mock-config.yaml", env, "4200")
	if port != "4200" {
		t.Errorf("Expected port number 4200 but got %s\n", port)
	}

	numQueues := len(envs[env].Queues)
	if numQueues != 0 {
		t.Errorf("Expected zero queues to be in the environment but got %s\n", numQueues)
	}
	numQueues = len(gosqs.SyncQueues.Queues)
	if numQueues != 0 {
		t.Errorf("Expected zero queues to be in the sqs topics but got %s\n", numQueues)
	}


	numTopics := len(envs[env].Topics)
	if numTopics != 0 {
		t.Errorf("Expected zero topics to be in the environment but got %s\n", numTopics)
	}
	numTopics = len(gosns.SyncTopics.Topics)
	if numTopics != 0 {
		t.Errorf("Expected zero topics to be in the sns topics but got %s\n", numTopics)
	}
}

func TestConfig_CreateQueuesTopicsAndSubscriptions(t *testing.T) {
	env := "Local"
	port := LoadYamlConfig("./mock-data/mock-config.yaml", env, "")
	if port != "4100" {
		t.Errorf("Expected port number 4100 but got %s\n", port)
	}

	numQueues := len(envs[env].Queues)
	if numQueues != 3 {
		t.Errorf("Expected three queues to be in the environment but got %s\n", numQueues)
	}
	numQueues = len(gosqs.SyncQueues.Queues)
	if numQueues != 5 {
		t.Errorf("Expected five queues to be in the sqs topics but got %s\n", numQueues)
	}


	numTopics := len(envs[env].Topics)
	if numTopics != 2 {
		t.Errorf("Expected two topics to be in the environment but got %s\n", numTopics)
	}
	numTopics = len(gosns.SyncTopics.Topics)
	if numTopics != 2 {
		t.Errorf("Expected two topics to be in the sns topics but got %s\n", numTopics)
	}

	numSubscriptions := 2
	if numSubscriptions != 2 {
		t.Errorf("Expected two Subscriptions to be in the environment but got %s\n", numTopics)
	}
}

