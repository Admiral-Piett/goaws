package gosqs

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"sort"

	"github.com/archa347/goaws/app"
	log "github.com/sirupsen/logrus"
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

func hashAttributes(attributes map[string]app.MessageAttributeValue) string {
	hasher := md5.New()

	keys := sortedKeys(attributes)
	for _, key := range keys {
		attributeValue := attributes[key]

		addStringToHash(hasher, key)
		addStringToHash(hasher, attributeValue.DataType)
		if attributeValue.ValueKey == "StringValue" {
			hasher.Write([]byte{1})
			addStringToHash(hasher, attributeValue.Value)
		} else if attributeValue.ValueKey == "BinaryValue" {
			hasher.Write([]byte{2})
			bytes, _ := base64.StdEncoding.DecodeString(attributeValue.Value)
			addBytesToHash(hasher, bytes)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func sortedKeys(attributes map[string]app.MessageAttributeValue) []string {
	var keys []string
	for key, _ := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func addStringToHash(hasher hash.Hash, str string) {
	bytes := []byte(str)
	addBytesToHash(hasher, bytes)
}

func addBytesToHash(hasher hash.Hash, arr []byte) {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(len(arr)))
	hasher.Write(bs)
	hasher.Write(arr)
}
