# GoAws
[![Build Status](https://travis-ci.org/p4tin/goaws.svg?branch=master)](https://travis-ci.org/p4tin/goaws)

You are always welcome to [tweet the creator in chief](https://twitter.com/gocodecloud) or [buy him a coffee](https://www.paypal.me/p4tin)

Written in Go this is a clone of the AWS SQS/SNS systems.  This system is designed to emulate SQS and SNS in a local environment so developers can test their interfaces without having to connect to the AWS Cloud and possibly incurring the expense, or even worse actually write to production topics/queues by mistake.  If you see any problems or would like to see a new feature, please open an issue here in github.  As well, I will logon to Gitter so we can discuss your deployment issues or the weather.


## SNS/SQS Api status:

All SNS/SQS APIs have been implemented except:
 - The full capabilities for Get and Set QueueAttributes.  At the moment you can only Get ALL the attributes.

Here is a list of the APIs:
 - [x] ListQueues
 - [x] CreateQueue
 - [x] GetQueueAttributes (unsupported attributes are mocked)
 - [x] GetQueueUrl
 - [x] SendMessage
 - [x] SendMessageBatch
 - [x] ReceiveMessage
 - [x] DeleteMessage
 - [x] DeleteMessageBatch
 - [x] PurgeQueue
 - [x] Delete Queue
 - [x] ChangeMessageVisibility
 - [ ] ChangeMessageVisibilityBatch
 - [ ] ListDeadLetterSourceQueues
 - [ ] ListQueueTags
 - [ ] RemovePermission
 - [x] SetQueueAttributes (Only supported attributes are set - see Supported Queue Attributes)
 - [ ] TagQueue
 - [ ] UntagQueue

## Supported Queue Attributes

 - [x] VisibilityTimeout
 - [x] ReceiveMessageWaitTimeSeconds
 - [x] RedrivePolicy

## Current SNS APIs implemented:

 - [x] ListTopics
 - [x] CreateTopic
 - [x] Subscribe (raw)
 - [x] ListSubscriptions
 - [x] Publish
 - [x] DeleteTopic
 - [x] Subscribe
 - [x] Unsubscribe
 - [X] ListSubscriptionsByTopic
 - [x] GetSubscriptionAttributes
 - [x] SetSubscriptionAttributes (Only supported attributes are set - see Supported Subscription Attributes)

## Supported Subscription Attributes

  - [x] RawMessageDelivery
  - [x] FilterPolicy (Only supported simplest "exact match" filter policy)


## Yaml Configuration Implemented

 - [x] Read config file
 - [x] -config flag to read a specific configuration file (e.g.: -config=myconfig.yaml)
 - [x] a command line argument to determine the environment to use in the config file (e.e.: Dev)
 - [x] IN the config file you can create Queues, Topic and Subscription see the example config file in the conf directory

### Example: Passing Configuration to Docker
```shell
docker run \
 --rm \
 --name=goaws \
 -v /path/on/host/goaws-config.yaml:/path/in/container/goaws-config.yaml \
 -p=4100:4100 \
 admiralpiett/goaws:v0.4.4-arm64 -config /path/in/container/goaws-config.yaml
```

## Debug logging can be turned on via a command line flag (e.g.: -debug)

## Note:  The system does not authenticate or presently use https

# Installation

    git clone git@github.com:Admiral-Piett/goaws.git

## Build and Run (Standalone)

    Build
        cd to GoAws directory
        go build -o goaws app/cmd/goaws.go  (The goaws executable should be in the currect directory, move it somewhere in your $PATH)

    Run
        ./goaws  (by default goaws listens on port 4100 but you can change it in the goaws.yaml file to another port of your choice)


## Run (Docker Version)

    Version >= v0.3.2
        docker pull admiralpiett/goaws

    Version <= v0.3.1
        docker pull pafortin/goaws

    run
        docker run -d --name goaws -p 4100:4100 admiralpiett/goaws



## Testing your installation

### Postman Environment: [LINK](..%2Fpostman%2FGoAWS%20Local.postman_environment.json)
### Postman Collection: [LINK](https://api.postman.com/collections/4714469-2b32c9da-aad4-4e9e-baee-6c11be6798a3?access_key=PMAT-01HG1KVFDXGGKH62KT141MBC0Z)

You can test that your installation is working correctly in one of two ways:

 1.  Using the postman collection, use this [link to import it](https://api.postman.com/collections/4714469-2b32c9da-aad4-4e9e-baee-6c11be6798a3?access_key=PMAT-01HG1KVFDXGGKH62KT141MBC0Z).  As well the Environment variable for the collection should be set as follows:  URL = http://localhost:4100/.

 2. by using the AWS cli tools ([download link](http://docs.aws.amazon.com/cli/latest/userguide/installing.html)) here are some samples, you can refer to the [aws cli tools docs](http://docs.aws.amazon.com/cli/latest/reference/) for further information.

* aws --endpoint-url http://localhost:4100 sqs create-queue --queue-name test1
```json
{
    "QueueUrl": "http://localhost:4100/test1"
}
```
* aws --endpoint-url http://localhost:4100 sqs list-queues
```json
{
    "QueueUrls": [
        "http://localhost:4100/test1"
    ]
}
```
* aws --endpoint-url http://localhost:4100 sqs send-message --queue-url http://localhost:4100/test1 --message-body "this is a test of the GoAws Queue messaging"
```json
{
    "MD5OfMessageBody": "9d3f5eaac3b1b4dd509f39e71e25f954",
    "MD5OfMessageAttributes": "b095c6d16871105acb75d59332513337",
    "MessageId": "66a1b4f5-cecf-473e-92b6-810156d41bbe"
}
```
* aws --endpoint-url http://localhost:4100 sqs receive-message --queue-url http://localhost:4100/test1
```json
{
    "Messages": [
        {
            "Body": "this is a test of the GoAws Queue messaging",
            "MD5OfMessageAttributes": "b095c6d16871105acb75d59332513337",
            "ReceiptHandle": "66a1b4f5-cecf-473e-92b6-810156d41bbe#f1fc455c-698e-442e-9747-f415bee5b461",
            "MD5OfBody": "9d3f5eaac3b1b4dd509f39e71e25f954",
            "MessageId": "66a1b4f5-cecf-473e-92b6-810156d41bbe"
        }
    ]
}
```
* aws --endpoint-url http://localhost:4100 sqs delete-message --queue-url http://localhost:4100/test1 --receipt-handle 66a1b4f5-cecf-473e-92b6-810156d41bbe#f1fc455c-698e-442e-9747-f415bee5b461
```
No output
```
* aws --endpoint-url http://localhost:4100 sqs receive-message --queue-url http://localhost:4100/test1
```
No output (No messages in Q)
```
* aws --endpoint-url http://localhost:4100 sqs delete-queue --queue-url http://localhost:4100/test1
```
No output
```
* aws --endpoint-url http://localhost:4100 sqs list-queues
```
No output (There are no Queues left)
```

* aws --endpoint-url http://localhost:4100 sns list-topics  (Example Response from list-topics)
```json
{
    "Topics": [
        {
            "TopicArn": "arn:aws:sns:local:000000000000:topic1"
        },
        {
            "TopicArn": "arn:aws:sns:local:000000000000:topic2"
        }
    ]
}
```
