package utils

import (
	"net/url"
	"testing"

	"github.com/Admiral-Piett/goaws/app/fixtures"
	"github.com/Admiral-Piett/goaws/app/mocks"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	InitializeDecoders()
	m.Run()
}

func TestTransformRequest_success_json(t *testing.T) {
	_, r := GenerateRequestInfo("POST", "url", fixtures.JSONRequestBody, true)

	mock := &mocks.MockRequestBody{}

	ok := TransformRequest(mock, r)

	assert.True(t, ok)
	assert.Equal(t, "mock-value", mock.RequestFieldStr)
	assert.False(t, mock.SetAttributesFromFormCalled)
}

func TestTransformRequest_success_xml(t *testing.T) {
	_, r := GenerateRequestInfo("POST", "url", nil, false)
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "UnitTestQueue1")
	form.Add("Attribute.1.Name", "VisibilityTimeout")
	form.Add("Attribute.1.Value", "60")
	form.Add("Attribute.2.Name", "MaximumMessageSize")
	form.Add("Attribute.2.Value", "2048")
	r.PostForm = form

	mock := &mocks.MockRequestBody{}

	ok := TransformRequest(mock, r)

	assert.True(t, ok)
	assert.True(t, mock.SetAttributesFromFormCalled)
	assert.Equal(t, []interface{}{form}, mock.SetAttributesFromFormCalledWith)
}

func TestTransformRequest_error_invalid_request_body_json(t *testing.T) {
	_, r := GenerateRequestInfo("POST", "url", "\"I-am-garbage", true)

	mock := &mocks.MockRequestBody{}

	ok := TransformRequest(mock, r)

	assert.False(t, ok)
	assert.Equal(t, "", mock.RequestFieldStr)
	assert.False(t, mock.SetAttributesFromFormCalled)
}

func TestTransformRequest_error_failure_to_parse_form_xml(t *testing.T) {
	_, r := GenerateRequestInfo("POST", "url", nil, false)

	mock := &mocks.MockRequestBody{}

	ok := TransformRequest(mock, r)

	assert.False(t, ok)
	assert.False(t, mock.SetAttributesFromFormCalled)
}

func TestTransformRequest_error_invalid_request_body_xml(t *testing.T) {
	_, r := GenerateRequestInfo("POST", "url", nil, false)

	form := url.Values{}
	form.Add("intField", "\"I-am-garbage")
	r.PostForm = form

	mock := &mocks.MockRequestBody{}

	ok := TransformRequest(mock, r)

	assert.False(t, ok)
	assert.False(t, mock.SetAttributesFromFormCalled)
}

func TestExtractQueueAttributes_success(t *testing.T) {
	u := url.Values{}
	u.Add("Attribute.1.Name", "DelaySeconds")
	u.Add("Attribute.1.Value", "20")
	u.Add("Attribute.2.Name", "VisibilityTimeout")
	u.Add("Attribute.2.Value", "30")
	u.Add("Attribute.3.Name", "Policy")

	attr := ExtractQueueAttributes(u)
	expected := map[string]string{
		"DelaySeconds":      "20",
		"VisibilityTimeout": "30",
	}

	assert.Equal(t, expected, attr)
}
