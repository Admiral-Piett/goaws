package gosns

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Admiral-Piett/goaws/app/models"

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

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/common"
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
	app.SyncTopics.Topics = make(map[string]*app.Topic)
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

func signMessage(privkey *rsa.PrivateKey, snsMsg *app.SNSMessage) (string, error) {
	fs, err := formatSignature(snsMsg)
	if err != nil {
		return "", nil
	}

	h := sha1.Sum([]byte(fs))
	signature_b, err := rsa.SignPKCS1v15(rand.Reader, privkey, crypto.SHA1, h[:])

	return base64.StdEncoding.EncodeToString(signature_b), err
}

func formatSignature(msg *app.SNSMessage) (formated string, err error) {
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

func ConfirmSubscription(w http.ResponseWriter, req *http.Request) {
	topicArn := req.Form.Get("TopicArn")
	confirmToken := req.Form.Get("Token")
	pendingConfirm := TOPIC_DATA[topicArn]
	if pendingConfirm.token == confirmToken {
		respStruct := models.ConfirmSubscriptionResponse{"http://queue.amazonaws.com/doc/2012-11-05/", models.SubscribeResult{SubscriptionArn: pendingConfirm.subArn}, app.ResponseMetadata{RequestId: uuid.NewString()}}

		SendResponseBack(w, req, respStruct, "application/xml")
	} else {
		createErrorResponse(w, req, "SubArnNotFound")
	}

}

func ListSubscriptionsByTopic(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	topicArn := req.FormValue("TopicArn")

	uriSegments := strings.Split(topicArn, ":")
	topicName := uriSegments[len(uriSegments)-1]

	if topic, ok := app.SyncTopics.Topics[topicName]; ok {
		uuid, _ := common.NewUUID()
		respStruct := app.ListSubscriptionsByTopicResponse{}
		respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
		respStruct.Metadata.RequestId = uuid
		respStruct.Result.Subscriptions.Member = make([]app.TopicMemberResult, 0, 0)

		for _, sub := range topic.Subscriptions {
			tar := app.TopicMemberResult{TopicArn: topic.Arn, Protocol: sub.Protocol,
				SubscriptionArn: sub.SubscriptionArn, Endpoint: sub.EndPoint, Owner: app.CurrentEnvironment.AccountID}
			respStruct.Result.Subscriptions.Member = append(respStruct.Result.Subscriptions.Member, tar)
		}
		SendResponseBack(w, req, respStruct, content)
	} else {
		createErrorResponse(w, req, "TopicNotFound")
	}
}

func SetSubscriptionAttributes(w http.ResponseWriter, req *http.Request) {
	content := req.FormValue("ContentType")
	subsArn := req.FormValue("SubscriptionArn")
	Attribute := req.FormValue("AttributeName")
	Value := req.FormValue("AttributeValue")

	for _, topic := range app.SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subsArn {
				if Attribute == "RawMessageDelivery" {
					app.SyncTopics.Lock()
					if Value == "true" {
						sub.Raw = true
					} else {
						sub.Raw = false
					}
					app.SyncTopics.Unlock()
					//Good Response == return
					uuid, _ := common.NewUUID()
					respStruct := app.SetSubscriptionAttributesResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: uuid}}
					SendResponseBack(w, req, respStruct, content)
					return
				}

				if Attribute == "FilterPolicy" {
					filterPolicy := &app.FilterPolicy{}
					err := json.Unmarshal([]byte(Value), filterPolicy)
					if err != nil {
						createErrorResponse(w, req, "ValidationError")
						return
					}

					app.SyncTopics.Lock()
					sub.FilterPolicy = filterPolicy
					app.SyncTopics.Unlock()

					//Good Response == return
					uuid, _ := common.NewUUID()
					respStruct := app.SetSubscriptionAttributesResponse{"http://queue.amazonaws.com/doc/2012-11-05/", app.ResponseMetadata{RequestId: uuid}}
					SendResponseBack(w, req, respStruct, content)
					return
				}

			}
		}
	}
	createErrorResponse(w, req, "SubscriptionNotFound")
}

func GetSubscriptionAttributes(w http.ResponseWriter, req *http.Request) {

	content := req.FormValue("ContentType")
	subsArn := req.FormValue("SubscriptionArn")

	for _, topic := range app.SyncTopics.Topics {
		for _, sub := range topic.Subscriptions {
			if sub.SubscriptionArn == subsArn {

				entries := make([]app.SubscriptionAttributeEntry, 0, 0)
				entry := app.SubscriptionAttributeEntry{Key: "Owner", Value: app.CurrentEnvironment.AccountID}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "RawMessageDelivery", Value: strconv.FormatBool(sub.Raw)}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "TopicArn", Value: sub.TopicArn}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "Endpoint", Value: sub.EndPoint}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "PendingConfirmation", Value: "false"}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "ConfirmationWasAuthenticated", Value: "true"}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "SubscriptionArn", Value: sub.SubscriptionArn}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "Protocol", Value: sub.Protocol}
				entries = append(entries, entry)
				entry = app.SubscriptionAttributeEntry{Key: "Endpoint", Value: sub.EndPoint}
				entries = append(entries, entry)

				if sub.FilterPolicy != nil {
					filterPolicyBytes, _ := json.Marshal(sub.FilterPolicy)
					entry = app.SubscriptionAttributeEntry{Key: "FilterPolicy", Value: string(filterPolicyBytes)}
					entries = append(entries, entry)
				}

				result := app.GetSubscriptionAttributesResult{SubscriptionAttributes: app.SubscriptionAttributes{Entries: entries}}
				uuid, _ := common.NewUUID()
				respStruct := app.GetSubscriptionAttributesResponse{"http://sns.amazonaws.com/doc/2010-03-31", result, app.ResponseMetadata{RequestId: uuid}}

				SendResponseBack(w, req, respStruct, content)

				return
			}
		}
	}
	createErrorResponse(w, req, "SubscriptionNotFound")
}

// NOTE: The use case for this is to use GoAWS to call some external system with the message payload.  Essentially
// it is a localized subscription to some non-AWS endpoint.

func callEndpoint(endpoint string, subArn string, msg app.SNSMessage, raw bool) error {
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

	defaultMsg, ok := msgWithProtocols[string(app.ProtocolDefault)]
	if !ok {
		return "", errors.New(app.ErrNoDefaultElementInJSON)
	}

	if m, ok := msgWithProtocols[protocol]; ok {
		return m, nil
	}

	return defaultMsg, nil
}

func createErrorResponse(w http.ResponseWriter, req *http.Request, err string) {
	er := models.SnsErrors[err]
	respStruct := models.ErrorResponse{
		Result:    models.ErrorResult{Type: er.Type, Code: er.Code, Message: er.Message},
		RequestId: "00000000-0000-0000-0000-000000000000",
	}

	w.WriteHeader(er.HttpError)
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func SendResponseBack(w http.ResponseWriter, req *http.Request, respStruct interface{}, content string) {
	if content == "JSON" {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		if err := enc.Encode(respStruct); err != nil {
			log.Printf("error: %v\n", err)
		}
	} else {
		w.Header().Set("Content-Type", "application/xml")
		enc := xml.NewEncoder(w)
		enc.Indent("  ", "    ")
		if err := enc.Encode(respStruct); err != nil {
			log.Printf("error: %v\n", err)
		}
	}
}
