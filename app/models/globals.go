package models

import (
	"sync"
)

// CurrentEnvironment should get overwritten when the app starts up and loads the config.  For the
// sake of generating "partial" apps piece-meal during test automation we'll slap these placeholder
// values in here so the resource URLs aren't wonky like `http://://new-queue`.
var CurrentEnvironment = Environment{
	Host:      "host",
	Port:      "port",
	Region:    "region",
	AccountID: "accountID",
}

var LogMessages bool
var LogFile string

var SyncTopics = struct {
	sync.RWMutex
	Topics map[string]*Topic
}{Topics: make(map[string]*Topic)}

var SyncQueues = struct {
	sync.RWMutex
	Queues map[string]*Queue
}{Queues: make(map[string]*Queue)}
