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
		RequestId: "00000000-0000-0000-0000-000000000000",
	}
	return err.StatusCode(), respStruct
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func HashAttributes(attributes map[string]models.MessageAttribute) string {
	hasher := md5.New()

	keys := sortedKeys(attributes)
	for _, key := range keys {
		attributeValue := attributes[key]

		addStringToHash(hasher, key)
		addStringToHash(hasher, attributeValue.DataType)
		if attributeValue.DataType == "String" {
			hasher.Write([]byte{1})
			addStringToHash(hasher, attributeValue.StringValue)
		} else if attributeValue.DataType == "Binary" {
			hasher.Write([]byte{2})
			bytes, _ := base64.StdEncoding.DecodeString(attributeValue.BinaryValue)
			addBytesToHash(hasher, []byte(bytes))
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func sortedKeys(attributes map[string]models.MessageAttribute) []string {
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
