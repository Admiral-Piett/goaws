package gosqs

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/mitchellh/copystructure"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	log "github.com/sirupsen/logrus"
)

func GetQueueAttributesV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewGetQueueAttributesRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - GetQueueAttributesV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}
	if requestBody.QueueUrl == "" {
		log.Error("Missing QueueUrl - GetQueueAttributesV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	requestedAttributes := func() map[string]bool {
		attrs := map[string]bool{}
		if len(requestBody.AttributeNames) == 0 {
			return map[string]bool{"All": true}
		}
		for _, attr := range requestBody.AttributeNames {
			if "All" == attr {
				return map[string]bool{"All": true}
			}
			attrs[attr] = true
		}
		return attrs
	}()

	dupe, _ := copystructure.Copy(models.AVAILABLE_QUEUE_ATTRIBUTES)
	includedAttributes, _ := dupe.(map[string]bool)
	_, ok = requestedAttributes["All"]
	if !ok {
		for attr, _ := range includedAttributes {
			_, ok := requestedAttributes[attr]
			if !ok {
				delete(includedAttributes, attr)
			}
		}
	}

	uriSegments := strings.Split(requestBody.QueueUrl, "/")
	queueName := uriSegments[len(uriSegments)-1]

	log.Infof("Get Queue QueueAttributes: %s", queueName)
	queueAttributes := make([]models.Attribute, 0, 0)

	app.SyncQueues.RLock()
	defer app.SyncQueues.RUnlock()
	queue, ok := app.SyncQueues.Queues[queueName]
	if !ok {
		log.Errorf("Get Queue URL: %s queue does not exist!!!", queueName)
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}

	if _, ok := includedAttributes["DelaySeconds"]; ok {
		attr := models.Attribute{Name: "DelaySeconds", Value: strconv.Itoa(queue.DelaySeconds)}
		queueAttributes = append(queueAttributes, attr)
	}
	if _, ok := includedAttributes["MaximumMessageSize"]; ok {
		attr := models.Attribute{Name: "MaximumMessageSize", Value: strconv.Itoa(queue.MaximumMessageSize)}
		queueAttributes = append(queueAttributes, attr)
	}
	if _, ok := includedAttributes["MessageRetentionPeriod"]; ok {
		attr := models.Attribute{Name: "MessageRetentionPeriod", Value: strconv.Itoa(queue.MessageRetentionPeriod)}
		queueAttributes = append(queueAttributes, attr)
	}
	if _, ok := includedAttributes["ReceiveMessageWaitTimeSeconds"]; ok {
		attr := models.Attribute{Name: "ReceiveMessageWaitTimeSeconds", Value: strconv.Itoa(queue.ReceiveMessageWaitTimeSeconds)}
		queueAttributes = append(queueAttributes, attr)
	}
	if _, ok := includedAttributes["VisibilityTimeout"]; ok {
		attr := models.Attribute{Name: "VisibilityTimeout", Value: strconv.Itoa(queue.VisibilityTimeout)}
		queueAttributes = append(queueAttributes, attr)
	}
	if _, ok := includedAttributes["ApproximateNumberOfMessages"]; ok {
		attr := models.Attribute{Name: "ApproximateNumberOfMessages", Value: strconv.Itoa(len(queue.Messages))}
		queueAttributes = append(queueAttributes, attr)
	}
	// TODO - implement
	//if _, ok := includedAttributes["ApproximateNumberOfMessagesDelayed"]; ok {
	//	attr := models.Attribute{Name: "ApproximateNumberOfMessagesDelayed", Value: strconv.Itoa(len(queue.Messages))}
	//	queueAttributes = append(queueAttributes, attr)
	//}
	if _, ok := includedAttributes["ApproximateNumberOfMessagesNotVisible"]; ok {
		attr := models.Attribute{Name: "ApproximateNumberOfMessagesNotVisible", Value: strconv.Itoa(numberOfHiddenMessagesInQueue(*queue))}
		queueAttributes = append(queueAttributes, attr)
	}
	if _, ok := includedAttributes["CreatedTimestamp"]; ok {
		attr := models.Attribute{Name: "CreatedTimestamp", Value: "0000000000"}
		queueAttributes = append(queueAttributes, attr)
	}
	if _, ok := includedAttributes["LastModifiedTimestamp"]; ok {
		attr := models.Attribute{Name: "LastModifiedTimestamp", Value: "0000000000"}
		queueAttributes = append(queueAttributes, attr)
	}
	if _, ok := includedAttributes["QueueArn"]; ok {
		attr := models.Attribute{Name: "QueueArn", Value: queue.Arn}
		queueAttributes = append(queueAttributes, attr)
	}
	// TODO - implement
	//if _, ok := includedAttributes["Policy"]; ok {
	//	attr := models.Attribute{Name: "Policy", Value: ""}
	//	queueAttributes = append(queueAttributes, attr)
	//}
	//if _, ok := includedAttributes["RedriveAllowPolicy"]; ok {
	//	attr := models.Attribute{Name: "RedriveAllowPolicy", Value: ""}
	//	queueAttributes = append(queueAttributes, attr)
	//}
	if _, ok := includedAttributes["RedrivePolicy"]; ok && queue.DeadLetterQueue != nil {
		attr := models.Attribute{Name: "RedrivePolicy", Value: fmt.Sprintf(`{"maxReceiveCount":"%d", "deadLetterTargetArn":"%s"}`, queue.MaxReceiveCount, queue.DeadLetterQueue.Arn)}
		queueAttributes = append(queueAttributes, attr)
	}

	respStruct := models.GetQueueAttributesResponse{
		Xmlns:    models.BASE_XMLNS,
		Result:   models.GetQueueAttributesResult{Attrs: queueAttributes},
		Metadata: models.BASE_RESPONSE_METADATA,
	}
	return http.StatusOK, respStruct
}
