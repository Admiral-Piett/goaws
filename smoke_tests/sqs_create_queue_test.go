package smoke_tests

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/stretchr/testify/assert"

	sf "github.com/Admiral-Piett/goaws/smoke_tests/fixtures"

	"github.com/gavv/httpexpect/v2"
)

func Test_CreateQueueV1_json_no_attributes(t *testing.T) {
	server := generateServer()
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	r := e.POST("/").
		WithHeaders(map[string]string{
			"Content-Type": "application/x-amz-json-1.0",
			"X-Amz-Target": "AmazonSQS.CreateQueue",
		}).
		WithJSON(sf.CreateQueueV1RequestBodyJSON).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp1 := app.CreateQueueResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Result:   app.CreateQueueResult{QueueUrl: fmt.Sprintf("%s/new-queue-1", sf.BASE_URL)},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}

	r1 := app.CreateQueueResponse{}
	xml.Unmarshal([]byte(r), &r1)
	assert.Equal(t, exp1, r1)

	r = e.POST("/").
		WithForm(sf.ListQueuesRequestBodyXML).
		Expect().
		Status(http.StatusOK).
		Body().Raw()

	exp2 := app.ListQueuesResponse{
		Xmlns:    "http://queue.amazonaws.com/doc/2012-11-05/",
		Result:   app.ListQueuesResult{QueueUrl: []string{fmt.Sprintf("%s/new-queue-1", sf.BASE_URL)}},
		Metadata: app.ResponseMetadata{RequestId: sf.REQUEST_ID},
	}
	r2 := app.ListQueuesResponse{}
	xml.Unmarshal([]byte(r), &r2)
	assert.Equal(t, exp2, r2)
}

func Test_CreateQueueV1_json_with_attributes(t *testing.T) {
}

func Test_CreateQueueV1_json_with_attributes_ints_as_strings(t *testing.T) {
}

func Test_CreateQueueV1_xml_no_attributes(t *testing.T) {
}

func Test_CreateQueueV1_xml_with_attributes(t *testing.T) {
}

func Test_CreateQueueV1_xml_with_attributes_ints_as_strings(t *testing.T) {
	// TODO - you will have to escape quote them in the form I think
}
