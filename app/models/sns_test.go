package models

import (
	"net/url"
	"testing"

	"github.com/Admiral-Piett/goaws/app"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeRequest_SetAttributesFromForm_success(t *testing.T) {
	form := url.Values{}
	form.Add("Attributes.entry.1.key", "RawMessageDelivery")
	form.Add("Attributes.entry.1.value", "true")
	form.Add("Attributes.entry.2.key", "FilterPolicy")
	form.Add("Attributes.entry.2.value", "{\"filter\": [\"policy\"]}")

	cqr := &SubscribeRequest{
		Attributes: SubscriptionAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.True(t, cqr.Attributes.RawMessageDelivery)
	assert.Equal(t, app.FilterPolicy{"filter": []string{"policy"}}, cqr.Attributes.FilterPolicy)
}

func TestSubscribeRequest_SetAttributesFromForm_skips_invalid_values(t *testing.T) {
	form := url.Values{}
	form.Add("Attributes.entry.1.key", "RawMessageDelivery")
	form.Add("Attributes.entry.1.value", "garbage")
	form.Add("Attributes.entry.2.key", "FilterPolicy")
	form.Add("Attributes.entry.2.value", "also-garbage")

	cqr := &SubscribeRequest{
		Attributes: SubscriptionAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.False(t, cqr.Attributes.RawMessageDelivery)
	assert.Equal(t, app.FilterPolicy(nil), cqr.Attributes.FilterPolicy)
}

func TestSubscribeRequest_SetAttributesFromForm_stops_if_attributes_not_numbered_sequentially(t *testing.T) {
	form := url.Values{}
	form.Add("Attributes.entry.2.key", "RawMessageDelivery")
	form.Add("Attributes.entry.2.value", "garbage")
	form.Add("Attributes.entry.3.key", "FilterPolicy")
	form.Add("Attributes.entry.3.value", "also-garbage")

	cqr := &SubscribeRequest{
		Attributes: SubscriptionAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.False(t, cqr.Attributes.RawMessageDelivery)
	assert.Equal(t, app.FilterPolicy(nil), cqr.Attributes.FilterPolicy)
}
