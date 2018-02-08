package gosqs

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/p4tin/goaws/app"
)

func TestListQueues_POST_NoQueues(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ListQueues)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := "<ListQueuesResult></ListQueuesResult>"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateQueuehandler_POST_CreateQueue(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "UnitTestQueue1")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateQueue)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := "UnitTestQueue1"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestSendQueue_POST_NonExistant(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/NON-EXISTANT")
	form.Add("MessageBody", "Test123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SendMessage)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body is what we expect.
	expected := "NonExistentQueue"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestSendMessageBatch_POST_NoEntry(t *testing.T) {
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SendMessageBatch)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body is what we expect.
	expected := "EmptyBatchRequest"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestSendMessageBatch_POST_IdNotDistinct(t *testing.T) {
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("SendMessageBatchRequestEntry.1.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.1.MessageBody", "test%20message%20body%201")
	form.Add("SendMessageBatchRequestEntry.2.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.2.MessageBody", "test%20message%20body%202")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SendMessageBatch)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body is what we expect.
	expected := "BatchEntryIdsNotDistinct"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestSendMessageBatch_POST_TooManyEntries(t *testing.T) {
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("SendMessageBatchRequestEntry.1.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.1.MessageBody", "test%20message%20body%201")
	form.Add("SendMessageBatchRequestEntry.2.Id", "test_msg_002")
	form.Add("SendMessageBatchRequestEntry.2.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.3.Id", "test_msg_003")
	form.Add("SendMessageBatchRequestEntry.3.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.4.Id", "test_msg_004")
	form.Add("SendMessageBatchRequestEntry.4.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.5.Id", "test_msg_005")
	form.Add("SendMessageBatchRequestEntry.5.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.6.Id", "test_msg_006")
	form.Add("SendMessageBatchRequestEntry.6.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.7.Id", "test_msg_007")
	form.Add("SendMessageBatchRequestEntry.7.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.8.Id", "test_msg_008")
	form.Add("SendMessageBatchRequestEntry.8.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.9.Id", "test_msg_009")
	form.Add("SendMessageBatchRequestEntry.9.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.10.Id", "test_msg_010")
	form.Add("SendMessageBatchRequestEntry.10.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.11.Id", "test_msg_011")
	form.Add("SendMessageBatchRequestEntry.11.MessageBody", "test%20message%20body%202")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SendMessageBatch)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Check the response body is what we expect.
	expected := "TooManyEntriesInBatchRequest"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestSendMessageBatch_POST_Success(t *testing.T) {
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	app.SyncQueues.Queues["testing"] = &app.Queue{Name: "testing"}

	form := url.Values{}
	form.Add("Action", "SendMessageBatch")
	form.Add("QueueUrl", "http://localhost:4100/queue/testing")
	form.Add("SendMessageBatchRequestEntry.1.Id", "test_msg_001")
	form.Add("SendMessageBatchRequestEntry.1.MessageBody", "test%20message%20body%201")
	form.Add("SendMessageBatchRequestEntry.2.Id", "test_msg_002")
	form.Add("SendMessageBatchRequestEntry.2.MessageBody", "test%20message%20body%202")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Name", "test_attribute_name_1")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Value.StringValue", "test_attribute_value_1")
	form.Add("SendMessageBatchRequestEntry.2.MessageAttribute.1.Value.DataType", "String")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SendMessageBatch)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := "<MD5OfMessageBody>1c538b76fce1a234bce865025c02b042</MD5OfMessageBody>"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
