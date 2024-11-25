package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterPolicy_IsSatisfiedBy(t *testing.T) {
	var tests = []struct {
		filterPolicy      *FilterPolicy
		messageAttributes map[string]MessageAttribute
		expected          bool
	}{
		{
			&FilterPolicy{"foo": {"bar"}},
			map[string]MessageAttribute{"foo": {DataType: "String", StringValue: "bar"}},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar", "xyz"}},
			map[string]MessageAttribute{"foo": {DataType: "String", StringValue: "xyz"}},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar", "xyz"}, "abc": {"def"}},
			map[string]MessageAttribute{"foo": {DataType: "String", StringValue: "xyz"},
				"abc": {DataType: "String", StringValue: "def"}},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar"}},
			map[string]MessageAttribute{"foo": {DataType: "String", StringValue: "baz"}},
			false,
		},
		{
			&FilterPolicy{"foo": {"bar"}},
			map[string]MessageAttribute{},
			false,
		},
		{
			&FilterPolicy{"foo": {"bar"}, "abc": {"def"}},
			map[string]MessageAttribute{"foo": {DataType: "String", StringValue: "bar"}},
			false,
		},
		{
			&FilterPolicy{"foo": {"bar"}},
			map[string]MessageAttribute{"foo": {DataType: "Binary", BinaryValue: "bar"}},
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
