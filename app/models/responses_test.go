package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: For now, we're only going to test those methods that do something other than just return a field

func TestGetQueueAttributesResponse_GetResult(t *testing.T) {
	gqa := GetQueueAttributesResponse{
		Result: GetQueueAttributesResult{Attrs: []Attribute{
			{Name: "attribute-name1", Value: "attribute-value1"},
			{Name: "attribute-name2", Value: "attribute-value2"},
		}},
	}

	expectedAttributes := map[string]map[string]string{
		"Attributes": {
			"attribute-name1": "attribute-value1",
			"attribute-name2": "attribute-value2",
		},
	}
	result := gqa.GetResult()

	assert.Equal(t, expectedAttributes, result)
}
