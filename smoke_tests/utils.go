package smoke_tests

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	urlLib "net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Admiral-Piett/goaws/app/router"
)

func generateServer() *httptest.Server {
	return httptest.NewServer(router.New())
}

// GenerateLocalProxyConfig use this to create AWS config that can be plugged into your sqs client, and
// force calls onto a local proxy.  This is helpful for testing directly with an HTTP inspection tool
// such as Charles or Proxyman.
// USAGE:
//
//	 //sdkConfig, _ := config.LoadDefaultConfig(context.TODO())
//		sdkConfig := GenerateLocalProxyConfig(9090)
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
