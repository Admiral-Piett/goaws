package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	urlLib "net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Admiral-Piett/goaws/app"
)

func ResetApp() {
	app.CurrentEnvironment = app.Environment{}
	ResetResources()
}

func ResetResources() {
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

// GenerateLocalProxyConfig use this to create AWS config that can be plugged into your sqs client, and
// force calls onto a local proxy.  This is helpful for testing directly with an HTTP inspection tool
// such as Charles or Proxyman.
func GenerateLocalProxyConfig(proxyPort int) aws.Config {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	proxyURL, _ := urlLib.Parse(fmt.Sprintf("http://127.0.0.1:%d", proxyPort))
	tr.Proxy = http.ProxyURL(proxyURL)
	client := &http.Client{Transport: tr}

	sdkConfig, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithHTTPClient(client),
	)
	return sdkConfig
}
