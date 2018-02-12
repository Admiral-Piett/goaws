package gosqs

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/p4tin/goaws/app"
)

// applyQueueAttributes applies the requested queue attributes to the given
// queue.
// TODO Currently it only supports the VisibilityTimeout attribute.
func applyQueueAttributes(q *app.Queue, u url.Values) {
	attr := extractQueueAttributes(u)
	visibilityTimeout, _ := strconv.Atoi(attr["VisibilityTimeout"])
	if visibilityTimeout != 0 {
		q.TimeoutSecs = visibilityTimeout
	}
}

func extractQueueAttributes(u url.Values) map[string]string {
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
