package gosqs

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/Admiral-Piett/goaws/app/conf"

	"github.com/Admiral-Piett/goaws/app/fixtures"

	"github.com/Admiral-Piett/goaws/app/models"
	"github.com/Admiral-Piett/goaws/app/utils"
	"github.com/stretchr/testify/assert"
)

// TODO - Admiral-Piett these are better but still screwy.  It's easy to have race conditions in here, so
// we have to name all the queues uniquely and leave them around so we're not resetting ourselves.
// Stupid.  Handle the global memory issues and this can be easily resolved.
func Test_PeriodicTasks_deletes_deduplication_period_upon_expiration(t *testing.T) {
	models.DeduplicationPeriod = 20 * time.Millisecond
	quit := make(chan bool)
	defer func() {
		models.ResetApp()
		quit <- true
		models.DeduplicationPeriod = 5 * time.Minute
	}()

	qName := "gosqs-dedupe-queue1"
	mainQueue := &models.Queue{
		Name: qName,
		URL:  fmt.Sprintf("%s/%s", fixtures.BASE_URL, qName),
		Arn:  fmt.Sprintf("%s:%s", fixtures.BASE_SQS_ARN, qName),
		Duplicates: map[string]time.Time{
			"12345": time.Now(),
		},
	}
	models.SyncQueues.Lock()
	models.SyncQueues.Queues[qName] = mainQueue
	models.SyncQueues.Unlock()

	go PeriodicTasks(10*time.Millisecond, quit)

	assertions := func() bool {
		models.SyncQueues.Lock()
		defer models.SyncQueues.Unlock()

		ok := 0 == len(mainQueue.Duplicates)
		if !ok {
			return false
		}
		return true
	}
	assert.Eventually(t, assertions, 10*time.Second, 10*time.Millisecond)
}

func Test_PeriodicTasks_VisibilityTimeout_expires(t *testing.T) {
	quit := make(chan bool)
	defer func() {
		models.ResetApp()
		quit <- true
	}()
	qName := "gosqs-visibility-queue1"
	mainQueue := &models.Queue{
		Name: qName,
		URL:  fmt.Sprintf("%s/%s", fixtures.BASE_URL, qName),
		Arn:  fmt.Sprintf("%s:%s", fixtures.BASE_SQS_ARN, qName),
	}
	mainQueue.Messages = append(mainQueue.Messages, models.SqsMessage{
		MessageBody:       "1",
		ReceiptHandle:     "12345",
		VisibilityTimeout: time.Now().Add(30 * time.Millisecond),
	})

	models.SyncQueues.Lock()
	models.SyncQueues.Queues[qName] = mainQueue
	models.SyncQueues.Unlock()

	go PeriodicTasks(10*time.Millisecond, quit)

	assertions := func() bool {
		models.SyncQueues.Lock()
		defer models.SyncQueues.Unlock()

		ok := !mainQueue.Messages[0].ReceiptTime.IsZero()
		if !ok {
			return false
		}
		ok = "1" == mainQueue.Messages[0].MessageBody
		if !ok {
			return false
		}
		ok = "" == mainQueue.Messages[0].ReceiptHandle
		if !ok {
			return false
		}
		ok = 1 == mainQueue.Messages[0].Retry
		if !ok {
			return false
		}
		return true
	}
	assert.Eventually(t, assertions, 10*time.Second, 10*time.Millisecond)
}

func Test_PeriodicTasks_moves_single_message_to_dead_letter_queue_upon_passing_receive_count(t *testing.T) {
	quit := make(chan bool)
	defer func() {
		models.ResetApp()
		quit <- true
	}()

	qName := "gosqs-main-queue1"
	dlqName := "gosqs-dead-letter-queue1"
	dlqQueue := &models.Queue{
		Arn:  fmt.Sprintf("%s/%s", fixtures.BASE_SQS_ARN, dlqName),
		Name: dlqName,
		URL:  fmt.Sprintf("%s/%s", fixtures.BASE_URL, dlqName),
	}
	mainQueue := &models.Queue{
		Arn:             fmt.Sprintf("%s/%s", fixtures.BASE_SQS_ARN, qName),
		DeadLetterQueue: dlqQueue,
		MaxReceiveCount: 1,
		Name:            qName,
		URL:             fmt.Sprintf("%s/%s", fixtures.BASE_URL, qName),
	}

	go PeriodicTasks(10*time.Millisecond, quit)

	models.SyncQueues.Lock()
	mainQueue.Messages = append(mainQueue.Messages, models.SqsMessage{
		MessageBody:       "1",
		Retry:             100,
		ReceiptHandle:     "12345",
		VisibilityTimeout: time.Now().Add(10 * time.Millisecond),
	})
	models.SyncQueues.Queues[qName] = mainQueue
	models.SyncQueues.Queues[dlqName] = dlqQueue
	models.SyncQueues.Unlock()

	assertions := func() bool {
		models.SyncQueues.Lock()
		defer models.SyncQueues.Unlock()

		ok := len(dlqQueue.Messages) == 1
		if !ok {
			return false
		}
		ok = "1" == dlqQueue.Messages[0].MessageBody
		if !ok {
			return false
		}
		return true
	}
	assert.Eventually(t, assertions, 10*time.Second, 10*time.Millisecond)
}

func Test_PeriodicTasks_moves_multiple_messages_to_dead_letter_queue_upon_passing_receive_count(t *testing.T) {
	quit := make(chan bool)
	conf.LoadYamlConfig("../conf/mock-data/mock-config.yaml", "BaseUnitTests")
	defer func() {
		models.ResetApp()
		quit <- true
	}()

	mainQueue := models.SyncQueues.Queues["unit-queue2"]
	dlqQueue := models.SyncQueues.Queues["dead-letter-queue1"]

	assert.Len(t, dlqQueue.Messages, 0)

	go PeriodicTasks(10*time.Millisecond, quit)

	models.SyncQueues.Lock()
	mainQueue.Messages = append(mainQueue.Messages, models.SqsMessage{
		MessageBody:   "1",
		Retry:         100,
		ReceiptHandle: "12345",
	})
	mainQueue.Messages = append(mainQueue.Messages, models.SqsMessage{
		MessageBody:   "2",
		Retry:         100,
		ReceiptHandle: "23456",
	})
	models.SyncQueues.Unlock()

	assertions := func() bool {
		models.SyncQueues.Lock()
		defer models.SyncQueues.Unlock()

		ok := len(dlqQueue.Messages) == 2
		if !ok {
			return false
		}
		ok = "1" == dlqQueue.Messages[0].MessageBody
		if !ok {
			return false
		}
		ok = "2" == dlqQueue.Messages[1].MessageBody
		if !ok {
			return false
		}
		return true
	}
	assert.Eventually(t, assertions, 10*time.Second, 10*time.Millisecond)
}

// TODO - I think all these below belong in handler tests, not in here.  Double check the relevant
// handlers for coverage and delete.
func TestSendingAndReceivingFromFIFOQueueReturnsSameMessageOnError(t *testing.T) {
	done := make(chan bool)
	go PeriodicTasks(1*time.Second, done)

	// create a queue
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "requeue-reset.fifo")
	form.Add("Attribute.1.Name", "VisibilityTimeout")
	form.Add("Attribute.1.Value", "2")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ := CreateQueueV1(req)
	assert.Equal(t, status, http.StatusOK)

	// send a message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("MessageBody", "1")
	form.Add("MessageGroupId", "GROUP-X")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}

	// send a message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("MessageBody", "2")
	form.Add("MessageGroupId", "GROUP-X")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}

	// receive message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, resp := ReceiveMessageV1(req)
	assert.Equal(t, status, http.StatusOK)

	result := resp.GetResult().(models.ReceiveMessageResult)
	receiptHandleFirst := result.Messages[0].ReceiptHandle
	if string(result.Messages[0].Body) != "1" {
		t.Fatalf("should have received body 1: %s", err)
	}

	// try to receive another message and we should get none
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = ReceiveMessageV1(req)
	assert.Equal(t, status, http.StatusOK)

	if len(models.SyncQueues.Queues["requeue-reset.fifo"].FIFOMessages) != 1 {
		t.Fatal("there should be only 1 group locked")
	}

	if models.SyncQueues.Queues["requeue-reset.fifo"].FIFOMessages["GROUP-X"] != 0 {
		t.Fatal("there should be GROUP-X locked")
	}

	// remove message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "DeleteMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
	form.Add("ReceiptHandle", receiptHandleFirst)
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = DeleteMessageV1(req)
	assert.Equal(t, status, http.StatusOK)

	if len(models.SyncQueues.Queues["requeue-reset.fifo"].Messages) != 1 {
		t.Fatal("there should be only 1 message in queue")
	}

	// receive message - loop until visibility timeouts
	for {
		req, err = http.NewRequest("POST", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		form = url.Values{}
		form.Add("Action", "ReceiveMessage")
		form.Add("QueueUrl", "http://localhost:4100/queue/requeue-reset.fifo")
		form.Add("Version", "2012-11-05")
		req.PostForm = form

		status, resp := ReceiveMessageV1(req)
		assert.Equal(t, status, http.StatusOK)

		result := resp.GetResult().(models.ReceiveMessageResult)
		if len(result.Messages) == 0 {
			continue
		}

		if string(result.Messages[0].Body) != "2" {
			t.Fatalf("should have received body 2: %s", err)
		}
		break
	}

	done <- true
}

func TestSendMessage_POST_DuplicatationNotAppliedToStandardQueue(t *testing.T) {
	done := make(chan bool)
	go PeriodicTasks(1*time.Second, done)

	// create a queue
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "stantdard-testing")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ := CreateQueueV1(req)

	assert.Equal(t, status, http.StatusOK)

	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/stantdard-testing")
	form.Add("MessageBody", "Test1")
	form.Add("MessageDeduplicationId", "123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)

	// Check the status code is what we expect.
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if len(models.SyncQueues.Queues["stantdard-testing"].Messages) == 0 {
		t.Fatal("there should be 1 message in queue")
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/stantdard-testing")
	form.Add("MessageBody", "Test2")
	form.Add("MessageDeduplicationId", "123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)

	// Check the status code is what we expect.
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if len(models.SyncQueues.Queues["stantdard-testing"].Messages) == 1 {
		t.Fatal("there should be 2 messages in queue")
	}
	done <- true
}

func TestSendMessage_POST_DuplicatationDisabledOnFifoQueue(t *testing.T) {
	done := make(chan bool)
	go PeriodicTasks(1*time.Second, done)

	// create a queue
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "no-dup-testing.fifo")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ := CreateQueueV1(req)

	assert.Equal(t, status, http.StatusOK)

	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/no-dup-testing.fifo")
	form.Add("MessageBody", "Test1")
	form.Add("MessageDeduplicationId", "123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)

	// Check the status code is what we expect.
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if len(models.SyncQueues.Queues["no-dup-testing.fifo"].Messages) == 0 {
		t.Fatal("there should be 1 message in queue")
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/no-dup-testing.fifo")
	form.Add("MessageBody", "Test2")
	form.Add("MessageDeduplicationId", "123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)

	// Check the status code is what we expect.
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if len(models.SyncQueues.Queues["no-dup-testing.fifo"].Messages) != 2 {
		t.Fatal("there should be 2 message in queue")
	}
	done <- true
}

func TestSendMessage_POST_DuplicatationEnabledOnFifoQueue(t *testing.T) {
	done := make(chan bool)
	go PeriodicTasks(1*time.Second, done)

	// create a queue
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "dup-testing.fifo")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ := CreateQueueV1(req)

	assert.Equal(t, status, http.StatusOK)

	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	models.SyncQueues.Queues["dup-testing.fifo"].EnableDuplicates = true

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/dup-testing.fifo")
	form.Add("MessageBody", "Test1")
	form.Add("MessageDeduplicationId", "123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)

	// Check the status code is what we expect.
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if len(models.SyncQueues.Queues["dup-testing.fifo"].Messages) == 0 {
		t.Fatal("there should be 1 message in queue")
	}

	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/dup-testing.fifo")
	form.Add("MessageBody", "Test2")
	form.Add("MessageDeduplicationId", "123")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)

	// Check the status code is what we expect.
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}
	if len(models.SyncQueues.Queues["dup-testing.fifo"].Messages) != 1 {
		t.Fatal("there should be 1 message in queue")
	}
	if body := models.SyncQueues.Queues["dup-testing.fifo"].Messages[0].MessageBody; string(body) == "Test2" {
		t.Fatal("duplicate message should not be added to queue")
	}
	done <- true
}

func TestSendMessage_POST_DelaySeconds(t *testing.T) {
	// create a queue
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	form := url.Values{}
	form.Add("Action", "CreateQueue")
	form.Add("QueueName", "sendmessage-delay")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ := CreateQueueV1(req)

	assert.Equal(t, status, http.StatusOK)

	// send a message
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	form = url.Values{}
	form.Add("Action", "SendMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/sendmessage-delay")
	form.Add("MessageBody", "1")
	form.Add("DelaySeconds", "2")
	form.Add("Version", "2012-11-05")
	req.PostForm = form

	status, _ = SendMessageV1(req)
	if status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got \n%v want %v",
			status, http.StatusOK)
	}

	// receive message before delay is up
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/sendmessage-delay")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	status, _ = ReceiveMessageV1(req)
	assert.Equal(t, status, http.StatusOK)

	// receive message with wait should return after delay
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	form = url.Values{}
	form.Add("Action", "ReceiveMessage")
	form.Add("QueueUrl", "http://localhost:4100/queue/sendmessage-delay")
	form.Add("WaitTimeSeconds", "10")
	form.Add("Version", "2012-11-05")
	req.PostForm = form
	start := time.Now()
	status, _ = ReceiveMessageV1(req)
	elapsed := time.Since(start)
	assert.Equal(t, status, http.StatusOK)
	if elapsed < 1*time.Second {
		t.Errorf("handler didn't wait at all")
	}
	if elapsed > 4*time.Second {
		t.Errorf("handler didn't need to wait all WaitTimeSeconds=10, only DelaySeconds=2")
	}
}

func TestCreateErrorResponseV1(t *testing.T) {
	expectedResponse := models.ErrorResponse{
		Result: models.ErrorResult{
			Type:    "Not Found",
			Code:    "AWS.SimpleQueueService.NonExistentQueue",
			Message: "The specified queue does not exist for this wsdl version.",
		},
		RequestId: "00000000-0000-0000-0000-000000000000",
	}
	status, response := utils.CreateErrorResponseV1("QueueNotFound", true)

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, expectedResponse, response)
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
// credits: https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
