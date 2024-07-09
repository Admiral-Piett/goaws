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

var ListTopicsRequestBodyXML = struct {
	Action  string `xml:"Action"`
	Version string `xml:"Version"`
}{
	Action:  "ListTopics",
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
	QueueUrl:   af.QueueUrl,
}

var SetQueueAttributesRequestBodyXML = struct {
	Action   string `xml:"Action"`
	Version  string `xml:"Version"`
	QueueUrl string `xml:"QueueUrl"`
}{
	Action:   "SetQueueAttributes",
	Version:  "2012-11-05",
	QueueUrl: af.QueueUrl,
}

var CreateQueueV1RequestBodyJSON = models.CreateQueueRequest{
	QueueName:  af.QueueName,
	Version:    "2012-11-05",
	Attributes: af.QueueAttributes,
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

var SendMessageRequestBodyXML = struct {
	Action      string `xml:"Action"`
	QueuUrl     string `xml:"QueueUrl"`
	MessageBody string `xml:"MessageBody"`
	Version     string `xml:"Version"`
}{
	Action:      "SendMessage",
	QueuUrl:     fmt.Sprintf("%s/new-queue-1", af.BASE_URL),
	MessageBody: "Hello World",
	Version:     "2012-11-05",
}

var ReceiveMessageRequestBodyXML = struct {
	Action  string `xml:"Action"`
	QueuUrl string `xml:"QueueUrl"`
	Version string `xml:"Version"`
}{
	Action:  "ReceiveMessage",
	QueuUrl: fmt.Sprintf("%s/new-queue-1", af.BASE_URL),
	Version: "2012-11-05",
}

var ChangeMessageVisibilityRequestBodyXML = struct {
	Action            string `xml:"Action"`
	QueuUrl           string `xml:"QueueUrl"`
	ReceiptHandle     string `xml:"ReceiptHandle"`
	VisibilityTimeout int    `xml:"VisibilityTimeout"`
	Version           string `xml:"Version"`
}{
	Action:            "ChangeMessageVisibility",
	QueuUrl:           fmt.Sprintf("%s/new-queue-1", af.BASE_URL),
	VisibilityTimeout: 2,
	Version:           "2012-11-05",
}

var DeleteMessageRequestBodyXML = struct {
	Action        string `xml:"Action"`
	QueuUrl       string `xml:"QueueUrl"`
	ReceiptHandle string `xml:"ReceiptHandle"`
	Version       string `xml:"Version"`
}{
	Action:  "DeleteMessage",
	QueuUrl: fmt.Sprintf("%s/new-queue-1", af.BASE_URL),
	Version: "2012-11-05",
}
