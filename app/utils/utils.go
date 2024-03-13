package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Admiral-Piett/goaws/app/interfaces"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/schema"
)

var XmlDecoder *schema.Decoder
var REQUEST_TRANSFORMER = TransformRequest

func InitializeDecoders() {
	XmlDecoder = schema.NewDecoder()
	XmlDecoder.IgnoreUnknownKeys(true)
}

// QUESTION - alternately we could have the router.actionHandler method call this, but then our router maps
// need to track the request type AND the function call.  I think there'd be a lot of interface switching
// back and forth.
func TransformRequest(resultingStruct interfaces.AbstractRequestBody, req *http.Request) (success bool) {
	switch req.Header.Get("Content-Type") {
	case "application/x-amz-json-1.0":
		//Read body data to parse json
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(resultingStruct)
		if err != nil {
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
