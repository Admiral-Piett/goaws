package fixtures

import (
	"fmt"
	"time"

	"github.com/Admiral-Piett/goaws/app"

	"github.com/Admiral-Piett/goaws/app/models"
)

var QueueName = "new-queue-1"
var QueueUrl = fmt.Sprintf("%s/%s", BASE_URL, QueueName)
var DeadLetterQueueName = "dead-letter-queue-1"

var FullyPopulatedQueue = &app.Queue{
	Name: QueueName,
	URL: fmt.Sprintf("http://%s.%s:%s/%s/%s",
		LOCAL_ENVIRONMENT.Region,
		LOCAL_ENVIRONMENT.Host,
		LOCAL_ENVIRONMENT.Port,
		LOCAL_ENVIRONMENT.AccountID,
		QueueName,
	),
	Arn: fmt.Sprintf("arn:aws:sqs:%s:%s:%s",
		LOCAL_ENVIRONMENT.Region,
		LOCAL_ENVIRONMENT.AccountID,
		QueueName,
	),
	VisibilityTimeout:             5,
	ReceiveMessageWaitTimeSeconds: 4,
	DelaySeconds:                  1,
	MaximumMessageSize:            2,
	MessageRetentionPeriod:        3,
	Duplicates:                    make(map[string]time.Time),
}

var CreateQueueRequest = models.CreateQueueRequest{
	QueueName:  QueueName,
	Attributes: QueueAttributes,
	Tags:       map[string]string{"my": "tag"},
}

var QueueAttributes = models.QueueAttributes{
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

var GetQueueAttributesRequest = models.GetQueueAttributesRequest{
	QueueUrl:       fmt.Sprintf("%s/unit-queue1", BASE_URL),
	AttributeNames: []string{"All"},
}

var GetQueueAttributesResponse = models.GetQueueAttributesResponse{
	Xmlns: models.BASE_XMLNS,
	Result: models.GetQueueAttributesResult{Attrs: []models.Attribute{
		models.Attribute{
			Name:  "DelaySeconds",
			Value: "0",
		},
		models.Attribute{
			Name:  "MaximumMessageSize",
			Value: "262144",
		},
		models.Attribute{
			Name:  "MessageRetentionPeriod",
			Value: "345600",
		},
		models.Attribute{
			Name:  "ReceiveMessageWaitTimeSeconds",
			Value: "0",
		},
		models.Attribute{
			Name:  "VisibilityTimeout",
			Value: "30",
		},
		models.Attribute{
			Name:  "ApproximateNumberOfMessages",
			Value: "0",
		},
		models.Attribute{
			Name:  "ApproximateNumberOfMessagesNotVisible",
			Value: "0",
		},
		models.Attribute{
			Name:  "CreatedTimestamp",
			Value: "0000000000",
		},
		models.Attribute{
			Name:  "LastModifiedTimestamp",
			Value: "0000000000",
		},
		models.Attribute{
			Name:  "QueueArn",
			Value: "arn:aws:sqs:region:accountID:unit-queue1",
		},
	}},
	Metadata: models.BASE_RESPONSE_METADATA,
}

var SetQueueAttributesRequest = models.SetQueueAttributesRequest{
	QueueUrl:   fmt.Sprintf("%s/%s", BASE_URL, "unit-queue1"),
	Attributes: QueueAttributes,
}
