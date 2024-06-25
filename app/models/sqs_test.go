package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/stretchr/testify/assert"
)

func TestNewCreateQueueRequest(t *testing.T) {
	app.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize = 262144
	app.CurrentEnvironment.QueueAttributeDefaults.MessageRetentionPeriod = 345600
	app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds = 10
	app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout = 30
	defer func() {
		test.ResetApp()
	}()

	expectedCreateQueueRequest := &CreateQueueRequest{
		Attributes: QueueAttributes{
			DelaySeconds:                  0,
			MaximumMessageSize:            262144,
			MessageRetentionPeriod:        345600,
			ReceiveMessageWaitTimeSeconds: 10,
			VisibilityTimeout:             30,
		},
	}

	result := NewCreateQueueRequest()

	assert.Equal(t, expectedCreateQueueRequest, result)
}

func TestCreateQueueRequest_SetAttributesFromForm_success(t *testing.T) {
	expectedRedrivePolicy := RedrivePolicy{
		MaxReceiveCount:     100,
		DeadLetterTargetArn: "dead-letter-queue-arn",
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "new-queue")
	form.Add("Version", "2012-11-05")
	form.Add("Attribute.1.Name", "DelaySeconds")
	form.Add("Attribute.1.Value", "1")
	form.Add("Attribute.2.Name", "MaximumMessageSize")
	form.Add("Attribute.2.Value", "2")
	form.Add("Attribute.3.Name", "MessageRetentionPeriod")
	form.Add("Attribute.3.Value", "3")
	form.Add("Attribute.4.Name", "Policy")
	form.Add("Attribute.4.Value", "{\"i-am\":\"the-policy\"}")
	form.Add("Attribute.5.Name", "ReceiveMessageWaitTimeSeconds")
	form.Add("Attribute.5.Value", "4")
	form.Add("Attribute.6.Name", "VisibilityTimeout")
	form.Add("Attribute.6.Value", "5")
	form.Add("Attribute.7.Name", "RedrivePolicy")
	form.Add("Attribute.7.Value", "{\"maxReceiveCount\": 100, \"deadLetterTargetArn\":\"dead-letter-queue-arn\"}")
	form.Add("Attribute.8.Name", "RedriveAllowPolicy")
	form.Add("Attribute.8.Value", "{\"i-am\":\"the-redrive-allow-policy\"}")

	cqr := &CreateQueueRequest{
		Attributes: QueueAttributes{
			DelaySeconds:                  1,
			MaximumMessageSize:            262144,
			MessageRetentionPeriod:        345600,
			ReceiveMessageWaitTimeSeconds: 10,
			VisibilityTimeout:             30,
		},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, StringToInt(1), cqr.Attributes.DelaySeconds)
	assert.Equal(t, StringToInt(2), cqr.Attributes.MaximumMessageSize)
	assert.Equal(t, StringToInt(3), cqr.Attributes.MessageRetentionPeriod)
	assert.Equal(t, map[string]interface{}{"i-am": "the-policy"}, cqr.Attributes.Policy)
	assert.Equal(t, StringToInt(4), cqr.Attributes.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, StringToInt(5), cqr.Attributes.VisibilityTimeout)
	assert.Equal(t, expectedRedrivePolicy, cqr.Attributes.RedrivePolicy)
	assert.Equal(t, map[string]interface{}{"i-am": "the-redrive-allow-policy"}, cqr.Attributes.RedriveAllowPolicy)
}

func TestCreateQueueRequest_SetAttributesFromForm_success_handles_redrive_recieve_count_int(t *testing.T) {
	expectedRedrivePolicy := RedrivePolicy{
		MaxReceiveCount:     100,
		DeadLetterTargetArn: "dead-letter-queue-arn",
	}

	form := url.Values{}
	form.Add("Attribute.1.Name", "RedrivePolicy")
	form.Add("Attribute.1.Value", "{\"maxReceiveCount\": 100, \"deadLetterTargetArn\":\"dead-letter-queue-arn\"}")

	cqr := &CreateQueueRequest{
		Attributes: QueueAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, expectedRedrivePolicy, cqr.Attributes.RedrivePolicy)
}

func TestCreateQueueRequest_SetAttributesFromForm_success_handles_redrive_recieve_count_string(t *testing.T) {
	expectedRedrivePolicy := RedrivePolicy{
		MaxReceiveCount:     100,
		DeadLetterTargetArn: "dead-letter-queue-arn",
	}

	form := url.Values{}
	form.Add("Attribute.1.Name", "RedrivePolicy")
	form.Add("Attribute.1.Value", "{\"maxReceiveCount\": \"100\", \"deadLetterTargetArn\":\"dead-letter-queue-arn\"}")

	cqr := &CreateQueueRequest{
		Attributes: QueueAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, expectedRedrivePolicy, cqr.Attributes.RedrivePolicy)
}

func TestCreateQueueRequest_SetAttributesFromForm_success_default_unparsable_redrive_recieve_count(t *testing.T) {
	defaultRedrivePolicy := RedrivePolicy{
		MaxReceiveCount:     10,
		DeadLetterTargetArn: "dead-letter-queue-arn",
	}

	form := url.Values{}
	form.Add("Attribute.1.Name", "RedrivePolicy")
	form.Add("Attribute.1.Value", "{\"maxReceiveCount\": null, \"deadLetterTargetArn\":\"dead-letter-queue-arn\"}")

	cqr := &CreateQueueRequest{
		Attributes: QueueAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, defaultRedrivePolicy, cqr.Attributes.RedrivePolicy)
}

func TestCreateQueueRequest_SetAttributesFromForm_success_skips_invalid_values(t *testing.T) {
	form := url.Values{}
	form.Add("Attribute.1.Name", "DelaySeconds")
	form.Add("Attribute.1.Value", "garbage")
	form.Add("Attribute.2.Name", "MaximumMessageSize")
	form.Add("Attribute.2.Value", "garbage")
	form.Add("Attribute.3.Name", "MessageRetentionPeriod")
	form.Add("Attribute.3.Value", "garbage")
	form.Add("Attribute.4.Name", "Policy")
	form.Add("Attribute.4.Value", "garbage")
	form.Add("Attribute.5.Name", "ReceiveMessageWaitTimeSeconds")
	form.Add("Attribute.5.Value", "garbage")
	form.Add("Attribute.6.Name", "VisibilityTimeout")
	form.Add("Attribute.6.Value", "garbage")
	form.Add("Attribute.7.Name", "RedrivePolicy")
	form.Add("Attribute.7.Value", "garbage")
	form.Add("Attribute.8.Name", "RedriveAllowPolicy")
	form.Add("Attribute.8.Value", "garbage")

	cqr := &CreateQueueRequest{
		Attributes: QueueAttributes{
			DelaySeconds:                  1,
			MaximumMessageSize:            262144,
			MessageRetentionPeriod:        345600,
			ReceiveMessageWaitTimeSeconds: 10,
			VisibilityTimeout:             30,
		},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, StringToInt(1), cqr.Attributes.DelaySeconds)
	assert.Equal(t, StringToInt(262144), cqr.Attributes.MaximumMessageSize)
	assert.Equal(t, StringToInt(345600), cqr.Attributes.MessageRetentionPeriod)
	assert.Equal(t, map[string]interface{}(nil), cqr.Attributes.Policy)
	assert.Equal(t, StringToInt(10), cqr.Attributes.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, StringToInt(30), cqr.Attributes.VisibilityTimeout)
	assert.Equal(t, RedrivePolicy{}, cqr.Attributes.RedrivePolicy)
	assert.Equal(t, map[string]interface{}(nil), cqr.Attributes.RedriveAllowPolicy)
}

func TestRedrivePolicy_UnmarshalJSON_handles_nested_json(t *testing.T) {
	request := struct {
		MaxReceiveCount     int    `json:"maxReceiveCount"`
		DeadLetterTargetArn string `json:"deadLetterTargetArn"`
	}{
		MaxReceiveCount:     100,
		DeadLetterTargetArn: "arn:redrive-queue",
	}
	b, _ := json.Marshal(request)
	var r = RedrivePolicy{}
	err := r.UnmarshalJSON(b)

	assert.Nil(t, err)
	assert.Equal(t, StringToInt(100), r.MaxReceiveCount)
	assert.Equal(t, fmt.Sprintf("%s:%s", "arn", "redrive-queue"), r.DeadLetterTargetArn)
}

func TestRedrivePolicy_UnmarshalJSON_handles_escaped_string(t *testing.T) {
	request := `{"maxReceiveCount":"100","deadLetterTargetArn":"arn:redrive-queue"}`
	b, _ := json.Marshal(request)
	var r = RedrivePolicy{}
	err := r.UnmarshalJSON(b)

	assert.Nil(t, err)
	assert.Equal(t, StringToInt(100), r.MaxReceiveCount)
	assert.Equal(t, fmt.Sprintf("%s:%s", "arn", "redrive-queue"), r.DeadLetterTargetArn)
}

func TestRedrivePolicy_UnmarshalJSON_invalid_json_request_returns_error(t *testing.T) {
	request := fmt.Sprintf(`{\"maxReceiveCount\":\"100\",\"deadLetterTargetArn\":\"arn:redrive-queue\"}`)
	var r = RedrivePolicy{}
	err := r.UnmarshalJSON([]byte(request))

	assert.Error(t, err)
	assert.Equal(t, StringToInt(0), r.MaxReceiveCount)
	assert.Equal(t, "", r.DeadLetterTargetArn)
}

func TestRedrivePolicy_UnmarshalJSON_invalid_type_returns_error(t *testing.T) {
	request := `{"maxReceiveCount":true,"deadLetterTargetArn":"arn:redrive-queue"}`
	b, _ := json.Marshal(request)
	var r = RedrivePolicy{}
	err := r.UnmarshalJSON(b)

	assert.Error(t, err)
	assert.Equal(t, StringToInt(0), r.MaxReceiveCount)
	assert.Equal(t, "", r.DeadLetterTargetArn)
}

func TestNewListQueuesRequest_SetAttributesFromForm(t *testing.T) {
	form := url.Values{}
	form.Add("MaxResults", "1")
	form.Add("NextToken", "next-token")
	form.Add("QueueNamePrefix", "queue-name-prefix")

	lqr := &ListQueueRequest{}
	lqr.SetAttributesFromForm(form)

	assert.Equal(t, 1, lqr.MaxResults)
	assert.Equal(t, "next-token", lqr.NextToken)
	assert.Equal(t, "queue-name-prefix", lqr.QueueNamePrefix)
}

func TestListQueuesRequest_SetAttributesFromForm_invalid_max_results(t *testing.T) {
	form := url.Values{}
	form.Add("MaxResults", "1.0")
	form.Add("NextToken", "next-token")
	form.Add("QueueNamePrefix", "queue-name-prefix")

	lqr := &ListQueueRequest{}
	lqr.SetAttributesFromForm(form)

	assert.Equal(t, 0, lqr.MaxResults)
	assert.Equal(t, "next-token", lqr.NextToken)
	assert.Equal(t, "queue-name-prefix", lqr.QueueNamePrefix)
}

func TestGetQueueAttributesRequest_SetAttributesFromForm(t *testing.T) {
	form := url.Values{}
	form.Add("QueueUrl", "queue-url")
	form.Add("AttributeName.1", "attribute-1")
	form.Add("AttributeName.2", "attribute-2")

	lqr := &GetQueueAttributesRequest{}
	lqr.SetAttributesFromForm(form)

	assert.Equal(t, "queue-url", lqr.QueueUrl)
	assert.Equal(t, 2, len(lqr.AttributeNames))
	assert.Contains(t, lqr.AttributeNames, "attribute-1")
	assert.Contains(t, lqr.AttributeNames, "attribute-2")
}

func TestGetQueueAttributesRequest_SetAttributesFromForm_skips_invalid_key_sequence(t *testing.T) {
	form := url.Values{}
	form.Add("QueueUrl", "queue-url")
	form.Add("AttributeName.1", "attribute-1")
	form.Add("AttributeName.3", "attribute-3")

	lqr := &GetQueueAttributesRequest{}
	lqr.SetAttributesFromForm(form)

	assert.Equal(t, "queue-url", lqr.QueueUrl)
	assert.Equal(t, 1, len(lqr.AttributeNames))
	assert.Contains(t, lqr.AttributeNames, "attribute-1")
}

func TestSendMessageRequest_SetAttributesFromForm_success(t *testing.T) {
	form := url.Values{}
	form.Add("MessageAttribute.1.Name", "Attr1")
	form.Add("MessageAttribute.1.Value.DataType", "String")
	form.Add("MessageAttribute.1.Value.StringValue", "Value1")
	form.Add("MessageAttribute.2.Name", "Attr2")
	form.Add("MessageAttribute.2.Value.DataType", "Binary")
	form.Add("MessageAttribute.2.Value.BinaryValue", "VmFsdWUy")
	form.Add("MessageAttribute.3.Name", "")
	form.Add("MessageAttribute.3.Value.DataType", "String")
	form.Add("MessageAttribute.3.Value.StringValue", "Value")
	form.Add("MessageAttribute.4.Name", "Attr4")
	form.Add("MessageAttribute.4.Value.DataType", "")
	form.Add("MessageAttribute.4.Value.StringValue", "Value4")

	r := &SendMessageRequest{
		MessageAttributes:       make(map[string]MessageAttributeValue),
		MessageSystemAttributes: make(map[string]MessageAttributeValue),
	}
	r.SetAttributesFromForm(form)

	assert.Equal(t, 2, len(r.MessageAttributes))

	assert.NotNil(t, r.MessageAttributes["Attr1"])
	attr1 := r.MessageAttributes["Attr1"]
	assert.Equal(t, "String", attr1.DataType)
	assert.Equal(t, "Value1", attr1.StringValue)
	assert.Equal(t, "", attr1.BinaryValue)

	assert.NotNil(t, r.MessageAttributes["Attr2"])
	attr2 := r.MessageAttributes["Attr2"]
	assert.Equal(t, "Binary", attr2.DataType)
	assert.Equal(t, "", attr2.StringValue)
	assert.Equal(t, "VmFsdWUy", attr2.BinaryValue)
}

func TestSetQueueAttributesRequest_SetAttributesFromForm_success(t *testing.T) {
	expectedRedrivePolicy := RedrivePolicy{
		MaxReceiveCount:     100,
		DeadLetterTargetArn: "dead-letter-queue-arn",
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "new-queue")
	form.Add("Version", "2012-11-05")
	form.Add("Attribute.1.Name", "DelaySeconds")
	form.Add("Attribute.1.Value", "1")
	form.Add("Attribute.2.Name", "MaximumMessageSize")
	form.Add("Attribute.2.Value", "2")
	form.Add("Attribute.3.Name", "MessageRetentionPeriod")
	form.Add("Attribute.3.Value", "3")
	form.Add("Attribute.4.Name", "Policy")
	form.Add("Attribute.4.Value", "{\"i-am\":\"the-policy\"}")
	form.Add("Attribute.5.Name", "ReceiveMessageWaitTimeSeconds")
	form.Add("Attribute.5.Value", "4")
	form.Add("Attribute.6.Name", "VisibilityTimeout")
	form.Add("Attribute.6.Value", "5")
	form.Add("Attribute.7.Name", "RedrivePolicy")
	form.Add("Attribute.7.Value", "{\"maxReceiveCount\": 100, \"deadLetterTargetArn\":\"dead-letter-queue-arn\"}")
	form.Add("Attribute.8.Name", "RedriveAllowPolicy")
	form.Add("Attribute.8.Value", "{\"i-am\":\"the-redrive-allow-policy\"}")

	cqr := &SetQueueAttributesRequest{
		Attributes: QueueAttributes{
			DelaySeconds:                  1,
			MaximumMessageSize:            262144,
			MessageRetentionPeriod:        345600,
			ReceiveMessageWaitTimeSeconds: 10,
			VisibilityTimeout:             30,
		},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, StringToInt(1), cqr.Attributes.DelaySeconds)
	assert.Equal(t, StringToInt(2), cqr.Attributes.MaximumMessageSize)
	assert.Equal(t, StringToInt(3), cqr.Attributes.MessageRetentionPeriod)
	assert.Equal(t, map[string]interface{}{"i-am": "the-policy"}, cqr.Attributes.Policy)
	assert.Equal(t, StringToInt(4), cqr.Attributes.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, StringToInt(5), cqr.Attributes.VisibilityTimeout)
	assert.Equal(t, expectedRedrivePolicy, cqr.Attributes.RedrivePolicy)
	assert.Equal(t, map[string]interface{}{"i-am": "the-redrive-allow-policy"}, cqr.Attributes.RedriveAllowPolicy)
}

func TestSetQueueAttributesRequest_SetAttributesFromForm_success_handles_redrive_recieve_count_int(t *testing.T) {
	expectedRedrivePolicy := RedrivePolicy{
		MaxReceiveCount:     100,
		DeadLetterTargetArn: "dead-letter-queue-arn",
	}

	form := url.Values{}
	form.Add("Attribute.1.Name", "RedrivePolicy")
	form.Add("Attribute.1.Value", "{\"maxReceiveCount\": 100, \"deadLetterTargetArn\":\"dead-letter-queue-arn\"}")

	cqr := &SetQueueAttributesRequest{
		Attributes: QueueAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, expectedRedrivePolicy, cqr.Attributes.RedrivePolicy)
}

func TestSetQueueAttributesRequest_SetAttributesFromForm_success_handles_redrive_recieve_count_string(t *testing.T) {
	expectedRedrivePolicy := RedrivePolicy{
		MaxReceiveCount:     100,
		DeadLetterTargetArn: "dead-letter-queue-arn",
	}

	form := url.Values{}
	form.Add("Attribute.1.Name", "RedrivePolicy")
	form.Add("Attribute.1.Value", "{\"maxReceiveCount\": \"100\", \"deadLetterTargetArn\":\"dead-letter-queue-arn\"}")

	cqr := &SetQueueAttributesRequest{
		Attributes: QueueAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, expectedRedrivePolicy, cqr.Attributes.RedrivePolicy)
}

func TestSetQueueAttributesRequest_SetAttributesFromForm_success_default_unparsable_redrive_recieve_count(t *testing.T) {
	defaultRedrivePolicy := RedrivePolicy{
		MaxReceiveCount:     10,
		DeadLetterTargetArn: "dead-letter-queue-arn",
	}

	form := url.Values{}
	form.Add("Attribute.1.Name", "RedrivePolicy")
	form.Add("Attribute.1.Value", "{\"maxReceiveCount\": null, \"deadLetterTargetArn\":\"dead-letter-queue-arn\"}")

	cqr := &SetQueueAttributesRequest{
		Attributes: QueueAttributes{},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, defaultRedrivePolicy, cqr.Attributes.RedrivePolicy)
}

func TestSetQueueAttributesRequest_SetAttributesFromForm_success_skips_invalid_values(t *testing.T) {
	form := url.Values{}
	form.Add("Attribute.1.Name", "DelaySeconds")
	form.Add("Attribute.1.Value", "garbage")
	form.Add("Attribute.2.Name", "MaximumMessageSize")
	form.Add("Attribute.2.Value", "garbage")
	form.Add("Attribute.3.Name", "MessageRetentionPeriod")
	form.Add("Attribute.3.Value", "garbage")
	form.Add("Attribute.4.Name", "Policy")
	form.Add("Attribute.4.Value", "garbage")
	form.Add("Attribute.5.Name", "ReceiveMessageWaitTimeSeconds")
	form.Add("Attribute.5.Value", "garbage")
	form.Add("Attribute.6.Name", "VisibilityTimeout")
	form.Add("Attribute.6.Value", "garbage")
	form.Add("Attribute.7.Name", "RedrivePolicy")
	form.Add("Attribute.7.Value", "garbage")
	form.Add("Attribute.8.Name", "RedriveAllowPolicy")
	form.Add("Attribute.8.Value", "garbage")

	cqr := &SetQueueAttributesRequest{
		Attributes: QueueAttributes{
			DelaySeconds:                  1,
			MaximumMessageSize:            262144,
			MessageRetentionPeriod:        345600,
			ReceiveMessageWaitTimeSeconds: 10,
			VisibilityTimeout:             30,
		},
	}
	cqr.SetAttributesFromForm(form)

	assert.Equal(t, StringToInt(1), cqr.Attributes.DelaySeconds)
	assert.Equal(t, StringToInt(262144), cqr.Attributes.MaximumMessageSize)
	assert.Equal(t, StringToInt(345600), cqr.Attributes.MessageRetentionPeriod)
	assert.Equal(t, map[string]interface{}(nil), cqr.Attributes.Policy)
	assert.Equal(t, StringToInt(10), cqr.Attributes.ReceiveMessageWaitTimeSeconds)
	assert.Equal(t, StringToInt(30), cqr.Attributes.VisibilityTimeout)
	assert.Equal(t, RedrivePolicy{}, cqr.Attributes.RedrivePolicy)
	assert.Equal(t, map[string]interface{}(nil), cqr.Attributes.RedriveAllowPolicy)
}
