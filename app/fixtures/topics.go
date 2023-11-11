package fixtures

import "github.com/Admiral-Piett/goaws/app"

var LOCAL_APP_SUBSCRIPTION_T1_Q4 = app.Subscription{
    EndPoint:  "arn:aws:sqs:::local-queue4",
    FilterPolicy:  nil,
    Protocol:  "sqs",
    Raw:  false,
    SubscriptionArn : "arn:aws:sns:::local-topic1:8cbd6c4f-349a-4176-a665-bc22049bfa6b",
    TopicArn:  "arn:aws:sns:::local-topic1",
}

var LOCAL_APP_SUBSCRIPTION_T1_Q5 = app.Subscription{
    EndPoint:  "arn:aws:sqs:::local-queue5",
    FilterPolicy:  &app.FilterPolicy{"foo": []string{"bar"}},
    Protocol:  "sqs",
    Raw:  true,
    SubscriptionArn : "arn:aws:sns:::local-topic1:8cbd6c4f-349a-4176-a665-bc22049bfa6b",
    TopicArn:  "arn:aws:sns:::local-topic1",
}

var LOCAL_APP_TOPIC_1 = app.Topic{
    Arn: "arn:aws:sns:::local-topic1",
    Name:  "local-topic1",
    Subscriptions : []*app.Subscription{
        &LOCAL_APP_SUBSCRIPTION_T1_Q4,
        &LOCAL_APP_SUBSCRIPTION_T1_Q5,
    },
}

var LOCAL_APP_TOPIC_2 = app.Topic{
    Arn: "arn:aws:sns:::local-topic2",
    Name:  "local-topic2",
    Subscriptions : []*app.Subscription{},
}

var LOCAL_APP_TOPICS = map[string]*app.Topic{
    "local-topic1": &LOCAL_APP_TOPIC_1,
    "local-topic2": &LOCAL_APP_TOPIC_2,
}
