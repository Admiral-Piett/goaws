package gosqs

import (
	"net/http"
	"time"

	"github.com/Admiral-Piett/goaws/app"
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

	queueUrl := "http://" + app.CurrentEnvironment.Host + ":" + app.CurrentEnvironment.Port +
		"/" + app.CurrentEnvironment.AccountID + "/" + queueName
	if app.CurrentEnvironment.Region != "" {
		queueUrl = "http://" + app.CurrentEnvironment.Region + "." + app.CurrentEnvironment.Host + ":" +
			app.CurrentEnvironment.Port + "/" + app.CurrentEnvironment.AccountID + "/" + queueName
	}
	queueArn := "arn:aws:sqs:" + app.CurrentEnvironment.Region + ":" + app.CurrentEnvironment.AccountID + ":" + queueName

	if _, ok := app.SyncQueues.Queues[queueName]; !ok {
		log.Println("Creating Queue:", queueName)
		queue := &app.Queue{
			Name:             queueName,
			URL:              queueUrl,
			Arn:              queueArn,
			IsFIFO:           app.HasFIFOQueueName(queueName),
			EnableDuplicates: app.CurrentEnvironment.EnableDuplicates,
			Duplicates:       make(map[string]time.Time),
		}
		if err := setQueueAttributesV1(queue, requestBody.Attributes); err != nil {
			return utils.CreateErrorResponseV1(err.Error(), true)
		}
		app.SyncQueues.Lock()
		app.SyncQueues.Queues[queueName] = queue
		app.SyncQueues.Unlock()
	}

	respStruct := models.CreateQueueResponse{
		Xmlns:    models.BASE_XMLNS,
		Result:   models.CreateQueueResult{QueueUrl: queueUrl},
		Metadata: models.BASE_RESPONSE_METADATA,
	}
	return http.StatusOK, respStruct
}
