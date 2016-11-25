package gosns

import (
	"encoding/json"
	"testing"
)

const (
	testArn              = "arn:aws:sns:local:000000000000:UnitTestTopic1"
	messageKey           = "Message"
	messageStructureJSON = "json"
)

// When simple message string is passed,
// it must be used for all subscribers (no matter the protocol)
func TestCreateMessageBody_NonJson(t *testing.T) {
	message := "message text"
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, testArn, protocol, "")
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]string
	err = json.Unmarshal(snsMessage, &unmarshalled)
	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessage, ok := unmarshalled[messageKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, message)
	}

	expected := "message text"
	if receivedMessage != expected {
		t.Errorf(`expected message "%s" but received "%s"`, expected, receivedMessage)
	}
}

// When no protocol specific message is passed,
// default message must be forwarded
func TestCreateMessageBody_OnlyDefaultValueInJson(t *testing.T) {
	message := `{"default": "default message text", "http": "HTTP message text"}`
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, testArn, protocol, messageStructureJSON)
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]string
	err = json.Unmarshal(snsMessage, &unmarshalled)
	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessage, ok := unmarshalled[messageKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, message)
	}

	expected := "default message text"
	if receivedMessage != expected {
		t.Errorf(`expected message "%s" but received "%s"`, expected, receivedMessage)
	}
}

// When only protocol specific message is passed,
// error must be returned
func TestCreateMessageBody_OnlySqsValueInJson(t *testing.T) {
	message := `{"sqs": "message text"}`
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, testArn, protocol, messageStructureJSON)
	if err == nil {
		t.Fatalf(`error expected but instead SNS message was returned: %s`, snsMessage)
	}
}

// when both default and protocol specific messages are passed,
// protocol specific message must be used
func TestCreateMessageBody_BothDefaultAndSqsValuesInJson(t *testing.T) {
	message := `{"default": "default message text", "sqs": "sqs message text"}`
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, testArn, protocol, messageStructureJSON)
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]string
	err = json.Unmarshal(snsMessage, &unmarshalled)
	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessage, ok := unmarshalled[messageKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, message)
	}

	expected := "sqs message text"
	if receivedMessage != expected {
		t.Errorf(`expected message "%s" but received "%s"`, expected, receivedMessage)
	}
}

// When simple message string is passed,
// it must be used as is (even if it contains JSON)
func TestCreateMessageBody_NonJsonContainingJson(t *testing.T) {
	message := `{"default": "default message text", "sqs": "sqs message text"}`
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, testArn, protocol, "")
	if err != nil {
		t.Fatalf(`error creating SNS message: %s`, err)
	}

	var unmarshalled map[string]string
	err = json.Unmarshal(snsMessage, &unmarshalled)
	if err != nil {
		t.Fatalf(`error unmarshalling SNS message "%s": %s`, snsMessage, err)
	}

	receivedMessage, ok := unmarshalled[messageKey]
	if !ok {
		t.Fatalf(`SNS message "%s" does not contain key "%s"`, snsMessage, message)
	}

	expected := `{"default": "default message text", "sqs": "sqs message text"}`
	if receivedMessage != expected {
		t.Errorf(`expected message "%s" but received "%s"`, expected, receivedMessage)
	}
}
