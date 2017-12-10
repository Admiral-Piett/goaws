package models

import (
	"strings"
	"testing"
)

func TestCreateQueue(t *testing.T) {
	q := CreateQueue("myhost", "queue1", 10)

	if q.Name != "queue1" {
		t.Errorf("Expect the queue name to be %s but got %s", "queue1", q.Name)
	}

	if q.TimeoutSecs != 10 {
		t.Errorf("Expect the queue timeout to be %d but got %d", 10, q.TimeoutSecs)
	}

	if !strings.Contains(q.URL, "http") {
		t.Errorf("Expect the queue URL to contain %s, but it did not (%s)", "http", q.URL)
	}

	if !strings.Contains(q.Arn, "myhost") {
		t.Errorf("Expect the queue URL to contain %s, but it did not (%s)", "myhost", q.Arn)
	}
}
