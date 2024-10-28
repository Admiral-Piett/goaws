package gosns

import (
	"net/http"
	"strings"

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

	messageCount := len(requestBody.PublishBatchRequestEntries.Member)
	if messageCount == 0 {
		return utils.CreateErrorResponseV1("EmptyBatchRequest", false)
	}

	// The marshaller will populate a nil entry for its 0 index, so pop it here since the requests
	//will start with index 1.
	if requestBody.PublishBatchRequestEntries.Member[0] == nil {
		requestBody.PublishBatchRequestEntries.Member = requestBody.PublishBatchRequestEntries.Member[1:]
	}

	if messageCount > 10 {
		return utils.CreateErrorResponseV1("TooManyEntriesInBatchRequest", false)
	}

	idMap := make(map[string]bool)
	for _, message := range requestBody.PublishBatchRequestEntries.Member {
		// The SDKs fail if you don't provide an ID, so make sure we honor that here too.  You need one anyway.
		if message.ID == "" {
			return utils.CreateErrorResponseV1("InvalidParameterValue", false)
		}
		_, seen := idMap[message.ID]
		if seen {
			return utils.CreateErrorResponseV1("BatchEntryIdsNotDistinct", false)
		}
		idMap[message.ID] = true
	}

	arnSegments := strings.Split(requestBody.TopicArn, ":")
	topicName := arnSegments[len(arnSegments)-1]
	topic, ok := models.SyncTopics.Topics[topicName]
	if !ok {
		return utils.CreateErrorResponseV1("TopicNotFound", false)
	}

	successfulEntries := []models.PublishBatchResultEntry{}
	failedEntries := []models.BatchResultErrorEntry{}
	for _, entry := range requestBody.PublishBatchRequestEntries.Member {
		messageId, err := publishMessageByTopicFunc(topic, entry)
		if err != nil {
			er := models.SnsErrors[err.Error()]
			failedEntries = append(failedEntries, models.BatchResultErrorEntry{
				Code:        er.Code,
				Id:          entry.ID,
				Message:     er.Message,
				SenderFault: true,
			})
			continue
		}
		successfulEntries = append(successfulEntries, models.PublishBatchResultEntry{
			Id:        entry.ID,
			MessageId: messageId,
		})
	}

	respStruct := models.PublishBatchResponse{
		Xmlns: models.BaseXmlns,
		Result: models.PublishBatchResult{
			Successful: models.PublishBatchSuccessful{SuccessEntries: successfulEntries},
			Failed:     models.PublishBatchFailed{ErrorEntries: failedEntries},
		},
		Metadata: models.ResponseMetadata{RequestId: uuid.NewString()},
	}
	return http.StatusOK, respStruct
}
