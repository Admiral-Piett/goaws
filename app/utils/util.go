package utils

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/schema"
)

var xmlDecoder = schema.NewDecoder()

// QUESTION - alternately we could have the router.actionHandler method call this, but then our router maps
// need to track the request type AND the function call.  I think there'd be a lot of interface switching
// back and forth.
func TransformRequest(resultingStruct interface{}, req *http.Request) (success bool) {
	// TODO - put this somewhere else so we don't keep on rehashing this?
	xmlDecoder.IgnoreUnknownKeys(true)
	// TODO - do I still need the byJSON?
	// Should remove this flag after validateAndSetQueueAttributes was updated
	//byJson := false

	switch req.Header.Get("Content-Type") {
	case "application/x-amz-json-1.0":
		//Read body data to parse json
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(resultingStruct)
		if err != nil {
			log.Debugf("TransformRequest Failure - %s", err.Error())
			return false
		}
		//byJson = true
	default:
		// TODO - parse XML
		err := req.ParseForm()
		if err != nil {
			log.Debugf("TransformRequest Failure - %s", err.Error())
			return false
		}
		err = xmlDecoder.Decode(resultingStruct, req.PostForm)
		if err != nil {
			log.Debugf("TransformRequest Failure - %s", err.Error())
			return false
		}
	}

	return true
}
