package interfaces

import (
	"net/url"

	"github.com/Admiral-Piett/goaws/app/models"
)

type AbstractRequestBody interface {
	SetAttributesFromForm(values url.Values)
}

type AbstractResponseBody interface {
	GetResult() interface{}
	GetRequestId() string
}

type AbstractErrorResponse interface {
	Response() models.ErrorResult
	StatusCode() int
}

type AbstractPublishEntry interface {
	GetMessage() string
	GetMessageAttributes() map[string]models.MessageAttributeValue
	GetMessageStructure() string
	GetSubject() string
}
