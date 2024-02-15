package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	urlLib "net/url"

	"github.com/Admiral-Piett/goaws/app"
)

func ResetApp() {
	app.CurrentEnvironment = app.Environment{}

	app.SyncQueues.Lock()
	app.SyncQueues.Queues = make(map[string]*app.Queue)
	app.SyncQueues.Unlock()
	app.SyncTopics.Lock()
	app.SyncTopics.Topics = make(map[string]*app.Topic)
	app.SyncTopics.Unlock()
}

func GenerateRequestInfo(method, url string, body interface{}, isJson bool) (*httptest.ResponseRecorder, *http.Request) {
	if url == "" {
		url = "/health-check"
	}
	if method == "" {
		method = "GET"
	}

	var req *http.Request
	var err error
	if isJson {
		if body != nil {
			b, _ := json.Marshal(body)
			request_body := bytes.NewBuffer(b)
			req, err = http.NewRequest(method, url, request_body)
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	} else {
		req, err = http.NewRequest(method, url, nil)
		req.Header.Set("Content-Type", "multipart/form-data")
		body, _ := body.(urlLib.Values)
		req.PostForm = body
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	return rr, req
}
