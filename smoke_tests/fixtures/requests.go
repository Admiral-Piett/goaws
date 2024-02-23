package fixtures

import (
	af "github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/models"
)

var ListQueuesRequestBodyXML = struct {
	Action  string `xml:"Action"`
	Version string `xml:"Version"`
}{
	Action:  "ListQueues",
	Version: "2012-11-05",
}

var CreateQueueV1RequestBodyJSON = struct {
	QueueName  string `json:"QueueName"`
	Action     string `json:"Action"`
	Version    string `json:"Version"`
	Attributes models.Attributes
}{
	QueueName:  af.QueueName,
	Action:     "CreateQueue",
	Version:    "2012-11-05",
	Attributes: af.CreateQueueAttributes,
}
