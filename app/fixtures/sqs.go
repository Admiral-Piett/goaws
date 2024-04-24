package fixtures

import (
	"fmt"

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

var CreateQueueResult = models.CreateQueueResult{
	QueueUrl: fmt.Sprintf("http://us-east-1.localhost:4200/100010001000/%s", QueueName),
}

var CreateQueueResponse = models.CreateQueueResponse{
	Xmlns:    models.BASE_XMLNS,
	Result:   CreateQueueResult,
	Metadata: models.BASE_RESPONSE_METADATA,
}

var ListQueuesResult = models.ListQueuesResult{
	QueueUrls: []string{
		fmt.Sprintf("%s/%s", BASE_URL, "unit-queue1"),
		fmt.Sprintf("%s/%s", BASE_URL, "unit-queue2"),
	},
}

var ListQueuesResponse = models.ListQueuesResponse{
	Xmlns:    models.BASE_XMLNS,
	Result:   ListQueuesResult,
	Metadata: models.BASE_RESPONSE_METADATA,
}
