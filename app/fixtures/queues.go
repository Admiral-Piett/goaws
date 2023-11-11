package fixtures

import (
    "github.com/Admiral-Piett/goaws/app"
    "time"
)

var LOCAL_APP_QUEUE4 = app.Queue{
    Name:                "local-queue4",
    URL:                 "http://://local-queue4",
    Arn:                 "arn:aws:sqs:::local-queue4",
    Duplicates: map[string]time.Time{},
}

var LOCAL_APP_QUEUE5 = app.Queue{
    Name:                "local-queue5",
    URL:                 "http://://local-queue5",
    Arn:                 "arn:aws:sqs:::local-queue5",
    Duplicates: map[string]time.Time{},
}

var LOCAL_APP_QUEUES = map[string]*app.Queue{
    "local-queue4": &LOCAL_APP_QUEUE4,
    "local-queue5": &LOCAL_APP_QUEUE5,
}
