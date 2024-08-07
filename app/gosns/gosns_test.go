package gosns

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Admiral-Piett/goaws/app/conf"
	"github.com/Admiral-Piett/goaws/app/test"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/common"
)

// TODO - add a subscription and I think this should work
func TestListSubscriptionByTopicResponse_No_Owner(t *testing.T) {
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "Local")
	defer func() {
		test.ResetApp()
	}()

	// set accountID to test value so it can be populated in response
	app.CurrentEnvironment.AccountID = "100010001000"

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("TopicArn", "arn:aws:sns:local:000000000000:local-topic1")
	req.PostForm = form

	// Prepare existant topic
	topic := &app.Topic{
		Name: "UnitTestTopic1",
		Arn:  "arn:aws:sns:local:100010001000:UnitTestTopic1",
		Subscriptions: []*app.Subscription{
			{
				TopicArn:        "",
				Protocol:        "",
				SubscriptionArn: "",
				EndPoint:        "",
				Raw:             false,
				FilterPolicy:    &app.FilterPolicy{},
			},
		},
	}
	app.SyncTopics.Topics["UnitTestTopic1"] = topic

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ListSubscriptionsByTopic)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `<Owner>` + app.CurrentEnvironment.AccountID + `</Owner>`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned empty owner for subscription member: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestSetSubscriptionAttributesHandler_FilterPolicy_POST_Success(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		test.ResetApp()
	}()

	topicName := "testing"
	topicArn := "arn:aws:sns:" + app.CurrentEnvironment.Region + ":000000000000:" + topicName
	subArn, _ := common.NewUUID()
	subArn = topicArn + ":" + subArn
	app.SyncTopics.Topics[topicName] = &app.Topic{Name: topicName, Arn: topicArn, Subscriptions: []*app.Subscription{
		{
			SubscriptionArn: subArn,
		},
	}}

	form := url.Values{}
	form.Add("SubscriptionArn", subArn)
	form.Add("AttributeName", "FilterPolicy")
	form.Add("AttributeValue", "{\"foo\": [\"bar\"]}")
	req.PostForm = form

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SetSubscriptionAttributes)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := "</SetSubscriptionAttributesResponse>"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	actualFilterPolicy := app.SyncTopics.Topics[topicName].Subscriptions[0].FilterPolicy
	if (*actualFilterPolicy)["foo"][0] != "bar" {
		t.Errorf("filter policy has not need applied")
	}
}
