package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"io"
	"flag"

	sqs "github.com/p4tin/goaws/gosqs"
	sns "github.com/p4tin/goaws/gosns"
	"log"
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

	/*** Bad Request ***/
	default:
		BadRequest(w, req)
	}
}

func main() {

	var portNumber string
	flag.StringVar(&portNumber, "port", "4100", "Port number to listen on")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", IndexServer).Methods("GET", "POST")

	log.Printf("GoAws listening on: 0.0.0.0:%s\n", portNumber)
	err := http.ListenAndServe("0.0.0.0:"+portNumber, r)
	log.Fatal(err)
}
