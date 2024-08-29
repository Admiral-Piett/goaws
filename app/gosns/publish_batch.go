package gosns

import (
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app"
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

	if len(requestBody.PublishBatchRequestEntries.Member) == 0 {
		return utils.CreateErrorResponseV1("EmptyBatchRequest", false)
	}
	if len(requestBody.PublishBatchRequestEntries.Member) > 10 {
		return utils.CreateErrorResponseV1("TooManyEntriesInBatchRequest", false)
	}

	topicArn := requestBody.TopicArn
	arnSegments := strings.Split(topicArn, ":")
	topicName := arnSegments[len(arnSegments)-1]
	topic, ok := app.SyncTopics.Topics[topicName]
	if !ok {
		return utils.CreateErrorResponseV1("TopicNotFound", false)
	}

	seen := make(map[string]bool)

	for _, entry := range requestBody.PublishBatchRequestEntries.Member {
		if entry == nil { // we use gorilla schema to parse value params. Indexing on the aws client starts at 1 but gorilla schema starts at 0 so we may have a nil entry at the start of the slice
			continue
		}
		if entry.ID == "" {
			// This is a required field for the PublishBatchRequestEntry entity but doesn't seem required in the request.
			// If it's not present in the request then assume we should generate one.
			entry.ID = uuid.NewString()
		}
		if seen[entry.ID] {
			return utils.CreateErrorResponseV1("BatchEntryIdsNotDistinct", false)
		}
		seen[entry.ID] = true
	}

	successfulEntries := []models.PublishBatchResultEntry{}
	failedEntries := []models.BatchResultErrorEntry{}
	for _, entry := range requestBody.PublishBatchRequestEntries.Member {
		if entry == nil { // we use gorilla schema to parse value params. Indexing on the aws client starts at 1 but gorilla schema starts at 0 so we may have a nil entry at the start of the slice
			continue
		}
		// we now know all the entry.IDs are unique and non-blank
		for _, sub := range topic.Subscriptions {
			switch app.Protocol(sub.Protocol) {
			case app.ProtocolSQS:
				if err := publishSQS(sub, topicName, entry); err != nil {
					er := models.SnsErrors[err.Error()]
					failedEntries = append(failedEntries, models.BatchResultErrorEntry{
						Code:        er.Code,
						Id:          entry.ID,
						Message:     er.Message,
						SenderFault: true,
					})
				} else {
					msgId := uuid.NewString()
					successfulEntries = append(successfulEntries, models.PublishBatchResultEntry{
						Id:        entry.ID,
						MessageId: msgId,
					})
				}
			case app.ProtocolHTTP:
				fallthrough
			case app.ProtocolHTTPS:
				publishHTTP(sub, topicArn, entry)
				msgId := uuid.NewString()
				successfulEntries = append(successfulEntries, models.PublishBatchResultEntry{
					Id:        entry.ID,
					MessageId: msgId,
				})
			}
		}
	}

	respStruct := models.PublishBatchResponse{
		Xmlns: models.BASE_XMLNS,
		Result: models.PublishBatchResult{
			Successful: models.PublishBatchSuccessful{SuccessEntries: successfulEntries},
			Failed:     models.PublishBatchFailed{ErrorEntries: failedEntries},
		},
		Metadata: app.ResponseMetadata{RequestId: uuid.NewString()},
	}
	return http.StatusOK, respStruct
}
