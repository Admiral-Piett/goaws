package main

import (
	"flag"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"

	sns "github.com/p4tin/goaws/gosns"
	sqs "github.com/p4tin/goaws/gosqs"
	"github.com/p4tin/goaws/conf"
	"github.com/spf13/pflag"
)

func BadRequest(w http.ResponseWriter, req *http.Request) {
	resp := "Bad Request"
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, resp)
}

// hello world, the web server
func IndexServer(w http.ResponseWriter, req *http.Request) {
	switch req.FormValue("Action") {
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
	case "ListSubscriptionsByTopic":
		sns.ListSubscriptionsByTopic(w, req)
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

	var snsPortNumber string
	var sqsPortNumber string
	var filename string
	flag.StringVar(&sqsPortNumber, "sqs-port", "", "Port number SQS listens on")
	flag.StringVar(&snsPortNumber, "sns-port", "", "Port number SNS listens on")
	flag.StringVar(&filename, "config", "", "config file location + name")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()

	loadedEnv, err := conf.LoadYamlConfig(filename, env, sqsPortNumber, snsPortNumber)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", IndexServer).Methods("GET", "POST")
	r.HandleFunc("/queue/{queueName}", IndexServer).Methods("GET", "POST")

	go func(port string) {
		log.Printf("GoAws SQS listening on: 0.0.0.0:%s\n", port)
		err := http.ListenAndServe(":"+port, r)
		log.Fatal(err)
	}(loadedEnv.SQSPort)

	log.Printf("GoAws SNS listening on: 0.0.0.0:%s\n", loadedEnv.SNSPort)
	err = http.ListenAndServe(":"+loadedEnv.SNSPort, r)
	log.Fatal(err)
}
