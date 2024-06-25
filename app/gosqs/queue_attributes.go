package gosqs

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app"
)

// TODO - Support:
//   - attr.MessageRetentionPeriod
//   - attr.Policy
//   - attr.RedriveAllowPolicy
func setQueueAttributesV1(q *app.Queue, attr models.QueueAttributes) error {
	// FIXME - are there better places to put these bottom-limit validations?
	if attr.DelaySeconds >= 0 {
		q.DelaySeconds = attr.DelaySeconds.Int()
	}
	if attr.MaximumMessageSize >= 0 {
		q.MaximumMessageSize = attr.MaximumMessageSize.Int()
	}
	// TODO - bottom limit should be the AWS limits
	// The following 2 don't support zero values
	if attr.MessageRetentionPeriod > 0 {
		q.MessageRetentionPeriod = attr.MessageRetentionPeriod.Int()
	}
	if attr.ReceiveMessageWaitTimeSeconds > 0 {
		q.ReceiveMessageWaitTimeSeconds = attr.ReceiveMessageWaitTimeSeconds.Int()
	}
	if attr.VisibilityTimeout >= 0 {
		q.VisibilityTimeout = attr.VisibilityTimeout.Int()
	}
	if attr.RedrivePolicy != (models.RedrivePolicy{}) {
		arnArray := strings.Split(attr.RedrivePolicy.DeadLetterTargetArn, ":")
		queueName := arnArray[len(arnArray)-1]
		deadLetterQueue, ok := app.SyncQueues.Queues[queueName]
		if !ok {
			log.Error("Invalid RedrivePolicy Attribute")
			return fmt.Errorf("InvalidAttributeValue")
		}
		q.DeadLetterQueue = deadLetterQueue
		q.MaxReceiveCount = attr.RedrivePolicy.MaxReceiveCount.Int()
	}
	return nil
}
