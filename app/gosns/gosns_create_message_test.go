package gosns

import (
	"encoding/json"
	"testing"

	"github.com/Admiral-Piett/goaws/app"
)

const (
	messageKey            = "Message"
	subjectKey            = "Subject"
	messageStructureJSON  = "json"
	messageStructureEmpty = ""
	messageAttributesKey  = "MessageAttributes"
)

// When simple message string is passed,
// it must be used for all subscribers (no matter the protocol)
func TestCreateMessageBody_NonJson(t *testing.T) {
	message := "message text"
	subject := "subject"
	subs := &app.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}

	snsMessage, err := CreateMessageBody(subs, message, subject, messageStructureEmpty, make(map[string]app.MessageAttributeValue))
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(snsMessage, &unmarshalled)
	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessage, ok := unmarshalled[messageKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, messageKey)
	}

	if receivedMessage != message {
		t.Errorf(`expected message "%s" but received "%s"`, message, receivedMessage)
	}

	receivedSubject, ok := unmarshalled[subjectKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, subjectKey)
	}

	if receivedSubject != subject {
		t.Errorf(`expected subject "%s" but received "%s"`, subject, receivedSubject)
	}
}

// When no protocol specific message is passed,
// default message must be forwarded
func TestCreateMessageBody_OnlyDefaultValueInJson(t *testing.T) {
	subs := &app.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	message := `{"default": "default message text", "http": "HTTP message text"}`
	subject := "subject"

	snsMessage, err := CreateMessageBody(subs, message, subject, messageStructureJSON, nil)
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(snsMessage, &unmarshalled)
	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessage, ok := unmarshalled[messageKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, messageKey)
	}

	expected := "default message text"
	if receivedMessage != expected {
		t.Errorf(`expected message "%s" but received "%s"`, expected, receivedMessage)
	}

	receivedSubject, ok := unmarshalled[subjectKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, subjectKey)
	}

	if receivedSubject != subject {
		t.Errorf(`expected subject "%s" but received "%s"`, subject, receivedSubject)
	}
}

// When only protocol specific message is passed,
// error must be returned
func TestCreateMessageBody_OnlySqsValueInJson(t *testing.T) {
	subs := &app.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	message := `{"sqs": "message text"}`
	subject := "subject"

	snsMessage, err := CreateMessageBody(subs, message, subject, messageStructureJSON, nil)
	if err == nil {
		t.Fatalf(`error expected but instead SNS message was returned: %s`, snsMessage)
	}
}

// when both default and protocol specific messages are passed,
// protocol specific message must be used
func TestCreateMessageBody_BothDefaultAndSqsValuesInJson(t *testing.T) {
	subs := &app.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	message := `{"default": "default message text", "sqs": "sqs message text"}`
	subject := "subject"

	snsMessage, err := CreateMessageBody(subs, message, subject, messageStructureJSON, nil)
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(snsMessage, &unmarshalled)
	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessage, ok := unmarshalled[messageKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, messageKey)
	}

	expected := "sqs message text"
	if receivedMessage != expected {
		t.Errorf(`expected message "%s" but received "%s"`, expected, receivedMessage)
	}

	receivedSubject, ok := unmarshalled[subjectKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, subjectKey)
	}

	if receivedSubject != subject {
		t.Errorf(`expected subject "%s" but received "%s"`, subject, receivedSubject)
	}
}

// When simple message string is passed,
// it must be used as is (even if it contains JSON)
func TestCreateMessageBody_NonJsonContainingJson(t *testing.T) {
	subs := &app.Subscription{
		Protocol:        "sns",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	message := `{"default": "default message text", "sqs": "sqs message text"}`
	subject := "subject"

	snsMessage, err := CreateMessageBody(subs, message, subject, "", nil)
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(snsMessage, &unmarshalled)
	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessage, ok := unmarshalled[messageKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, messageKey)
	}

	expected := `{"default": "default message text", "sqs": "sqs message text"}`
	if receivedMessage != expected {
		t.Errorf(`expected message "%s" but received "%s"`, expected, receivedMessage)
	}

	receivedSubject, ok := unmarshalled[subjectKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, subjectKey)
	}

	if receivedSubject != subject {
		t.Errorf(`expected subject "%s" but received "%s"`, subject, receivedSubject)
	}
}

// When message attributes are passed,
// the requested Name, DataType, StringValue (or NumberValue) are correctly mapped to Name, Type, Value
func TestCreateMessageBody_WithMessageAttributes(t *testing.T) {
	message := "message text"
	subject := "subject"
	subs := &app.Subscription{
		Protocol:        "sqs",
		TopicArn:        "topic-arn",
		SubscriptionArn: "subs-arn",
		Raw:             false,
	}
	stringMessageAttributeValue := app.MessageAttributeValue{ValueKey: "StringValue", Value: "test", DataType: "String"}
	attributes := map[string]app.MessageAttributeValue{
		stringMessageAttributeValue.DataType: stringMessageAttributeValue,
	}
	snsMessage, err := CreateMessageBody(subs, message, subject, messageStructureEmpty, attributes)
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(snsMessage, &unmarshalled)

	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessageAttributes, ok := unmarshalled[messageAttributesKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, messageAttributesKey)
	}

	attributesMap, ok := receivedMessageAttributes.(map[string]interface{})
	if !ok {
		t.Fatalf(`SNS messageAttributes is invalid interface`)
	}

	attribute, ok := attributesMap[stringMessageAttributeValue.DataType]
	if !ok {
		t.Fatalf(`SNS messageAttributes does not contain key "%s"`, stringMessageAttributeValue.DataType)
	}

	attributeMap, ok := attribute.(map[string]interface{})
	if !ok {
		t.Fatalf(`SNS messageAttribute is invalid interface`)
	}

	attributeType, _ := attributeMap["Type"]
	if attributeType != stringMessageAttributeValue.DataType {
		t.Fatalf(`expected Type "%s" but received %s`, stringMessageAttributeValue.DataType, attributeType)
	}

	attributeValue, _ := attributeMap["Value"]
	if attributeValue != stringMessageAttributeValue.Value {
		t.Fatalf(`expected Value "%s" but received %s`, stringMessageAttributeValue.Value, attributeValue)
	}
}
