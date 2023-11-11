package utils

import (
    "github.com/Admiral-Piett/goaws/app"
    "github.com/stretchr/testify/assert"
    "strings"
    "testing"
)

func ResetAppTopics() {
    app.SyncTopics.Lock()
    app.SyncTopics.Topics = map[string]*app.Topic{}
    app.SyncTopics.Unlock()
}

func ResetAppQueues() {
    app.SyncQueues.Lock()
    app.SyncQueues.Queues = map[string]*app.Queue{}
    app.SyncQueues.Unlock()
}

func AssertTopicsMatchFixture(t *testing.T, fixture map[string]*app.Topic) {
    for k, top := range app.SyncTopics.Topics {
        assert.Equal(t, fixture[k].Arn, top.Arn)
        assert.Equal(t, fixture[k].Name, top.Name)
        for i, sub := range top.Subscriptions {
            expectedSubscription := fixture[k].Subscriptions[i]
            assert.Equal(t, expectedSubscription.EndPoint, sub.EndPoint)
            assert.Equal(t, expectedSubscription.FilterPolicy, sub.FilterPolicy)
            assert.Equal(t, expectedSubscription.Protocol, sub.Protocol)
            assert.Equal(t, expectedSubscription.Raw, sub.Raw)
            assert.Equal(t, expectedSubscription.TopicArn, sub.TopicArn)
            assert.True(t, strings.HasPrefix(sub.SubscriptionArn, expectedSubscription.TopicArn))
        }
    }
}

func AssertQueuesMatchFixture(t *testing.T, fixture map[string]*app.Queue) {
    for k, que := range app.SyncQueues.Queues {
        assert.Equal(t, fixture[k].Arn, que.Arn)
        assert.Equal(t, fixture[k].Name, que.Name)
        assert.Equal(t, fixture[k].URL, que.URL)
        assert.Equal(t, fixture[k].TimeoutSecs, que.TimeoutSecs)
        assert.Equal(t, fixture[k].ReceiveWaitTimeSecs, que.ReceiveWaitTimeSecs)
        assert.Equal(t, fixture[k].DelaySecs, que.DelaySecs)
        assert.Equal(t, fixture[k].MaximumMessageSize, que.MaximumMessageSize)
        assert.Equal(t, fixture[k].Messages, que.Messages)
        assert.Equal(t, fixture[k].DeadLetterQueue, que.DeadLetterQueue)
        assert.Equal(t, fixture[k].MaxReceiveCount, que.MaxReceiveCount)
        assert.Equal(t, fixture[k].IsFIFO, que.IsFIFO)
        assert.Equal(t, fixture[k].FIFOMessages, que.FIFOMessages)
        assert.Equal(t, fixture[k].FIFOSequenceNumbers, que.FIFOSequenceNumbers)
        assert.Equal(t, fixture[k].EnableDuplicates, que.EnableDuplicates)
        assert.Equal(t, fixture[k].Duplicates, que.Duplicates)
    }
}