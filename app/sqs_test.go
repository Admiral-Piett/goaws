package app

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestMessage_IsReadyForReceipt(t *testing.T) {
	os.Setenv("GOAWS_RANDOM_LATENCY_MIN", "100")
	os.Setenv("GOAWS_RANDOM_LATENCY_MAX", "100")
	msg := Message{
		SentTime:               time.Now(),
	}
	assert.False(t, msg.IsReadyForReceipt())
	duration, _ := time.ParseDuration("105ms")
	time.Sleep(duration)
	assert.True(t, msg.IsReadyForReceipt())
}
