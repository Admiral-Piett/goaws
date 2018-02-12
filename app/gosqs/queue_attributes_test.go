package gosqs

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/p4tin/goaws/app"
)

func TestApplyQueueAttributes(t *testing.T) {
	q := &app.Queue{TimeoutSecs: 30}
	u := url.Values{}
	u.Add("Attribute.1.Name", "DelaySeconds")
	u.Add("Attribute.1.Value", "20")
	u.Add("Attribute.2.Name", "VisibilityTimeout")
	u.Add("Attribute.2.Value", "60")
	u.Add("Attribute.3.Name", "Policy")
	applyQueueAttributes(q, u)
	expected := &app.Queue{TimeoutSecs: 60}
	if ok := reflect.DeepEqual(q, expected); !ok {
		t.Fatalf("expected %+v, got %+v", expected, q)
	}
}

func TestExtractQueueAttributes(t *testing.T) {
	u := url.Values{}
	u.Add("Attribute.1.Name", "DelaySeconds")
	u.Add("Attribute.1.Value", "20")
	u.Add("Attribute.2.Name", "VisibilityTimeout")
	u.Add("Attribute.2.Value", "30")
	u.Add("Attribute.3.Name", "Policy")
	attr := extractQueueAttributes(u)
	expected := map[string]string{
		"DelaySeconds":      "20",
		"VisibilityTimeout": "30",
	}
	if ok := reflect.DeepEqual(attr, expected); !ok {
		t.Fatalf("expected %+v, got %+v", expected, attr)
	}
}
