package interfaces

import (
	"net/url"
)

type AbstractRequestBody interface {
	SetAttributesFromForm(values url.Values)
}
