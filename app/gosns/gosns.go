package gosns

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	"github.com/Admiral-Piett/goaws/app/utils"

	"github.com/google/uuid"

	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"math/big"

	log "github.com/sirupsen/logrus"
)

type pendingConfirm struct {
	subArn string
	token  string
}

var PemKEY []byte
var PrivateKEY *rsa.PrivateKey
var TOPIC_DATA map[string]*pendingConfirm

func init() {
	models.SyncTopics.Topics = make(map[string]*models.Topic)
	TOPIC_DATA = make(map[string]*pendingConfirm)

	PrivateKEY, PemKEY, _ = createPemFile()
}

func createPemFile() (privkey *rsa.PrivateKey, pemkey []byte, err error) {
	template := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{11, 22, 33},
		SerialNumber:          big.NewInt(1111),
		Subject: pkix.Name{
			Country:      []string{"USA"},
			Organization: []string{"Amazon"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Duration(5) * time.Second),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// generate private key
	privkey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	// create a self-signed certificate
	parent := template
	cert, err := x509.CreateCertificate(rand.Reader, template, parent, &privkey.PublicKey, privkey)
	if err != nil {
		return
	}

	pemkey = pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		},
	)
	return
}

func signMessage(privkey *rsa.PrivateKey, snsMsg *models.SNSMessage) (string, error) {
	fs, err := formatSignature(snsMsg)
	if err != nil {
		return "", nil
	}

	h := sha1.Sum([]byte(fs))
	signature_b, err := rsa.SignPKCS1v15(rand.Reader, privkey, crypto.SHA1, h[:])

	return base64.StdEncoding.EncodeToString(signature_b), err
}

func formatSignature(msg *models.SNSMessage) (formated string, err error) {
	if msg.Type == "Notification" && msg.Subject != "" {
		formated = fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
			"Message", msg.Message,
			"MessageId", msg.MessageId,
			"Subject", msg.Subject,
			"Timestamp", msg.Timestamp,
			"TopicArn", msg.TopicArn,
			"Type", msg.Type,
		)
	} else if msg.Type == "Notification" && msg.Subject == "" {
		formated = fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
			"Message", msg.Message,
			"MessageId", msg.MessageId,
			"Timestamp", msg.Timestamp,
			"TopicArn", msg.TopicArn,
			"Type", msg.Type,
		)
	} else if msg.Type == "SubscriptionConfirmation" || msg.Type == "UnsubscribeConfirmation" {
		formated = fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
			"Message", msg.Message,
			"MessageId", msg.MessageId,
			"SubscribeURL", msg.SubscribeURL,
			"Timestamp", msg.Timestamp,
			"Token", msg.Token,
			"TopicArn", msg.TopicArn,
			"Type", msg.Type,
		)
	} else {
		return formated, errors.New("Unable to determine SNSMessage type")
	}

	return
}

// NOTE: The use case for this is to use GoAWS to call some external system with the message payload.  Essentially
// it is a localized subscription to some non-AWS endpoint.
func callEndpoint(endpoint string, subArn string, msg models.SNSMessage, raw bool) error {
	log.WithFields(log.Fields{
		"sns":      msg,
		"subArn":   subArn,
		"endpoint": endpoint,
	}).Debug("Calling endpoint")
	var err error
	var byteData []byte

	if raw {
		byteData, err = json.Marshal(msg.Message)
	} else {
		byteData, err = json.Marshal(msg)
	}
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(byteData))
	if err != nil {
		return err
	}

	//req.Header.Add("Authorization", "Basic YXV0aEhlYWRlcg==")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-amz-sns-message-type", msg.Type)
	req.Header.Add("x-amz-sns-message-id", msg.MessageId)
	req.Header.Add("x-amz-sns-topic-arn", msg.TopicArn)
	req.Header.Add("x-amz-sns-subscription-arn", subArn)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("response is nil")
	}

	//Amazon considers a Notification delivery attempt successful if the endpoint
	//responds in the range of 200-499. Response codes outside that range will
	//trigger the Subscription's retry policy.
	//https://docs.aws.amazon.com/sns/latest/dg/SendMessageToHttp.prepare.html
	if res.StatusCode < 200 || res.StatusCode > 499 {
		log.WithFields(log.Fields{
			"statusCode": res.StatusCode,
			"status":     res.Status,
			"header":     res.Header,
			"endpoint":   endpoint,
		}).Error("Response outside of acceptable (200-499) range")
		return errors.New("Response outside of acceptable (200-499) range")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"body": string(body),
		"res":  res,
	}).Debug("Received successful response")

	return nil
}

func extractMessageFromJSON(msg string, protocol string) (string, error) {
	var msgWithProtocols map[string]string
	if err := json.Unmarshal([]byte(msg), &msgWithProtocols); err != nil {
		return "", err
	}

	defaultMsg, ok := msgWithProtocols[string(models.ProtocolDefault)]
	if !ok {
		return "", errors.New("Invalid parameter: Message Structure - No default entry in JSON message body")
	}

	if m, ok := msgWithProtocols[protocol]; ok {
		return m, nil
	}

	return defaultMsg, nil
}

func getSubscription(subsArn string) *models.Subscription {
	for _, topic := range models.SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subsArn {
				return sub
			}
		}
	}
	return nil
}

func createMessageBody(subs *models.Subscription, entry interfaces.AbstractPublishEntry,
	messageAttributes map[string]models.SqsMessageAttributeValue) ([]byte, error) {

	msgId := uuid.NewString()
	message := models.SNSMessage{
		Type:              "Notification",
		MessageId:         msgId,
		TopicArn:          subs.TopicArn,
		Subject:           entry.GetSubject(),
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		SignatureVersion:  "1",
		SigningCertURL:    fmt.Sprintf("http://%s:%s/SimpleNotificationService/%s.pem", models.CurrentEnvironment.Host, models.CurrentEnvironment.Port, msgId),
		UnsubscribeURL:    fmt.Sprintf("http://%s:%s/?Action=Unsubscribe&SubscriptionArn=%s", models.CurrentEnvironment.Host, models.CurrentEnvironment.Port, subs.SubscriptionArn),
		MessageAttributes: formatAttributes(messageAttributes),
	}

	if models.MessageStructure(entry.GetMessageStructure()) == models.MessageStructureJSON {
		m, err := extractMessageFromJSON(entry.GetMessage(), subs.Protocol)
		if err != nil {
			return nil, err
		}
		message.Message = m
	} else {
		message.Message = entry.GetMessage()
	}

	signature, err := signMessage(PrivateKEY, &message)
	if err != nil {
		log.Error(err)
	} else {
		message.Signature = signature
	}

	byteMsg, _ := json.Marshal(message)
	return byteMsg, nil
}

func formatAttributes(values map[string]models.SqsMessageAttributeValue) map[string]models.MessageAttributeValue {
	attr := make(map[string]models.MessageAttributeValue)
	for k, v := range values {
		if v.DataType == "String" {
			attr[k] = models.MessageAttributeValue{
				DataType:    v.DataType,
				StringValue: v.Value,
			}
		} else {
			attr[k] = models.MessageAttributeValue{
				DataType:    v.DataType,
				BinaryValue: v.Value, // TODO - this may need to be a []byte?
			}
		}
	}
	return attr
}

func publishHTTP(subs *models.Subscription, topicArn string, entry interfaces.AbstractPublishEntry) {
	messageAttributes := utils.ConvertToOldMessageAttributeValueStructure(entry.GetMessageAttributes())
	id := uuid.NewString()
	msg := models.SNSMessage{
		Type:              "Notification",
		MessageId:         id,
		TopicArn:          topicArn,
		Subject:           entry.GetSubject(),
		Message:           entry.GetMessage(),
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		SignatureVersion:  "1",
		SigningCertURL:    fmt.Sprintf("http://%s:%s/SimpleNotificationService/%s.pem", models.CurrentEnvironment.Host, models.CurrentEnvironment.Port, id),
		UnsubscribeURL:    fmt.Sprintf("http://%s:%s/?Action=Unsubscribe&SubscriptionArn=%s", models.CurrentEnvironment.Host, models.CurrentEnvironment.Port, subs.SubscriptionArn),
		MessageAttributes: formatAttributes(messageAttributes),
	}

	signature, err := signMessage(PrivateKEY, &msg)
	if err != nil {
		log.Error(err)
	} else {
		msg.Signature = signature
	}
	err = callEndpoint(subs.EndPoint, subs.SubscriptionArn, msg, subs.Raw)
	if err != nil {
		log.WithFields(log.Fields{
			"EndPoint": subs.EndPoint,
			"ARN":      subs.SubscriptionArn,
			"error":    err.Error(),
		}).Error("Error calling endpoint")
	}
}

// NOTE: The important thing to know here is that essentially the RAW delivery means we take the message body and
// put it in the resulting `body`, so that's all that's in that field when the message is received.  If it's not
// raw, then we put all this other junk in there too, similar to how AWS stores its metadata in there.
func publishSQS(subscription *models.Subscription, topic *models.Topic, entry interfaces.AbstractPublishEntry) error {
	messageAttributes := utils.ConvertToOldMessageAttributeValueStructure(entry.GetMessageAttributes())
	if subscription.FilterPolicy != nil && !subscription.FilterPolicy.IsSatisfiedBy(messageAttributes) {
		return nil
	}

	endPoint := subscription.EndPoint
	uriSegments := strings.Split(endPoint, "/")
	queueName := uriSegments[len(uriSegments)-1]
	arnSegments := strings.Split(queueName, ":")
	queueName = arnSegments[len(arnSegments)-1]

	if _, ok := models.SyncQueues.Queues[queueName]; ok {
		msg := models.SqsMessage{}

		if subscription.Raw {
			msg.MessageAttributes = messageAttributes
			msg.MD5OfMessageAttributes = utils.HashAttributes(messageAttributes)

			// NOTE: Admiral-Piett - commenting this out.  I don't understand what this is supposed to achieve
			// for raw message delivery.  I suspect this doesn't work at all, otherwise you'd have to match the
			//json message structure pattern with a `default` key at the root to indicate your base message and
			//all the rest.  I don't think that makes sense for raw delivery.
			//m, err := extractMessageFromJSON(entry.GetMessage(), subscription.Protocol)
			//if err == nil {
			//	msg.MessageBody = []byte(m)
			//} else {
			//	msg.MessageBody = []byte(entry.GetMessage())
			//}
			msg.MessageBody = []byte(entry.GetMessage())
		} else {
			m, err := createMessageBody(subscription, entry, messageAttributes)
			if err != nil {
				return err
			}

			msg.MessageBody = m
		}

		msg.MD5OfMessageBody = utils.GetMD5Hash(entry.GetMessage())
		msg.Uuid = uuid.NewString()
		models.SyncQueues.Lock()
		models.SyncQueues.Queues[queueName].Messages = append(models.SyncQueues.Queues[queueName].Messages, msg)
		models.SyncQueues.Unlock()

		log.Debugf("SQS Publish Success - Topic: %s(%s), Message: %s\n", topic.Name, queueName, msg.MessageBody)
	} else {
		log.Warnf("SQS Publish Failure - Queue %s does not exist, message discarded\n", queueName)
	}
	return nil
}

var publishSqsMessageFunc = publishSQS
var publishHttpMessageFunc = publishHTTP
var publishMessageByTopicFunc = publishMessageByTopic

// publishMessageByTopic - we'll return an error (the last one) at the end if any of the publishes fail.  For
// now, we won't worry about cataloging each one even though, if you have multiple failures, we'll stomp on the
// first ones.  For the current callers, just knowing that any failed will consider that entry failed.
// We will also consider it a success if you have no subscriptions. "You didn't ask us to do anything, so we won't."
func publishMessageByTopic(topic *models.Topic, message interfaces.AbstractPublishEntry) (messageId string, err error) {
	messageId = uuid.NewString()
	for _, sub := range topic.Subscriptions {
		switch models.Protocol(sub.Protocol) {
		case models.ProtocolSQS:
			err = publishSqsMessageFunc(sub, topic, message)
			if err != nil {
				log.WithFields(log.Fields{"Topic": topic.Name, "Queue": sub.EndPoint}).Warn("Failed to publish message through subscription")
			}
		case models.ProtocolHTTP:
			fallthrough
		case models.ProtocolHTTPS:
			publishHttpMessageFunc(sub, topic.Arn, message)
		}
	}
	return messageId, err
}
