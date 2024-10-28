package gosns

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	log "github.com/sirupsen/logrus"
)

// TODO - Admiral-Piett - Pick a MessageAttribute style and get rid of `utils.ConvertToOldMessageAttributeValueStructure`

// aws --endpoint-url http://localhost:47194 sns publish --topic-arn arn:aws:sns:yopa-local:000000000000:test1 --message "This is a test"
func PublishV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewPublishRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - PublishV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	// TODO - support TargetArn
	if requestBody.TopicArn == "" || requestBody.Message == "" {
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	arnSegments := strings.Split(requestBody.TopicArn, ":")
	topicName := arnSegments[len(arnSegments)-1]

	topic, ok := models.SyncTopics.Topics[topicName]
	if !ok {
		return utils.CreateErrorResponseV1("TopicNotFound", false)
	}
	log.WithFields(log.Fields{
		"topic":    topicName,
		"topicArn": requestBody.TopicArn,
		"subject":  requestBody.Subject,
	}).Debug("Publish to Topic")

	messageId, err := publishMessageByTopicFunc(topic, requestBody)
	if err != nil {
		utils.CreateErrorResponseV1(err.Error(), false)
	}

	//Create the response
	respStruct := models.PublishResponse{
		Xmlns: models.BaseXmlns,
		Result: models.PublishResult{
			MessageId: messageId,
		},
		Metadata: models.ResponseMetadata{
			RequestId: uuid.NewString(),
		},
	}
	return http.StatusOK, respStruct
}
