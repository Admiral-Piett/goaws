package models

import (
	"sync"
)

type Queue struct {
	Name        string
	URL         string
	Arn         string
	TimeoutSecs int
	Messages    []Message
}

var SyncQueues = struct {
	sync.RWMutex
	Queues map[string]*Queue
}{Queues: make(map[string]*Queue)}
