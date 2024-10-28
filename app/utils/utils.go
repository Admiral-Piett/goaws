package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/Admiral-Piett/goaws/app/models"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/schema"
)

var XmlDecoder *schema.Decoder
var REQUEST_TRANSFORMER = TransformRequest

func init() {
	XmlDecoder = schema.NewDecoder()
	XmlDecoder.IgnoreUnknownKeys(true)
}

// QUESTION - alternately we could have the router.actionHandler method call this, but then our router maps
// need to track the request type AND the function call.  I think there'd be a lot of interface switching
// back and forth.
func TransformRequest(resultingStruct interfaces.AbstractRequestBody, req *http.Request, emptyRequestValid bool) (success bool) {
	switch req.Header.Get("Content-Type") {
	case "application/x-amz-json-1.0":
		//Read body data to parse json
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(resultingStruct)
		if err != nil {
			if emptyRequestValid && err == io.EOF {
				return true
			}
			log.Debugf("TransformRequest Failure - %s", err.Error())
			return false
		}
	default:
		err := req.ParseForm()
		if err != nil {
			log.Debugf("TransformRequest Failure - %s", err.Error())
			return false
		}
		err = XmlDecoder.Decode(resultingStruct, req.PostForm)
		if err != nil {
			log.Debugf("TransformRequest Failure - %s", err.Error())
			return false
		}
		resultingStruct.SetAttributesFromForm(req.PostForm)
	}

	return true
}

func ExtractQueueAttributes(u url.Values) map[string]string {
	attr := map[string]string{}
	for i := 1; true; i++ {
		nameKey := fmt.Sprintf("Attribute.%d.Name", i)
		attrName := u.Get(nameKey)
		if attrName == "" {
			break
		}

		valueKey := fmt.Sprintf("Attribute.%d.Value", i)
		attrValue := u.Get(valueKey)
		if attrValue != "" {
			attr[attrName] = attrValue
		}
	}
	return attr
}

func CreateErrorResponseV1(errKey string, isSqs bool) (int, interfaces.AbstractResponseBody) {
	var err interfaces.AbstractErrorResponse
	if isSqs {
		err = models.SqsErrors[errKey]
	} else {
		err = models.SnsErrors[errKey]
	}

	respStruct := models.ErrorResponse{
		Result:    err.Response(),
		RequestId: "00000000-0000-0000-0000-000000000000", // TODO - fix
	}
	return err.StatusCode(), respStruct
}

// TODO:
// Refactor internal model for MessageAttribute between SendMessage and ReceiveMessage
// from app.SqsMessageAttributeValue(old) to models.MessageAttributeValue(new) and remove this temporary function.
func ConvertToOldMessageAttributeValueStructure(newValues map[string]models.MessageAttributeValue) map[string]models.SqsMessageAttributeValue {
	attributes := make(map[string]models.SqsMessageAttributeValue)

	for name, entry := range newValues {
		// StringListValue and BinaryListValue is currently not implemented
		// Please refer app/gosqs/message_attributes.go
		value := ""
		valueKey := ""
		if entry.StringValue != "" {
			value = entry.StringValue
			valueKey = "StringValue"
		} else if entry.BinaryValue != "" {
			value = entry.BinaryValue
			valueKey = "BinaryValue"
		}
		attributes[name] = models.SqsMessageAttributeValue{
			Name:     name,
			DataType: entry.DataType,
			Value:    value,
			ValueKey: valueKey,
		}
	}

	return attributes
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func HashAttributes(attributes map[string]models.SqsMessageAttributeValue) string {
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

func sortedKeys(attributes map[string]models.SqsMessageAttributeValue) []string {
	var keys []string
	for key := range attributes {
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

func HasFIFOQueueName(queueName string) bool {
	return strings.HasSuffix(queueName, ".fifo")
}
