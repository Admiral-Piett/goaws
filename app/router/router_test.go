package router

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	sns "github.com/Admiral-Piett/goaws/app/gosns"

	sqs "github.com/Admiral-Piett/goaws/app/gosqs"

	"github.com/stretchr/testify/assert"

	"github.com/Admiral-Piett/goaws/app/utils"
)

func TestMain(m *testing.M) {
	utils.InitializeDecoders()
	m.Run()
}

func TestIndexServerhandler_POST_BadRequest(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "BadRequest")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	New().ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestIndexServerhandler_POST_GoodRequest(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "ListTopics")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	New().ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestIndexServerhandler_POST_GoodRequest_With_URL(t *testing.T) {

	req, err := http.NewRequest("POST", "/100010001000/local-queue1", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "local-queue1")
	req.PostForm = form
	rr := httptest.NewRecorder()
	New().ServeHTTP(rr, req)

	form = url.Values{}
	form.Add("Action", "GetQueueAttributes")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	New().ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestIndexServerhandler_POST_GoodRequest_With_URL_And_Aws_Json_Protocol(t *testing.T) {
	json, _ := json.Marshal(map[string]string{
		"QueueName": "local-queue1",
	})
	req, err := http.NewRequest("POST", "/100010001000/local-queue1", bytes.NewBuffer(json))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Amz-Target", "AmazonSQS.CreateQueue")
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")

	rr := httptest.NewRecorder()

	New().ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestIndexServerhandler_GET_GoodRequest_Pem_cert(t *testing.T) {

	req, err := http.NewRequest("GET", "/SimpleNotificationService/100010001000.pem", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	New().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestEncodeResponse_success(t *testing.T) {
	w, _ := utils.GenerateRequestInfo("POST", "/url", nil, false)

	type responseStruct struct {
		Message string `xml:"message"`
	}

	encodeResponse(w, http.StatusOK, responseStruct{Message: "response-body"})

	assert.Equal(t, http.StatusOK, w.Code)

	tmp := responseStruct{}
	xml.Unmarshal(w.Body.Bytes(), &tmp)
	assert.Equal(t, responseStruct{Message: "response-body"}, tmp)
}

func TestEncodeResponse_success_skips_malformed_body(t *testing.T) {
	w, _ := utils.GenerateRequestInfo("POST", "/url", nil, false)

	encodeResponse(w, http.StatusOK, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, &bytes.Buffer{}, w.Body)
}

func TestActionHandler_json(t *testing.T) {
	type responseStruct struct {
		Message string `xml:"message"`
	}
	defer func() {
		routingTableV1 = map[string]func(r *http.Request) (int, interface{}){
			"CreateQueue": sqs.CreateQueueV1,
		}
	}()

	mockCalled := false
	mockFunction := func(req *http.Request) (int, interface{}) {
		mockCalled = true
		return http.StatusOK, responseStruct{Message: "response-body"}
	}
	routingTableV1 = map[string]func(r *http.Request) (int, interface{}){
		"CreateQueue": mockFunction,
	}

	w, r := utils.GenerateRequestInfo("POST", "/url", nil, true)
	r.Header.Set("X-Amz-Target", "QueueService.CreateQueue")

	actionHandler(w, r)

	assert.True(t, mockCalled)
	assert.Equal(t, http.StatusOK, w.Code)

	tmp := responseStruct{}
	xml.Unmarshal(w.Body.Bytes(), &tmp)
	assert.Equal(t, responseStruct{Message: "response-body"}, tmp)
}

func TestActionHandler_xml(t *testing.T) {
	defer func() {
		routingTableV1 = map[string]func(r *http.Request) (int, interface{}){
			"CreateQueue": sqs.CreateQueueV1,
		}
		routingTable = map[string]http.HandlerFunc{
			// SQS
			"ListQueues": sqs.ListQueues,
			//"CreateQueue":             sqs.CreateQueue,
			"GetQueueAttributes":      sqs.GetQueueAttributes,
			"SetQueueAttributes":      sqs.SetQueueAttributes,
			"SendMessage":             sqs.SendMessage,
			"SendMessageBatch":        sqs.SendMessageBatch,
			"ReceiveMessage":          sqs.ReceiveMessage,
			"DeleteMessage":           sqs.DeleteMessage,
			"DeleteMessageBatch":      sqs.DeleteMessageBatch,
			"GetQueueUrl":             sqs.GetQueueUrl,
			"PurgeQueue":              sqs.PurgeQueue,
			"DeleteQueue":             sqs.DeleteQueue,
			"ChangeMessageVisibility": sqs.ChangeMessageVisibility,

			// SNS
			"ListTopics":                sns.ListTopics,
			"CreateTopic":               sns.CreateTopic,
			"DeleteTopic":               sns.DeleteTopic,
			"Subscribe":                 sns.Subscribe,
			"SetSubscriptionAttributes": sns.SetSubscriptionAttributes,
			"GetSubscriptionAttributes": sns.GetSubscriptionAttributes,
			"ListSubscriptionsByTopic":  sns.ListSubscriptionsByTopic,
			"ListSubscriptions":         sns.ListSubscriptions,
			"Unsubscribe":               sns.Unsubscribe,
			"Publish":                   sns.Publish,

			// SNS Internal
			"ConfirmSubscription": sns.ConfirmSubscription,
		}
	}()

	mockCalled := false
	mockFunction := func(w http.ResponseWriter, req *http.Request) {
		mockCalled = true
		w.WriteHeader(http.StatusOK)
	}
	routingTableV1 = map[string]func(r *http.Request) (int, interface{}){}
	routingTable = map[string]http.HandlerFunc{
		"CreateQueue": mockFunction,
	}

	w, r := utils.GenerateRequestInfo("POST", "/url", nil, false)
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	r.PostForm = form

	actionHandler(w, r)

	assert.True(t, mockCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}
