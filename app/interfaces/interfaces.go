package interfaces

import (
	"net/url"
)

type AbstractRequestBody interface {
	SetAttributesFromForm(values url.Values)
}

type AbstractResponseBody interface {
	GetResult() interface{}
	GetRequestId() string
}
