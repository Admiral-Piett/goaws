package gosns

import (
	"net/http"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func ConfirmSubscriptionV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewConfirmSubscriptionRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - ConfirmSubscriptionV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}
	topicArn := requestBody.TopicArn
	confirmToken := requestBody.Token
	var pendingConfirm pendingConfirm

	if pending, ok := TOPIC_DATA[topicArn]; !ok {
		return utils.CreateErrorResponseV1("SubscriptionNotFound", false)
	} else {
		pendingConfirm = *pending
	}

	if pendingConfirm.token != confirmToken {
		return utils.CreateErrorResponseV1("SubscriptionNotFound", false)
	}
	respStruct := models.ConfirmSubscriptionResponse{
		Xmlns:    models.BASE_XMLNS,
		Result:   models.ConfirmSubscriptionResult{SubscriptionArn: pendingConfirm.subArn},
		Metadata: app.ResponseMetadata{RequestId: uuid.NewString()},
	}
	return http.StatusOK, respStruct
}
