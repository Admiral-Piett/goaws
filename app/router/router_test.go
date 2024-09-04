package router

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	af "github.com/Admiral-Piett/goaws/app/fixtures"

	"github.com/Admiral-Piett/goaws/app/mocks"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	sqs "github.com/Admiral-Piett/goaws/app/gosqs"

	"github.com/stretchr/testify/assert"

	"github.com/Admiral-Piett/goaws/app/test"
)

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
	form.Add("QueueUrl", fmt.Sprintf("%s/local-queue1", af.BASE_URL))
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	New().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
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

func TestEncodeResponse_success_xml(t *testing.T) {
	w, r := test.GenerateRequestInfo("POST", "/url", nil, false)

	encodeResponse(w, r, http.StatusOK, mocks.BaseResponse{Message: "test"})

	assert.Equal(t, http.StatusOK, w.Code)

	tmp := mocks.BaseResponse{}
	xml.Unmarshal(w.Body.Bytes(), &tmp)
	assert.Equal(t, mocks.BaseResponse{Message: "test"}, tmp)
}

func TestEncodeResponse_success_skips_nil_body_xml(t *testing.T) {
	w, r := test.GenerateRequestInfo("POST", "/url", nil, false)

	encodeResponse(w, r, http.StatusOK, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, &bytes.Buffer{}, w.Body)
}

func TestEncodeResponse_success_json(t *testing.T) {
	w, r := test.GenerateRequestInfo("POST", "/url", nil, true)

	encodeResponse(w, r, http.StatusOK, mocks.BaseResponse{Message: "test"})

	assert.Equal(t, http.StatusOK, w.Code)

	tmp := mocks.BaseResponse{}
	json.Unmarshal(w.Body.Bytes(), &tmp)
	assert.Equal(t, mocks.BaseResponse{Message: "test"}, tmp)
}

func TestEncodeResponse_success_skips_malformed_body_json(t *testing.T) {
	mock := mocks.BaseResponse{
		Message: "test",
	}
	mock.MockGetResult = func() interface{} {
		return make(chan int)
	}
	w, r := test.GenerateRequestInfo("POST", "/url", nil, true)

	encodeResponse(w, r, http.StatusOK, mock)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "General Error", strings.TrimSpace(string(w.Body.Bytes())))
}

func TestActionHandler_v1_json(t *testing.T) {
	defer func() {
		routingTableV1 = map[string]func(r *http.Request) (int, interfaces.AbstractResponseBody){
			"CreateQueue": sqs.CreateQueueV1,
		}
	}()

	mockCalled := false
	mockFunction := func(req *http.Request) (int, interfaces.AbstractResponseBody) {
		mockCalled = true
		return http.StatusOK, mocks.BaseResponse{Message: "response-body"}
	}
	routingTableV1 = map[string]func(r *http.Request) (int, interfaces.AbstractResponseBody){
		"CreateQueue": mockFunction,
	}

	w, r := test.GenerateRequestInfo("POST", "/url", nil, true)
	r.Header.Set("X-Amz-Target", "QueueService.CreateQueue")

	actionHandler(w, r)

	assert.True(t, mockCalled)
	assert.Equal(t, http.StatusOK, w.Code)

	tmp := mocks.BaseResponse{}
	json.Unmarshal(w.Body.Bytes(), &tmp)
	assert.Equal(t, mocks.BaseResponse{Message: "response-body"}, tmp)
}

func TestActionHandler_v1_xml(t *testing.T) {
	defer func() {
		routingTableV1 = map[string]func(r *http.Request) (int, interfaces.AbstractResponseBody){
			"CreateQueue": sqs.CreateQueueV1,
		}
	}()

	mockCalled := false
	mockFunction := func(req *http.Request) (int, interfaces.AbstractResponseBody) {
		mockCalled = true
		return http.StatusOK, mocks.BaseResponse{Message: "response-body"}
	}
	routingTableV1 = map[string]func(r *http.Request) (int, interfaces.AbstractResponseBody){
		"CreateQueue": mockFunction,
	}

	w, r := test.GenerateRequestInfo("POST", "/url", nil, false)
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	r.PostForm = form

	actionHandler(w, r)

	assert.True(t, mockCalled)
	assert.Equal(t, http.StatusOK, w.Code)

	tmp := mocks.BaseResponse{}
	xml.Unmarshal(w.Body.Bytes(), &tmp)
	assert.Equal(t, mocks.BaseResponse{Message: "response-body"}, tmp)
}
