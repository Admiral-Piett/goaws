package models

import (
	"encoding/json"
	"strconv"
)

// StringToInt this is a custom type that will allow our request bodies to support either a string OR an int.
// It has its own UnmarshalJSON method to handle both types automatically and it can return an `int`
// from the `Int` method.
type StringToInt int

func (s *StringToInt) UnmarshalJSON(data []byte) error {
	var i int
	err := json.Unmarshal(data, &i)
	if err == nil {
		*s = StringToInt(i)
		return nil
	}

	var str string
	err = json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	tmp, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	*s = StringToInt(tmp)
	return nil
}

func (s *StringToInt) Int() int {
	return int(*s)
}
