package gosns

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/common"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func PublishBatchV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewPublishBatchRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - PublishBatchV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	if requestBody.TopicArn == "" {
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	topicArn := requestBody.TopicArn
	arnSegments := strings.Split(topicArn, ":")
	topicName := arnSegments[len(arnSegments)-1]
	topic, ok := app.SyncTopics.Topics[topicName]
	if !ok {
		return utils.CreateErrorResponseV1("TopicNotFound", false)
	}

	seen := make(map[string]bool)

	for _, entry := range requestBody.PublishBatchRequestEntries {
		if entry.ID == "" {
			// This is a required field, its absence likely indicates there are no further entries.
			// It is unclear from the AWS docs if an error is returned if there are other fields
			// present for PublishBatchRequestEntries.member.N where N is some integer in range [1,10].
			break
		}
		if seen[entry.ID] {
			return utils.CreateErrorResponseV1("BatchEntryIdsNotDistinct", false)
		}
		seen[entry.ID] = true
	}

	if len(requestBody.PublishBatchRequestEntries) == 0 {
		return utils.CreateErrorResponseV1("EmptyBatchRequest", false)
	}
	if len(requestBody.PublishBatchRequestEntries) > 10 {
		return utils.CreateErrorResponseV1("TooManyEntriesInBatchRequest", false)
	}

	successfulEntries := []models.PublishBatchResultEntry{}
	failedEntries := []models.BatchResultErrorEntry{}
	for _, entry := range requestBody.PublishBatchRequestEntries {
		// we now know all the entry.IDs are unique
		if entry.ID == "" {
			// This is a required field, its absence likely indicates there are no further entries.
			// It is unclear from the AWS docs if an error is returned if there are other fields
			// present for PublishBatchRequestEntries.member.N where N is some integer in range [1,10].
			break
		}
		for _, sub := range topic.Subscriptions {
			switch app.Protocol(sub.Protocol) {
			case app.ProtocolSQS:
				oldMessageAttributes := utils.ConvertToOldMessageAttributeValueStructure(entry.MessageAttributes)
				if err := publishSQS(sub, entry.Message, oldMessageAttributes, entry.Subject, topicName, entry.MessageStructure); err != nil {
					er := models.SnsErrors[err.Error()]
					failedEntries = append(failedEntries, models.BatchResultErrorEntry{
						Code:        er.Code,
						Id:          entry.ID,
						Message:     er.Message,
						SenderFault: true,
					})
				} else {
					msgId, _ := common.NewUUID()
					successfulEntries = append(successfulEntries, models.PublishBatchResultEntry{
						Id:        entry.ID,
						MessageId: msgId,
					})
				}
			case app.ProtocolHTTP:
				fallthrough
			case app.ProtocolHTTPS:
				oldMessageAttributes := utils.ConvertToOldMessageAttributeValueStructure(entry.MessageAttributes)
				publishHTTP(sub, entry.Message, oldMessageAttributes, entry.Subject, topicArn)
				msgId, _ := common.NewUUID()
				successfulEntries = append(successfulEntries, models.PublishBatchResultEntry{
					Id:        entry.ID,
					MessageId: msgId,
				})
			}
		}
	}

	respStruct := models.PublishBatchResponse{
		// "https://sns.amazonaws.com/doc/2010-03-31/",
		Xmlns: models.BASE_XMLNS,
		Result: models.PublishBatchResult{
			Successful: models.PublishBatchSuccessful{SuccessEntries: successfulEntries},
			Failed:     models.PublishBatchFailed{ErrorEntries: failedEntries},
		},
		Metadata: app.ResponseMetadata{RequestId: uuid.NewString()},
	}
	return http.StatusOK, respStruct
}
