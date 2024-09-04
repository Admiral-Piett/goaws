package router

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	log "github.com/sirupsen/logrus"

	"fmt"

	sns "github.com/Admiral-Piett/goaws/app/gosns"
	sqs "github.com/Admiral-Piett/goaws/app/gosqs"
	"github.com/gorilla/mux"
)

// New returns a new router
func New() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", actionHandler).Methods("GET", "POST")
	r.HandleFunc("/health", health).Methods("GET")
	r.HandleFunc("/{account}", actionHandler).Methods("GET", "POST")
	r.HandleFunc("/queue/{queueName}", actionHandler).Methods("GET", "POST")
	r.HandleFunc("/SimpleNotificationService/{id}.pem", pemHandler).Methods("GET")
	r.HandleFunc("/{account}/{queueName}", actionHandler).Methods("GET", "POST")

	return r
}

func encodeResponse(w http.ResponseWriter, req *http.Request, statusCode int, body interfaces.AbstractResponseBody) {
	protocol := resolveProtocol(req)
	switch protocol {
	case AwsJsonProtocol:
		w.Header().Set("x-amzn-RequestId", body.GetRequestId())
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		// Stupidly these `WriteHeader` calls have to be here, if they're at the start
		// they lock the headers, at the end they're ignored.
		w.WriteHeader(statusCode)
		if body.GetResult() == nil {
			return
		}
		err := json.NewEncoder(w).Encode(body.GetResult())
		if err != nil {
			log.Errorf("Response Encoding Error: %v\nResponse: %+v", err, body)
			http.Error(w, "General Error", http.StatusInternalServerError)
		}
	case AwsQueryProtocol:
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(statusCode)
		result, err := xml.Marshal(body)
		if err != nil {
			log.Errorf("Response Encoding Error: %v\nResponse: %+v", err, body)
			http.Error(w, "General Error", http.StatusInternalServerError)
		}
		_, _ = w.Write(result)
	}
}

// V1 - includes JSON Support (and of course the old XML).
var routingTableV1 = map[string]func(r *http.Request) (int, interfaces.AbstractResponseBody){
	// SQS
	"CreateQueue":             sqs.CreateQueueV1,
	"ListQueues":              sqs.ListQueuesV1,
	"GetQueueAttributes":      sqs.GetQueueAttributesV1,
	"SetQueueAttributes":      sqs.SetQueueAttributesV1,
	"SendMessage":             sqs.SendMessageV1,
	"ReceiveMessage":          sqs.ReceiveMessageV1,
	"ChangeMessageVisibility": sqs.ChangeMessageVisibilityV1,
	"DeleteMessage":           sqs.DeleteMessageV1,
	"GetQueueUrl":             sqs.GetQueueUrlV1,
	"PurgeQueue":              sqs.PurgeQueueV1,
	"DeleteQueue":             sqs.DeleteQueueV1,
	"SendMessageBatch":        sqs.SendMessageBatchV1,
	"DeleteMessageBatch":      sqs.DeleteMessageBatchV1,

	// SNS
	"Subscribe":                 sns.SubscribeV1,
	"Unsubscribe":               sns.UnsubscribeV1,
	"Publish":                   sns.PublishV1,
	"ListTopics":                sns.ListTopicsV1,
	"CreateTopic":               sns.CreateTopicV1,
	"DeleteTopic":               sns.DeleteTopicV1,
	"ListSubscriptions":         sns.ListSubscriptionsV1,
	"GetSubscriptionAttributes": sns.GetSubscriptionAttributesV1,
	"SetSubscriptionAttributes": sns.SetSubscriptionAttributesV1,
	"ListSubscriptionsByTopic":  sns.ListSubscriptionsByTopicV1,

	// SNS Internal
	"ConfirmSubscription": sns.ConfirmSubscriptionV1,
}

func health(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	fmt.Fprint(w, "OK")
}

func actionHandler(w http.ResponseWriter, req *http.Request) {
	action := extractAction(req)
	log.WithFields(
		log.Fields{
			"action": action,
			"url":    req.URL,
		}).Debug("Handling URL request")
	// If we don't find a match in this table, pass on to the existing flow.
	jsonFn, ok := routingTableV1[action]
	if ok {
		statusCode, responseBody := jsonFn(req)
		encodeResponse(w, req, statusCode, responseBody)
		return
	}
	log.Println("Bad Request - Action:", action)
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, "Bad Request")
}

func pemHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(sns.PemKEY)
}

type AwsProtocol int

const (
	AwsJsonProtocol  AwsProtocol = iota
	AwsQueryProtocol AwsProtocol = iota
)

// Extract target Action from the request.
// How contains the Action name is different with aws-query protocol and aws-json protocol.
func extractAction(req *http.Request) string {
	protocol := resolveProtocol(req)
	switch protocol {
	case AwsJsonProtocol:
		// Get action from X-Amz-Target header
		action := req.Header.Get("X-Amz-Target")
		// Action value will be like as "AmazonSQS.CreateQueue".
		// After dot should be the action name.
		return strings.Split(action, ".")[1]
	case AwsQueryProtocol:
		return req.FormValue("Action")
	}
	return ""
}

// Determine which protocol is used.
func resolveProtocol(req *http.Request) AwsProtocol {
	// Use content-type to determine protocol
	if req.Header.Get("Content-Type") == "application/x-amz-json-1.0" {
		return AwsJsonProtocol
	}
	return AwsQueryProtocol
}
