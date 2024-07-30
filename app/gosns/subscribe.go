package gosns

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/Admiral-Piett/goaws/app/interfaces"
	log "github.com/sirupsen/logrus"
)

// aws --endpoint-url http://localhost:47194 sns subscribe --topic-arn arn:aws:sns:us-west-2:0123456789012:my-topic --protocol email --notification-endpoint my-email@example.com
func SubscribeV1(req *http.Request) (int, interfaces.AbstractResponseBody) {
	requestBody := models.NewSubscribeRequest()
	ok := utils.REQUEST_TRANSFORMER(requestBody, req, false)
	if !ok {
		log.Error("Invalid Request - SubscribeV1")
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}

	uriSegments := strings.Split(requestBody.TopicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]
	extraLogFields := log.Fields{
		"topicArn":     requestBody.TopicArn,
		"topicName":    topicName,
		"protocol":     requestBody.Protocol,
		"endpoint":     requestBody.Endpoint,
		"filterPolicy": requestBody.Attributes.FilterPolicy,
		"raw":          requestBody.Attributes.RawMessageDelivery,
	}
	log.WithFields(extraLogFields).Info("Creating Subscription")

	subscription := &app.Subscription{EndPoint: requestBody.Endpoint, Protocol: requestBody.Protocol, TopicArn: requestBody.TopicArn, Raw: requestBody.Attributes.RawMessageDelivery, FilterPolicy: &requestBody.Attributes.FilterPolicy}

	subscription.SubscriptionArn = fmt.Sprintf("%s:%s", requestBody.TopicArn, uuid.NewString())

	//Create the response
	requestId := uuid.NewString()
	respStruct := models.SubscribeResponse{Xmlns: models.BASE_XMLNS, Result: models.SubscribeResult{SubscriptionArn: subscription.SubscriptionArn}, Metadata: app.ResponseMetadata{RequestId: requestId}}
	if app.SyncTopics.Topics[topicName] != nil {
		app.SyncTopics.Lock()
		isDuplicate := false
		// Duplicate check
		for _, sub := range app.SyncTopics.Topics[topicName].Subscriptions {
			if sub.EndPoint == requestBody.Endpoint && sub.TopicArn == requestBody.TopicArn {
				isDuplicate = true
				sub.SubscriptionArn = subscription.SubscriptionArn
			}
		}
		if !isDuplicate {
			app.SyncTopics.Topics[topicName].Subscriptions = append(app.SyncTopics.Topics[topicName].Subscriptions, subscription)
			log.WithFields(extraLogFields).Debug("Created subscription")
		}
		app.SyncTopics.Unlock()

		if app.Protocol(subscription.Protocol) == app.ProtocolHTTP || app.Protocol(subscription.Protocol) == app.ProtocolHTTPS {
			id := uuid.NewString()
			token := uuid.NewString()

			TOPIC_DATA[requestBody.TopicArn] = &pendingConfirm{
				subArn: subscription.SubscriptionArn,
				token:  token,
			}

			//QUESTION - do we need this?
			time.Sleep(time.Second)

			snsMSG := &app.SNSMessage{
				Type:             "SubscriptionConfirmation",
				MessageId:        id,
				Token:            token,
				TopicArn:         requestBody.TopicArn,
				Message:          fmt.Sprintf("You have chosen to subscribe to the topic %s.\nTo confirm the subscription, visit the SubscribeURL included in this message.", requestBody.TopicArn),
				SigningCertURL:   fmt.Sprintf("http://%s:%s/SimpleNotificationService/%s.pem", app.CurrentEnvironment.Host, app.CurrentEnvironment.Port, requestId),
				SignatureVersion: "1",
				SubscribeURL:     fmt.Sprintf("http://%s:%s/?Action=ConfirmSubscription&TopicArn=%s&Token=%s", app.CurrentEnvironment.Host, app.CurrentEnvironment.Port, requestBody.TopicArn, token),
				Timestamp:        time.Now().UTC().Format(time.RFC3339),
			}
			signature, err := signMessage(PrivateKEY, snsMSG)
			if err != nil {
				log.Error("Error signing message")
			} else {
				snsMSG.Signature = signature
			}
			err = callEndpoint(subscription.EndPoint, requestId, *snsMSG, subscription.Raw)
			if err != nil {
				log.Error("Error posting to url ", err)
			}
		}

	} else {
		return utils.CreateErrorResponseV1("InvalidParameterValue", false)
	}
	return http.StatusOK, respStruct
}
