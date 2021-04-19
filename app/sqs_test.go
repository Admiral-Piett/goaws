package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessage_IsReadyForReceipt(t *testing.T) {
	CurrentEnvironment.RandomLatency.Min = 100
	CurrentEnvironment.RandomLatency.Max = 100
	msg := Message{
		SentTime: time.Now(),
	}
	assert.False(t, msg.IsReadyForReceipt())
	duration, _ := time.ParseDuration("105ms")
	time.Sleep(duration)
	assert.True(t, msg.IsReadyForReceipt())
}
