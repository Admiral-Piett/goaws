package gosqs

import (
	"net/url"
	"time"

	"github.com/Admiral-Piett/goaws/app/models"

	log "github.com/sirupsen/logrus"
)

func init() {
	models.SyncQueues.Queues = make(map[string]*models.Queue)
}

func PeriodicTasks(d time.Duration, quit chan bool) {
	ticker := time.NewTicker(d)
	for {
		select {
		case <-ticker.C:
			models.SyncQueues.Lock()
			for qName := range models.SyncQueues.Queues {
				queue := models.SyncQueues.Queues[qName]

				// Reset deduplication period
				for dedupId, startTime := range queue.Duplicates {
					if time.Now().After(startTime.Add(models.DeduplicationPeriod)) {
						log.Debugf("deduplication period for message with deduplicationId [%s] expired", dedupId)
						delete(queue.Duplicates, dedupId)
					}
				}

				log.Debugf("Queue [%s] length [%d]", queue.Name, len(queue.Messages))
				for i := 0; i < len(queue.Messages); i++ {
					msg := &queue.Messages[i]

					if msg.ReceiptHandle != "" {
						if msg.VisibilityTimeout.Before(time.Now()) {
							log.Debugf("Making message visible again %s", msg.ReceiptHandle)
							queue.UnlockGroup(msg.GroupID)
							msg.ReceiptHandle = ""
							msg.ReceiptTime = time.Now().UTC()
							msg.Retry++
							if queue.MaxReceiveCount > 0 &&
								queue.DeadLetterQueue != nil &&
								msg.Retry >= queue.MaxReceiveCount {
								queue.DeadLetterQueue.Messages = append(queue.DeadLetterQueue.Messages, *msg)
								queue.Messages = append(queue.Messages[:i], queue.Messages[i+1:]...)
								i--
							}
						}
					}
				}
			}
			models.SyncQueues.Unlock()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func numberOfHiddenMessagesInQueue(queue models.Queue) int {
	num := 0
	for _, m := range queue.Messages {
		if m.ReceiptHandle != "" || m.DelaySecs > 0 && time.Now().Before(m.SentTime.Add(time.Duration(m.DelaySecs)*time.Second)) {
			num++
		}
	}
	return num
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
