package fixtures

import (
	"fmt"

	"github.com/Admiral-Piett/goaws/app"

	"github.com/Admiral-Piett/goaws/app/models"
)

var QueueName = "new-queue-1"
var DeadLetterQueueName = "dead-letter-queue-1"

var CreateQueueRequest = models.CreateQueueRequest{
	QueueName:  QueueName,
	Attributes: CreateQueueAttributes,
	Tags:       map[string]string{"my": "tag"},
}

var CreateQueueAttributes = models.Attributes{
	DelaySeconds:                  1,
	MaximumMessageSize:            2,
	MessageRetentionPeriod:        3,
	Policy:                        map[string]interface{}{"this-is": "the-policy"}, //IAM Policy
	ReceiveMessageWaitTimeSeconds: 4,
	VisibilityTimeout:             5,
	//RedrivePolicy: models.RedrivePolicy{
	//	MaxReceiveCount:     100,
	//	DeadLetterTargetArn: fmt.Sprintf("arn:aws:sqs:us-east-1:100010001000:%s", DeadLetterQueueName),
	//},
	RedriveAllowPolicy: map[string]interface{}{"this-is": "the-redrive-allow-policy"},
}

var CreateQueueResponse = app.CreateQueueResponse{
	Xmlns: "http://queue.amazonaws.com/doc/2012-11-05/",
	Result: app.CreateQueueResult{
		QueueUrl: fmt.Sprintf("http://us-east-1.localhost:4200/100010001000/%s", QueueName),
	},
	Metadata: app.ResponseMetadata{
		RequestId: "00000000-0000-0000-0000-000000000000",
	},
}
