package gosns

import (
	"net/http"
	"fmt"
	"log"
	"sync"
	"encoding/xml"
)

type Topic struct {
	Name 		string
	Arn 		string
}


var SyncTopics = struct{
	sync.RWMutex
	Topics map[string]*Topic
}{Topics: make(map[string]*Topic)}

func init() {
	SyncTopics.Topics = make(map[string]*Topic)
	SyncTopics.Topics["topic1"] = &Topic{ Name: "topic1", Arn: "arn:aws:sns:local:000000000000:topic1" }
	SyncTopics.Topics["topic2"] = &Topic{ Name: "topic2", Arn: "arn:aws:sns:local:000000000000:topic2" }
}

func ListTopics(w http.ResponseWriter, req *http.Request) {
	respStruct := ListTopicsResponse{}
	respStruct.Xmlns = "http://queue.amazonaws.com/doc/2012-11-05/"
	respStruct.Metadata = ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000000"}

	respStruct.Result.Topics.Member = make([]TopicArnResult, 0, 0)
	log.Println("Listing Topics")
	for _, topic := range SyncTopics.Topics {
		ta := TopicArnResult{TopicArn: topic.Arn}
		respStruct.Result.Topics.Member = append(respStruct.Result.Topics.Member, ta)
	}
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
