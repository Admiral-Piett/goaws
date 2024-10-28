package gosqs

import (
	"fmt"
	"net/http"

	"github.com/Admiral-Piett/goaws/app/models"
	log "github.com/sirupsen/logrus"
)

func extractMessageAttributes(req *http.Request, prefix string) map[string]models.SqsMessageAttributeValue {
	attributes := make(map[string]models.SqsMessageAttributeValue)
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
				attributes[name] = models.SqsMessageAttributeValue{name, dataType, value, valueKey}
			}
		}

		if _, ok := attributes[name]; !ok {
			log.Warnf("StringValue or BinaryValue of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
		}
	}

	return attributes
}

func getMessageAttributeResult(a *models.SqsMessageAttributeValue) *models.ResultMessageAttribute {
	v := &models.ResultMessageAttributeValue{
		DataType: a.DataType,
	}

	switch a.DataType {
	case "Binary":
		v.BinaryValue = a.Value
	default:
		v.StringValue = a.Value
	}

	return &models.ResultMessageAttribute{
		Name:  a.Name,
		Value: v,
	}
}
