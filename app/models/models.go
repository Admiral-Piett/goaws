package models

import (
	"github.com/Admiral-Piett/goaws/app"
)

var BASE_XMLNS = "http://queue.amazonaws.com/doc/2012-11-05/"
var BASE_RESPONSE_METADATA = app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}

var AVAILABLE_QUEUE_ATTRIBUTES = map[string]bool{
	"DelaySeconds":                          true,
	"MaximumMessageSize":                    true,
	"MessageRetentionPeriod":                true,
	"Policy":                                true,
	"ReceiveMessageWaitTimeSeconds":         true,
	"VisibilityTimeout":                     true,
	"RedrivePolicy":                         true,
	"RedriveAllowPolicy":                    true,
	"ApproximateNumberOfMessages":           true,
	"ApproximateNumberOfMessagesDelayed":    true,
	"ApproximateNumberOfMessagesNotVisible": true,
	"CreatedTimestamp":                      true,
	"LastModifiedTimestamp":                 true,
	"QueueArn":                              true,
}
