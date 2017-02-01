package main

import (
	"flag"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/p4tin/goaws/app/conf"
	sns "github.com/p4tin/goaws/app/gosns"
	sqs "github.com/p4tin/goaws/app/gosqs"
)

var RoutingTable = map[string]interface{}{
	// SQS
	"ListQueues":         sqs.ListQueues,
	"CreateQueue":        sqs.CreateQueue,
	"GetQueueAttributes": sqs.GetQueueAttributes,
	"SetQueueAttributes": sqs.SetQueueAttributes,
	"SendMessage":        sqs.SendMessage,
	"ReceiveMessage":     sqs.ReceiveMessage,
	"DeleteMessage":      sqs.DeleteMessage,
	"GetQueueUrl":        sqs.GetQueueUrl,
	"PurgeQueue":         sqs.PurgeQueue,
	"DeleteQueue":        sqs.DeleteQueue,

	// SNS
	"ListTopics":                sns.ListTopics,
	"CreateTopic":               sns.CreateTopic,
	"DeleteTopic":               sns.DeleteTopic,
	"Subscribe":                 sns.Subscribe,
	"SetSubscriptionAttributes": sns.SetSubscriptionAttributes,
	"ListSubscriptionsByTopic":  sns.ListSubscriptionsByTopic,
	"ListSubscriptions":         sns.ListSubscriptions,
	"Unsubscribe":               sns.Unsubscribe,
	"Publish":                   sns.Publish,
}

func BadRequest(w http.ResponseWriter, req *http.Request) {
	resp := "Bad Request"
	w.WriteHeader(http.StatusBadRequest)
	io.WriteString(w, resp)
}

// hello world, the web server
func IndexServer(w http.ResponseWriter, req *http.Request) {
	if val, ok := RoutingTable[req.FormValue("Action")]; ok {
		val.(func(http.ResponseWriter, *http.Request))(w, req)
	} else {
		log.Println("Action:", req.FormValue("Action"))
		BadRequest(w, req)
	}

	//
	//switch req.FormValue("Action") {
	///*** SQS Actions ***/
	//case "ListQueues":
	//	sqs.ListQueues(w, req)
	//case "CreateQueue":
	//	sqs.CreateQueue(w, req)
	//case "GetQueueAttributes":
	//	sqs.GetQueueAttributes(w, req)
	//case "SetQueueAttributes":
	//	sqs.SetQueueAttributes(w, req)
	//case "SendMessage":
	//	sqs.SendMessage(w, req)
	//case "ReceiveMessage":
	//	sqs.ReceiveMessage(w, req)
	//case "DeleteMessage":
	//	sqs.DeleteMessage(w, req)
	//case "GetQueueUrl":
	//	sqs.GetQueueUrl(w, req)
	//case "PurgeQueue":
	//	sqs.PurgeQueue(w, req)
	//case "DeleteQueue":
	//	sqs.DeleteQueue(w, req)
	//
	///*** SNS Actions ***/
	//case "ListTopics":
	//	sns.ListTopics(w, req)
	//case "CreateTopic":
	//	sns.CreateTopic(w, req)
	//case "DeleteTopic":
	//	sns.DeleteTopic(w, req)
	//case "Subscribe":
	//	sns.Subscribe(w, req)
	//case "SetSubscriptionAttributes":
	//	sns.SetSubscriptionAttributes(w, req)
	//case "ListSubscriptionsByTopic":
	//	sns.ListSubscriptionsByTopic(w, req)
	//case "ListSubscriptions":
	//	sns.ListSubscriptions(w, req)
	//case "Unsubscribe":
	//	sns.Unsubscribe(w, req)
	//case "Publish":
	//	sns.Publish(w, req)

	/*** Bad Request ***/
	//default:
	//	log.Println("Action:", req.FormValue("Action"))
	//	BadRequest(w, req)
	//}
}

func main() {
	env := "Local"
	if len(os.Args) == 2 {
		env = os.Args[1]
	}

	var filename string
	flag.StringVar(&filename, "config", "", "config file location + name")
	flag.Parse()

	portNumbers := conf.LoadYamlConfig(filename, env)

	r := mux.NewRouter()
	r.HandleFunc("/", IndexServer).Methods("GET", "POST")
	r.HandleFunc("/queue/{queueName}", IndexServer).Methods("GET", "POST")

	if len(portNumbers) == 1 {
		log.Printf("GoAws listening on: 0.0.0.0:%s\n", portNumbers[0])
		err := http.ListenAndServe("0.0.0.0:"+portNumbers[0], r)
		log.Fatal(err)
	} else if len(portNumbers) == 2 {
		go func() {
			log.Printf("GoAws listening on: 0.0.0.0:%s\n", portNumbers[0])
			err := http.ListenAndServe("0.0.0.0:"+portNumbers[0], r)
			log.Fatal(err)
		}()
		log.Printf("GoAws listening on: 0.0.0.0:%s\n", portNumbers[1])
		err := http.ListenAndServe("0.0.0.0:"+portNumbers[1], r)
		log.Fatal(err)
	} else {
		log.Fatal("Not enough or too many ports defined to start GoAws.")
	}
}
