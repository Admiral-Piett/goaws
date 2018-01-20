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

	log "github.com/sirupsen/logrus"
)

type MessageAttributeValue struct {
	dataType string
	value    string
	valueKey string
}

func extractMessageAttributes(req *http.Request) map[string]MessageAttributeValue {
	attributes := make(map[string]MessageAttributeValue)

	for i := 1; true; i++ {
		name := req.FormValue(fmt.Sprintf("MessageAttribute.%d.Name", i))
		if name == "" {
			break
		}

		dataType := req.FormValue(fmt.Sprintf("MessageAttribute.%d.Value.DataType", i))
		if dataType == "" {
			log.Warnf("DataType of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
			continue
		}

		// StringListValue and BinaryListValue is currently not implemented
		for _, valueKey := range [...]string{"StringValue", "BinaryValue"} {
			value := req.FormValue(fmt.Sprintf("MessageAttribute.%d.Value.%s", i, valueKey))
			if value != "" {
				attributes[name] = MessageAttributeValue{dataType, value, valueKey}
			}
		}

		if _, ok := attributes[name]; !ok {
			log.Warnf("StringValue or BinaryValue of MessageAttribute %s is missing, MD5 checksum will most probably be wrong!\n", name)
		}
	}

	return attributes
}

func hashAttributes(attributes map[string]MessageAttributeValue) string {
	hasher := md5.New()

	keys := sortedKeys(attributes)
	for _, key := range keys {
		attributeValue := attributes[key]

		addStringToHash(hasher, key)
		addStringToHash(hasher, attributeValue.dataType)
		if attributeValue.valueKey == "StringValue" {
			hasher.Write([]byte{1})
			addStringToHash(hasher, attributeValue.value)
		} else if attributeValue.valueKey == "BinaryValue" {
			hasher.Write([]byte{2})
			bytes, _ := base64.StdEncoding.DecodeString(attributeValue.value)
			addBytesToHash(hasher, bytes)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func sortedKeys(attributes map[string]MessageAttributeValue) []string {
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
