package gosqs

import (
	"fmt"
	"github.com/p4tin/goaws/app"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func extractMessageAttributes(req *http.Request, prefix string) map[string]app.MessageAttributeValue {
	attributes := make(map[string]app.MessageAttributeValue)
	if prefix != "" {
		prefix += "."
	}

	for i := 1; true; i++ {
		name := req.FormValue(fmt.Sprintf("%sMessageAttribute.%d.Name", prefix, i))
		if name == "" {
			break
		}

		dataType := req.FormValue(fmt.Sprintf("%sMessageAttribute.%d.Value.DataType", prefix, i))
		if dataType == "" {
			log.Warnf("DataType of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
			continue
		}

		// StringListValue and BinaryListValue is currently not implemented
		for _, valueKey := range [...]string{"StringValue", "BinaryValue"} {
			value := req.FormValue(fmt.Sprintf("%sMessageAttribute.%d.Value.%s", prefix, i, valueKey))
			if value != "" {
				attributes[name] = app.MessageAttributeValue{name, dataType, value, valueKey}
			}
		}

		if _, ok := attributes[name]; !ok {
			log.Warnf("StringValue or BinaryValue of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
		}
	}

	return attributes
}

func getMessageAttributeResult(a *app.MessageAttributeValue) *app.ResultMessageAttribute {
	v := &app.ResultMessageAttributeValue{
		DataType: a.DataType,
	}

	switch a.DataType {
	case "Binary":
		v.BinaryValue = a.Value
	case "String":
		v.StringValue = a.Value
	}

	return &app.ResultMessageAttribute{
		Name:  a.Name,
		Value: v,
	}
}
