package models

import (
	"net/url"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/stretchr/testify/assert"
)

func TestNewCreateTopicRequest(t *testing.T) {
	defer func() {
		test.ResetApp()
	}()

	result := NewCreateTopicRequest()

	assert.Equal(t, false, result.Attributes.FifoTopic)
	assert.Equal(t, StringToInt(1), result.Attributes.SignatureVersion)
	assert.Equal(t, "Active", result.Attributes.TracingConfig)
	assert.Equal(t, false, result.Attributes.ContentBasedDeduplication)
}

func TestCreateTopicRequest_SetAttributesFromForm_success(t *testing.T) {
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "new-queue")
	form.Add("Version", "2012-11-05")
	form.Add("Attribute.1.Name", "DeliveryPolicy")
	form.Add("Attribute.1.Value", "{\"i-am\":\"the-policy\", \"name\":\"delivery-policy\"}")
	form.Add("Attribute.2.Name", "DisplayName")
	form.Add("Attribute.2.Value", "Foo")
	form.Add("Attribute.3.Name", "FifoTopic")
	form.Add("Attribute.3.Value", "true")
	form.Add("Attribute.4.Name", "Policy")
	form.Add("Attribute.4.Value", "{\"i-am\":\"the-policy\", \"name\":\"policy\"}")
	form.Add("Attribute.5.Name", "SignatureVersion")
	form.Add("Attribute.5.Value", "99")
	form.Add("Attribute.6.Name", "TracingConfig")
	form.Add("Attribute.6.Value", "PassThrough")
	form.Add("Attribute.7.Name", "KmsMasterKeyId")
	form.Add("Attribute.7.Value", "1234abcd-12ab-34cd-56ef-1234567890ab")
	form.Add("Attribute.8.Name", "ArchivePolicy")
	form.Add("Attribute.8.Value", "{\"i-am\":\"the-policy\", \"name\":\"archive-policy\"}")
	form.Add("Attribute.9.Name", "BeginningArchiveTime")
	form.Add("Attribute.9.Value", "2024-07-01T23:59:59+09:00")
	form.Add("Attribute.10.Name", "ContentBasedDeduplication")
	form.Add("Attribute.10.Value", "true")

	ctr := &CreateTopicRequest{}
	ctr.SetAttributesFromForm(form)

	assert.Equal(t, 2, len(ctr.Attributes.DeliveryPolicy))
	assert.Equal(t, "the-policy", ctr.Attributes.DeliveryPolicy["i-am"])
	assert.Equal(t, "delivery-policy", ctr.Attributes.DeliveryPolicy["name"])
	assert.Equal(t, "Foo", ctr.Attributes.DisplayName)
	assert.Equal(t, true, ctr.Attributes.FifoTopic)
	assert.Equal(t, 2, len(ctr.Attributes.Policy))
	assert.Equal(t, "the-policy", ctr.Attributes.Policy["i-am"])
	assert.Equal(t, "policy", ctr.Attributes.Policy["name"])
	assert.Equal(t, StringToInt(99), ctr.Attributes.SignatureVersion)
	assert.Equal(t, "PassThrough", ctr.Attributes.TracingConfig)
	assert.Equal(t, "1234abcd-12ab-34cd-56ef-1234567890ab", ctr.Attributes.KmsMasterKeyId)
	assert.Equal(t, 2, len(ctr.Attributes.ArchivePolicy))
	assert.Equal(t, "the-policy", ctr.Attributes.ArchivePolicy["i-am"])
	assert.Equal(t, "archive-policy", ctr.Attributes.ArchivePolicy["name"])
	assert.Equal(t, "2024-07-01T23:59:59+09:00", ctr.Attributes.BeginningArchiveTime)
	assert.Equal(t, true, ctr.Attributes.ContentBasedDeduplication)
}

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
