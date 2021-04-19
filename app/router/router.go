package router

import (
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"fmt"

	"github.com/gorilla/mux"
	sns "github.com/p4tin/goaws/app/gosns"
	sqs "github.com/p4tin/goaws/app/gosqs"
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

var routingTable = map[string]http.HandlerFunc{
	// SQS
	"ListQueues":              sqs.ListQueues,
	"CreateQueue":             sqs.CreateQueue,
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
	//	"ChangeMessageVisibilityBatch": sqs.ChangeMessageVisibilityBatch,
	//	"AddPermission": sqs.AddPermission,
	//	"RemovePermission": sqs.RemovePermission,
	//	"ListDeadLetterSourceQueues": sqs.ListDeadLetterSourceQueues,
	"ListQueueTags": sqs.ListQueueTags,
	"TagQueue":      sqs.TagQueue,
	"UntagQueue":    sqs.UntagQueue,

	// SNS
	"ListTopics":                sns.ListTopics,
	"CreateTopic":               sns.CreateTopic,
	"DeleteTopic":               sns.DeleteTopic,
	"Subscribe":                 sns.Subscribe,
	"GetTopicAttributes":        sns.GetTopicAttributes,
	"ConfirmSubscription":       sns.ConfirmSubscription,
	"SetSubscriptionAttributes": sns.SetSubscriptionAttributes,
	"GetSubscriptionAttributes": sns.GetSubscriptionAttributes,
	"ListSubscriptionsByTopic":  sns.ListSubscriptionsByTopic,
	"ListSubscriptions":         sns.ListSubscriptions,
	"Unsubscribe":               sns.Unsubscribe,
	"Publish":                   sns.Publish,
}

func health(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	fmt.Fprint(w, "OK")
}

func actionHandler(w http.ResponseWriter, req *http.Request) {
	log.WithFields(
		log.Fields{
			"action": req.FormValue("Action"),
			"url":    req.URL,
		}).Debug("Handling URL request")
	fn, ok := routingTable[req.FormValue("Action")]
	if !ok {
		log.Println("Bad Request - Action:", req.FormValue("Action"))
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
