package fixtures

import (
	"fmt"

	"github.com/Admiral-Piett/goaws/app/models"

	af "github.com/Admiral-Piett/goaws/app/fixtures"

	"github.com/Admiral-Piett/goaws/app"
)

var BASE_GET_QUEUE_ATTRIBUTES_RESPONSE = models.GetQueueAttributesResponse{
	Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/",
	Result: models.GetQueueAttributesResult{Attrs: []models.Attribute{
		{
			Name:  "DelaySeconds",
			Value: "0",
		},
		{
			Name:  "MaximumMessageSize",
			Value: "0",
		},
		{
			Name:  "MessageRetentionPeriod",
			Value: "0",
		},
		{
			Name:  "ReceiveMessageWaitTimeSeconds",
			Value: "0",
		},
		{
			Name:  "VisibilityTimeout",
			Value: "0",
		},
		{
			Name:  "ApproximateNumberOfMessages",
			Value: "0",
		},
		{
			Name:  "ApproximateNumberOfMessagesNotVisible",
			Value: "0",
		},
		{
			Name:  "CreatedTimestamp",
			Value: "0000000000",
		},
		{
			Name:  "LastModifiedTimestamp",
			Value: "0000000000",
		},
		{
			Name:  "QueueArn",
			Value: fmt.Sprintf("%s:new-queue-1", af.BASE_SQS_ARN),
		},
	}},
	Metadata: app.ResponseMetadata{RequestId: REQUEST_ID},
}
