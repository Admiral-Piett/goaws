package models

import (
	"encoding/json"
	"testing"

	"github.com/Admiral-Piett/goaws/app/test"
	"github.com/stretchr/testify/assert"
)

type StringToIntStruct struct {
	Field1 StringToInt `json:"Field1"`
	Field2 StringToInt `json:"Field2"`
}

func TestStringToInt_unmarshalJSON_int(t *testing.T) {
	body := struct {
		Field1 int `json:"Field1"`
		Field2 int `json:"Field2"`
	}{
		Field1: 1,
		Field2: 2,
	}
	_, r := test.GenerateRequestInfo("POST", "/", body, true)

	result := &StringToIntStruct{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(result)

	assert.Nil(t, err)
	assert.Equal(t, StringToInt(1), result.Field1)
	assert.Equal(t, StringToInt(2), result.Field2)
}

func TestStringToInt_unmarshalJSON_string(t *testing.T) {
	body := struct {
		Field1 string `json:"Field1"`
		Field2 string `json:"Field2"`
	}{
		Field1: "1",
		Field2: "2",
	}
	_, r := test.GenerateRequestInfo("POST", "/", body, true)

	result := &StringToIntStruct{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(result)

	assert.Nil(t, err)
	assert.Equal(t, StringToInt(1), result.Field1)
	assert.Equal(t, StringToInt(2), result.Field2)
}

func TestStringToInt_unmarshalJSON_invalid_type_returns_error(t *testing.T) {
	body := struct {
		Field1 bool `json:"Field1"`
		Field2 bool `json:"Field2"`
	}{
		Field1: true,
		Field2: false,
	}
	_, r := test.GenerateRequestInfo("POST", "/", body, true)

	result := &StringToIntStruct{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(result)

	assert.Error(t, err)
}

func TestStringToInt_int_returns_int_type(t *testing.T) {
	s := StringToInt(1)

	assert.Equal(t, int(1), s.Int())
}
