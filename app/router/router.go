package router

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"

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

func encodeResponse(w http.ResponseWriter, statusCode int, body interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)
	// TODO - replace with gorilla/schema
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(body); err != nil {
		log.Errorf("error: %v\n", err)
	}
}

// V1 - includes JSON Support (and of course the old XML).
var routingTableV1 = map[string]func(r *http.Request) (int, interface{}){
	"CreateQueue": sqs.CreateQueueV1,
}

var routingTable = map[string]http.HandlerFunc{
	// SQS
	"ListQueues": sqs.ListQueues,
	//"CreateQueue":             sqs.CreateQueue,
	"GetQueueAttributes":      sqs.GetQueueAttributes,
	"SetQueueAttributes":      sqs.SetQueueAttributes,
	"SendMessage":             sqs.SendMessage,
	"SendMessageBatch":        sqs.SendMessageBatch,
	"ReceiveMessage":          sqs.ReceiveMessage,
	"DeleteMessage":           sqs.DeleteMessage,
	"DeleteMessageBatch":      sqs.DeleteMessageBatch,
	"GetQueueUrl":             sqs.GetQueueUrl,
	"PurgeQueue":              sqs.PurgeQueue,
	"DeleteQueue":             sqs.DeleteQueue,
	"ChangeMessageVisibility": sqs.ChangeMessageVisibility,

	// SNS
	"ListTopics":                sns.ListTopics,
	"CreateTopic":               sns.CreateTopic,
	"DeleteTopic":               sns.DeleteTopic,
	"Subscribe":                 sns.Subscribe,
	"SetSubscriptionAttributes": sns.SetSubscriptionAttributes,
	"GetSubscriptionAttributes": sns.GetSubscriptionAttributes,
	"ListSubscriptionsByTopic":  sns.ListSubscriptionsByTopic,
	"ListSubscriptions":         sns.ListSubscriptions,
	"Unsubscribe":               sns.Unsubscribe,
	"Publish":                   sns.Publish,

	// SNS Internal
	"ConfirmSubscription": sns.ConfirmSubscription,
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
		encodeResponse(w, statusCode, responseBody)
		return
	}
	fn, ok := routingTable[action]
	if !ok {
		log.Println("Bad Request - Action:", action)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Bad Request")
		return
	}

	http.HandlerFunc(fn).ServeHTTP(w, req)
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
