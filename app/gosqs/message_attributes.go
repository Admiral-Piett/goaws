package gosqs

import (
	"github.com/Admiral-Piett/goaws/app"
)

func getMessageAttributeResult(a *app.MessageAttributeValue) *app.ResultMessageAttribute {
	v := &app.ResultMessageAttributeValue{
		DataType: a.DataType,
	}

	switch a.DataType {
	case "Binary":
		v.BinaryValue = a.Value
	case "String":
		v.StringValue = a.Value
	case "Number":
		v.StringValue = a.Value
	}

	return &app.ResultMessageAttribute{
		Name:  a.Name,
		Value: v,
	}
}
