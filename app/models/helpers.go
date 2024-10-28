package models

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// ---- Unit Tests ----
func ResetApp() {
	CurrentEnvironment = Environment{}
	ResetResources()
}

func ResetResources() {
	SyncQueues.Lock()
	SyncQueues.Queues = make(map[string]*Queue)
	SyncQueues.Unlock()
	SyncTopics.Lock()
	SyncTopics.Topics = make(map[string]*Topic)
	SyncTopics.Unlock()
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func generateRandomLatency() (time.Duration, error) {
	min := CurrentEnvironment.RandomLatency.Min
	max := CurrentEnvironment.RandomLatency.Max
	if min == 0 && max == 0 {
		return time.Duration(0), nil
	}
	var randomLatencyValue int
	if max == min {
		randomLatencyValue = max
	} else {
		randomLatencyValue = rand.Intn(max-min) + min
	}
	randomDuration, err := time.ParseDuration(fmt.Sprintf("%dms", randomLatencyValue))
	if err != nil {
		return time.Duration(0), errors.New(fmt.Sprintf("Error parsing random latency value: %dms", randomLatencyValue))
	}
	return randomDuration, nil
}
