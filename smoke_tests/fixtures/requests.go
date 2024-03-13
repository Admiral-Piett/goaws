package fixtures

import (
	"fmt"

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

var GetQueueAttributesRequestBodyXML = struct {
	Action     string `xml:"Action"`
	Version    string `xml:"Version"`
	Attribute1 string `xml:"AttributeName.1"`
	QueueUrl   string `xml:"QueueUrl"`
}{
	Action:     "GetQueueAttributes",
	Version:    "2012-11-05",
	Attribute1: "All",
	QueueUrl:   fmt.Sprintf("%s/new-queue-1", af.BASE_URL),
}

var CreateQueueV1RequestBodyJSON = models.CreateQueueRequest{
	QueueName:  af.QueueName,
	Version:    "2012-11-05",
	Attributes: af.CreateQueueAttributes,
}

var CreateQueueV1RequestXML_NoAttributes = struct {
	Action    string `xml:"Action"`
	Version   string `xml:"Version"`
	QueueName string `xml:"QueueName"`
}{
	Action:    "CreateQueue",
	Version:   "2012-11-05",
	QueueName: af.QueueName,
}
