package gosns

import (
	"net/http"
	"io"
	"fmt"
)

type Topic struct {
	Name 		string
	Arn 		string
	Subscritions 	[]string
}

var Topics map[string]Topic

func init() {
	Topics = make(map[string]Topic)
	Topics["topic1"] = Topic{Name: "topic1"}
	Topics["topic2"] = Topic{Name: "topic2"}
}

func ListTopics(w http.ResponseWriter, req *http.Request) {
	resp := ""
	for _, topic := range Topics {
		resp = resp + fmt.Sprintf("%s<br>", topic.Name)
	}
	io.WriteString(w, resp)
}
