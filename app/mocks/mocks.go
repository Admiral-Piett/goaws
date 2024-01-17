package mocks

import (
	"net/url"

	af "github.com/Admiral-Piett/goaws/app/fixtures"
)

type MockRequestBody struct {
	RequestFieldStr string `json:"field" schema:"field"`
	RequestFieldInt int    `json:"intField" schema:"intField"`

	SetAttributesFromFormCalled     bool
	SetAttributesFromFormCalledWith []interface{}

	MockSetAttributesFromFormCalledWith func(values url.Values)
}

func (m *MockRequestBody) SetAttributesFromForm(values url.Values) {
	m.SetAttributesFromFormCalled = true
	m.SetAttributesFromFormCalledWith = append(m.SetAttributesFromFormCalledWith, values)
	if m.MockSetAttributesFromFormCalledWith != nil {
		m.MockSetAttributesFromFormCalledWith(values)
	}
}

type BaseResponse struct {
	Message string `json:"Message" xml:"Message"`

	MockGetResult    func() interface{} `json:"-" xml:"-"`
	MockGetRequestId func() string      `json:"-" xml:"-"`
}

func (r BaseResponse) GetResult() interface{} {
	if r.MockGetResult != nil {
		return r.MockGetResult()
	}
	return r
}

func (r BaseResponse) GetRequestId() string {
	if r.MockGetRequestId != nil {
		return r.GetRequestId()
	}
	return af.REQUEST_ID
}
