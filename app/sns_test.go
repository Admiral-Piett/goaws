package app

import "testing"

func TestFilterPolicy_IsSatisfiedBy(t *testing.T) {
	var tests = []struct {
		filterPolicy      *FilterPolicy
		messageAttributes *TopicMessageAttributes
		expected          bool
	}{
		{
			&FilterPolicy{"foo": {"bar"}},
			&TopicMessageAttributes{"foo": "bar"},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar", "xyz"}},
			&TopicMessageAttributes{"foo": "xyz"},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar", "xyz"}, "abc": {"def"}},
			&TopicMessageAttributes{"foo": "xyz", "abc": "def"},
			true,
		},
		{
			&FilterPolicy{"foo": {"bar"}},
			&TopicMessageAttributes{"foo": "baz"},
			false,
		},
		{
			&FilterPolicy{"foo": {"bar"}},
			&TopicMessageAttributes{},
			false,
		},
		{
			&FilterPolicy{"foo": {"bar"}, "abc": {"def"}},
			&TopicMessageAttributes{"foo": "bar"},
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
