package gosqs

import (
	"net/http"
	"time"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	log "github.com/sirupsen/logrus"
)

func CreateQueueV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewCreateQueueRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - CreateQueueV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", true)
	}
	queueName := requestBody.QueueName

	queueUrl := "http://" + models.CurrentEnvironment.Host + ":" + models.CurrentEnvironment.Port +
		"/" + models.CurrentEnvironment.AccountID + "/" + queueName
	if models.CurrentEnvironment.Region != "" {
		queueUrl = "http://" + models.CurrentEnvironment.Region + "." + models.CurrentEnvironment.Host + ":" +
			models.CurrentEnvironment.Port + "/" + models.CurrentEnvironment.AccountID + "/" + queueName
	}
	queueArn := "arn:aws:sqs:" + models.CurrentEnvironment.Region + ":" + models.CurrentEnvironment.AccountID + ":" + queueName

	if _, ok := models.SyncQueues.Queues[queueName]; !ok {
		log.Infof("Creating Queue: %s", queueName)
		queue := &models.Queue{
			Name:             queueName,
			URL:              queueUrl,
			Arn:              queueArn,
			IsFIFO:           utils.HasFIFOQueueName(queueName),
			EnableDuplicates: models.CurrentEnvironment.EnableDuplicates,
			Duplicates:       make(map[string]time.Time),
		}
		if err := setQueueAttributesV1(queue, requestBody.Attributes); err != nil {
			return utils.CreateErrorResponseV1(err.Error(), true)
		}
		models.SyncQueues.Lock()
		models.SyncQueues.Queues[queueName] = queue
		models.SyncQueues.Unlock()
	}

	respStruct := models.CreateQueueResponse{
		Xmlns:    models.BaseXmlns,
		Result:   models.CreateQueueResult{QueueUrl: queueUrl},
		Metadata: models.BaseResponseMetadata,
	}
	return http.StatusOK, respStruct
}
