package mocks

import "net/url"

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
