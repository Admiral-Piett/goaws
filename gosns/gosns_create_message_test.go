package gosns

import (
	"encoding/json"
	"testing"
)

const (
	testArn              = "arn:aws:sns:local:000000000000:UnitTestTopic1"
	messageKey           = "Message"
	subjectKey           = "Subject"
	messageStructureJSON = "json"
)

// When simple message string is passed,
// it must be used for all subscribers (no matter the protocol)
func TestCreateMessageBody_NonJson(t *testing.T) {
	message := "message text"
	subject := "subject"
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, subject, testArn, protocol, "")
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
	message := `{"default": "default message text", "http": "HTTP message text"}`
	subject := "subject"
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, subject, testArn, protocol, messageStructureJSON)
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
	message := `{"sqs": "message text"}`
	subject := "subject"
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, subject, testArn, protocol, messageStructureJSON)
	if err == nil {
		t.Fatalf(`error expected but instead SNS message was returned: %s`, snsMessage)
	}
}

// when both default and protocol specific messages are passed,
// protocol specific message must be used
func TestCreateMessageBody_BothDefaultAndSqsValuesInJson(t *testing.T) {
	message := `{"default": "default message text", "sqs": "sqs message text"}`
	subject := "subject"
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, subject, testArn, protocol, messageStructureJSON)
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
	message := `{"default": "default message text", "sqs": "sqs message text"}`
	subject := "subject"
	protocol := "sqs"

	snsMessage, err := CreateMessageBody(message, subject, testArn, protocol, "")
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
