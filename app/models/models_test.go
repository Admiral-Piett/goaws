package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterPolicy_IsSatisfiedBy(t *testing.T) {
	var tests = []struct {
		filterPolicy      *FilterPolicy
		messageAttributes map[string]SqsMessageAttributeValue
		expected          bool
	}{
		{
			&FilterPolicy{"foo": {"bar"}},
			map[string]SqsMessageAttributeValue{"foo": {DataType: "String", Value: "bar"}},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar", "xyz"}},
			map[string]SqsMessageAttributeValue{"foo": {DataType: "String", Value: "xyz"}},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar", "xyz"}, "abc": {"def"}},
			map[string]SqsMessageAttributeValue{"foo": {DataType: "String", Value: "xyz"},
				"abc": {DataType: "String", Value: "def"}},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar"}},
			map[string]SqsMessageAttributeValue{"foo": {DataType: "String", Value: "baz"}},
			false,
		},
		{
			&FilterPolicy{"foo": {"bar"}},
			map[string]SqsMessageAttributeValue{},
			false,
		},
		{
			&FilterPolicy{"foo": {"bar"}, "abc": {"def"}},
			map[string]SqsMessageAttributeValue{"foo": {DataType: "String", Value: "bar"}},
			false,
		},
		{
			&FilterPolicy{"foo": {"bar"}},
			map[string]SqsMessageAttributeValue{"foo": {DataType: "Binary", Value: "bar"}},
			false,
		},
	}

	for i, tt := range tests {
		actual := tt.filterPolicy.IsSatisfiedBy(tt.messageAttributes)
		if tt.filterPolicy.IsSatisfiedBy(tt.messageAttributes) != tt.expected {
			t.Errorf("#%d FilterPolicy: expected %t, actual %t", i, tt.expected, actual)
		}
	}

}

func TestMessage_IsReadyForReceipt(t *testing.T) {
	CurrentEnvironment.RandomLatency.Min = 100
	CurrentEnvironment.RandomLatency.Max = 100
	msg := SqsMessage{
		SentTime: time.Now(),
	}
	assert.False(t, msg.IsReadyForReceipt())
	duration, _ := time.ParseDuration("105ms")
	time.Sleep(duration)
	assert.True(t, msg.IsReadyForReceipt())
}
