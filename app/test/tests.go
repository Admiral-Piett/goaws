package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	urlLib "net/url"
)

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
		// Default request body when none is provided
		if body == nil {
			req, err = http.NewRequest(method, url, http.NoBody)
		} else {
			b, _ := json.Marshal(body)
			request_body := bytes.NewBuffer(b)
			req, err = http.NewRequest(method, url, request_body)
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
