package models

import (
	"time"
)

var BaseXmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
var BaseResponseMetadata = ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}

var DeduplicationPeriod = 5 * time.Minute

var AvailableQueueAttributes = map[string]bool{
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

const (
	ProtocolSQS     Protocol = "sqs"
	ProtocolHTTP    Protocol = "http"
	ProtocolHTTPS   Protocol = "https"
	ProtocolDefault Protocol = "default"
)

const (
	MessageStructureJSON MessageStructure = "json"
)
