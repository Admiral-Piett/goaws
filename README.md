# GoAws  [![Join the chat at https://gitter.im/p4tin/GoAws](https://badges.gitter.im/p4tin/GoAws.svg)](https://gitter.im/p4tin/GoAws?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

 
Written in Go this is a clone of the AWS SQS/SNS systems.  This system is designed to emulate SQS ans SNS in a local environment so developers can test their interfaces without having to connect the the AWS Cloud and possibly incurring the expense, or even worse actually write to production topics/queues by mistake.  If you see any problems or would like to see a new feature, please open an issue here in github.  As well, I will logon to Gitter so we can discuss your deployment issues or the weather.


## Current SQS APIs implemented:

 * ListQueues
 * CreateQueue
 * GetQueueAttributes (Always returns all attributes - depth and arn are set correctly others are mocked)
 * GetQueueUrl
 * SendMessage
 * ReceiveMessage
 * DeleteMessage
 * PurgeQueue
 * Delete Queue
 
 Note: Not Yet Done (SetQueueAttributes)


 ## Next implement the following SNS APIs:
 
 * ListTopics
 * CreateTopic
 * DeleteTopic
 * CreateSubscription (+raw)
 * DeleteSubscription
 * DeteteTopic


## Installation

    go get github.com/p4tin/GoAws

## Build and Run (Stamdalone)

    Build
        cd to GoAws directory
        go build . 
        
    Run
        ./goaws  (by default goaws listens on port 4100 but you can change it with -port=XXXX)
        

## Run (Docker Version)

    Get it
        docker pull pafortin/goaws
        
    run
        docker run -d --name goaws -p 4100:4100 pafortin/goaws



## Testing your installation

You can test that your installation is working correctly in one of two ways:

 1.  Usign the postman collection provided in this repository ( refer to [postman docs](https://www.getpostman.com/docs/) on how to use it )

 2. by using the [AWS cli tools](http://docs.aws.amazon.com/cli/latest/userguide/installing.html) and using the following commands.

* aws --endpoint-url http://localhost:4100 sqs create-queue --queue-name test1  
```
    {
        "QueueUrl": "http://localhost:4100/test1"
    }
```
* aws --endpoint-url http://localhost:4100 sqs list-queues  
```
    {
        "QueueUrls": [
            "http://localhost:4100/test1"
        ]
    }
```
* aws --endpoint-url http://localhost:4100 sqs send-message --queue-url http://localhost:4100/test1 --message-body "this is a test of the GoAws Queue messaging"
```
    {
        "MD5OfMessageBody": "9d3f5eaac3b1b4dd509f39e71e25f954", 
        "MD5OfMessageAttributes": "b095c6d16871105acb75d59332513337", 
        "MessageId": "66a1b4f5-cecf-473e-92b6-810156d41bbe"
    }
```
* aws --endpoint-url http://localhost:4100 sqs receive-message --queue-url http://localhost:4100/test1  
```
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

