package gosns

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	log "github.com/sirupsen/logrus"
)

func ListTopicsV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewListTopicsRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - ListTopicsV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	log.Debug("Listing Topics")
	arnList := make([]models.TopicArnResult, 0)

	for _, topic := range app.SyncTopics.Topics {
		ta := models.TopicArnResult{TopicArn: topic.Arn}
		arnList = append(arnList, ta)
	}

	requestId := uuid.NewString()
	respStruct := models.ListTopicsResponse{
		Xmlns:    models.BASE_XMLNS,
		Result:   models.ListTopicsResult{Topics: models.TopicNamestype{Member: arnList}},
		Metadata: app.ResponseMetadata{RequestId: requestId},
	}

	return http.StatusOK, respStruct
}
