package gosqs

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app/models"

	log "github.com/sirupsen/logrus"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/gorilla/mux"
)

func init() {
	app.SyncQueues.Queues = make(map[string]*app.Queue)
}

func PeriodicTasks(d time.Duration, quit <-chan struct{}) {
	ticker := time.NewTicker(d)
	for {
		select {
		case <-ticker.C:
			app.SyncQueues.Lock()
			for j := range app.SyncQueues.Queues {
				queue := app.SyncQueues.Queues[j]

				log.Debugf("Queue [%s] length [%d]", queue.Name, len(queue.Messages))
				for i := 0; i < len(queue.Messages); i++ {
					msg := &queue.Messages[i]

					// Reset deduplication period
					for dedupId, startTime := range queue.Duplicates {
						if time.Now().After(startTime.Add(app.DeduplicationPeriod)) {
							log.Debugf("deduplication period for message with deduplicationId [%s] expired", dedupId)
							delete(queue.Duplicates, dedupId)
						}
					}

					if msg.ReceiptHandle != "" {
						if msg.VisibilityTimeout.Before(time.Now()) {
							log.Debugf("Making message visible again %s", msg.ReceiptHandle)
							queue.UnlockGroup(msg.GroupID)
							msg.ReceiptHandle = ""
							msg.ReceiptTime = time.Now().UTC()
							msg.Retry++
							if queue.MaxReceiveCount > 0 &&
								queue.DeadLetterQueue != nil &&
								msg.Retry > queue.MaxReceiveCount {
								queue.DeadLetterQueue.Messages = append(queue.DeadLetterQueue.Messages, *msg)
								queue.Messages = append(queue.Messages[:i], queue.Messages[i+1:]...)
								i++
							}
						}
					}
				}
			}
			app.SyncQueues.Unlock()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func numberOfHiddenMessagesInQueue(queue app.Queue) int {
	num := 0
	for _, m := range queue.Messages {
		if m.ReceiptHandle != "" || m.DelaySecs > 0 && time.Now().Before(m.SentTime.Add(time.Duration(m.DelaySecs)*time.Second)) {
			num++
		}
	}
	return num
}

type DeleteEntry struct {
	Id            string
	ReceiptHandle string
	Error         string
	Deleted       bool
}

func DeleteMessageBatch(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	req.ParseForm()

	queueUrl := getQueueFromPath(req.FormValue("QueueUrl"), req.URL.String())
	queueName := ""
	if queueUrl == "" {
		vars := mux.Vars(req)
		queueName = vars["queueName"]
	} else {
		uriSegments := strings.Split(queueUrl, "/")
		queueName = uriSegments[len(uriSegments)-1]
	}

	deleteEntries := []DeleteEntry{}

	for k, v := range req.Form {
		keySegments := strings.Split(k, ".")
		if keySegments[0] == "DeleteMessageBatchRequestEntry" {
			keyIndex, err := strconv.Atoi(keySegments[1])
			if err != nil {
				createErrorResponse(w, req, "Error")
				return
			}

			if len(deleteEntries) < keyIndex {
				newDeleteEntries := make([]DeleteEntry, keyIndex)
				copy(newDeleteEntries, deleteEntries)
				deleteEntries = newDeleteEntries
			}

			if keySegments[2] == "Id" {
				deleteEntries[keyIndex-1].Id = v[0]
			}

			if keySegments[2] == "ReceiptHandle" {
				deleteEntries[keyIndex-1].ReceiptHandle = v[0]
			}
		}
	}

	deletedEntries := make([]app.DeleteMessageBatchResultEntry, 0)

	app.SyncQueues.Lock()
	if _, ok := app.SyncQueues.Queues[queueName]; ok {
		for _, deleteEntry := range deleteEntries {
			for i, msg := range app.SyncQueues.Queues[queueName].Messages {
				if msg.ReceiptHandle == deleteEntry.ReceiptHandle {
					// Unlock messages for the group
					log.Printf("FIFO Queue %s unlocking group %s:", queueName, msg.GroupID)
					app.SyncQueues.Queues[queueName].UnlockGroup(msg.GroupID)
					app.SyncQueues.Queues[queueName].Messages = append(app.SyncQueues.Queues[queueName].Messages[:i], app.SyncQueues.Queues[queueName].Messages[i+1:]...)
					delete(app.SyncQueues.Queues[queueName].Duplicates, msg.DeduplicationID)

					deleteEntry.Deleted = true
					deletedEntry := app.DeleteMessageBatchResultEntry{Id: deleteEntry.Id}
					deletedEntries = append(deletedEntries, deletedEntry)
					break
				}
			}
		}
	}
	app.SyncQueues.Unlock()

	notFoundEntries := make([]app.BatchResultErrorEntry, 0)
	for _, deleteEntry := range deleteEntries {
		if deleteEntry.Deleted {
			notFoundEntries = append(notFoundEntries, app.BatchResultErrorEntry{
				Code:        "1",
				Id:          deleteEntry.Id,
				Message:     "Message not found",
				SenderFault: true})
		}
	}

	respStruct := app.DeleteMessageBatchResponse{
		"http://queue.amazonaws.com/doc/2012-11-05/",
		app.DeleteMessageBatchResult{Entry: deletedEntries, Error: notFoundEntries},
		app.ResponseMetadata{RequestId: "00000000-0000-0000-0000-000000000001"}}

	enc := xml.NewEncoder(w)
	enc.Indent(" ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}

func getQueueFromPath(formVal string, theUrl string) string {
	if formVal != "" {
		return formVal
	}
	u, err := url.Parse(theUrl)
	if err != nil {
		return ""
	}
	return u.Path
}

func createErrorResponse(w http.ResponseWriter, req *http.Request, err string) {
	er := models.SqsErrors[err]
	respStruct := models.ErrorResponse{
		Result:    models.ErrorResult{Type: er.Type, Code: er.Code, Message: er.Message},
		RequestId: "00000000-0000-0000-0000-000000000000",
	}

	w.WriteHeader(er.HttpError)
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(respStruct); err != nil {
		log.Printf("error: %v\n", err)
	}
}
