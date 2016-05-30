package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"io"
	"flag"
	"log"

	sqs "github.com/p4tin/goaws/gosqs"
	sns "github.com/p4tin/goaws/gosns"
	"github.com/p4tin/goaws/conf"
	"os"
)

func BadRequest(w http.ResponseWriter, req *http.Request) {
	resp := "Bad Request"
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, resp)
}


// hello world, the web server
func IndexServer(w http.ResponseWriter, req *http.Request) {
	switch(req.FormValue("Action")) {
	/*** SQS Actions ***/
	case "ListQueues":
		sqs.ListQueues(w, req)
	case "CreateQueue":
		sqs.CreateQueue(w, req)
	case "GetQueueAttributes":
		sqs.GetQueueAttributes(w, req)
	case "SendMessage":
		sqs.SendMessage(w, req)
	case "ReceiveMessage":
		sqs.ReceiveMessage(w, req)
	case "DeleteMessage":
		sqs.DeleteMessage(w, req)
	case "GetQueueUrl":
		sqs.GetQueueUrl(w, req)
	case "PurgeQueue":
		sqs.PurgeQueue(w, req)
	case "DeleteQueue":
		sqs.DeleteQueue(w, req)

	/*** SNS Actions ***/
	case "ListTopics":
		sns.ListTopics(w, req)
	/*** SNS Actions ***/
	case "CreateTopic":
		sns.CreateTopic(w, req)
	case "DeleteTopic":
		sns.DeleteTopic(w, req)
	case "Subscribe":
		sns.Subscribe(w, req)
	case "SetSubscriptionAttributes":
		sns.SetSubscriptionAttributes(w, req)
	case "ListSubscriptions":
		sns.ListSubscriptions(w, req)
	case "Unsubscribe":
		sns.Unsubscribe(w, req)
	case "Publish":
		sns.Publish(w, req)

	/*** Bad Request ***/
	default:
		log.Println("Action:", req.FormValue("Action"))
		BadRequest(w, req)
	}
}

func main() {
	env := "Local"
	if len(os.Args) == 2 {
		env = os.Args[1]
	}
	var portNumber string
	flag.StringVar(&portNumber, "port", "", "Port number to listen on")
	flag.Parse()

	portNumber = conf.LoadYamlConfig(env, portNumber)

	r := mux.NewRouter()
	r.HandleFunc("/", IndexServer).Methods("GET", "POST")
	r.HandleFunc("/queue/{queueName}", IndexServer).Methods("GET", "POST")

	log.Printf("GoAws listening on: 0.0.0.0:%s\n", portNumber)
	err := http.ListenAndServe("0.0.0.0:"+portNumber, r)
	log.Fatal(err)
}
