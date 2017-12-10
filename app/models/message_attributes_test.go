package models

import "testing"

func TestHashAttributes(t *testing.T) {
	expectedMd5 := "6f1b6e9ea669d43e4f7a7fe88fc6919e"
	attrs := make(map[string]MessageAttributeValue)
	attrs["test"] = MessageAttributeValue{
		dataType: "String",
		value:    "test",
		valueKey: "test",
	}

	hashedAttrs := HashAttributes(attrs)

	if expectedMd5 != hashedAttrs {
		t.Errorf("Bad Hash of Attrs on HashAttributes test, wanted '%s' but got '%s'", expectedMd5, hashedAttrs)
	}
}

func TestExtractMessageAttributes(t *testing.T) {
	//expectedMd5 := "6f1b6e9ea669d43e4f7a7fe88fc6919e"
	//attrs := make(map[string]MessageAttributeValue)
	//attrs["test"] = MessageAttributeValue{
	//	dataType: "String",
	//	value:    "test",
	//	valueKey: "test",
	//}
	//
	//msg := CreateMessage("test", attrs)
	//
	//attrs2 := Ex
	//if msg.Attributes["test"].value != "test" {
	//	t.Errorf("Bad Value of Attrs on ExtractMessageAttributes, wanted '%s' but got '%s'", expectedMd5, msg.MD5OfMessageAttributes)
	//}
}
