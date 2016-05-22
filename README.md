# GoAws

Written in Go this is a clone of the AWS SQS/SNS systems.  This system is designed to emulate SQS ans SNS in a local environment in order for developpers to test their interfaces without having to connect the the AWS Cloud and possibly incurring the expense, or even worse actually write to production topics/queues by mistake.  If you see any problems or would like to see a new feature, please open an issue here in github.  As well, I will logon to Gitter so we can discuss your deployment issues or the weather.

## Install

    go get github.com/p4tin/GoAws

## Build and Run

    Build
        cd to GoAws directory
        go build . 
        
    Run
        ./goaws  (by default goaws listens on port 4100 but you can change it with -port=XXXX)
        

## USING DOCKER

    Get it
        docker pull pafortin/goaws
        
    run
        docker run -d --name goaws -p 4100:4100 pafortin/goaws



## Testing your installation

You can test that your installation is working correctly by using the [AWS cli tools](http://docs.aws.amazon.com/cli/latest/userguide/installing.html) and using the following commands.

* aws --endpoint-url http://localhost:4100 sqs create-queue --queue-name test1  
```
    {
        "QueueUrl": "http://localhost:4100/test1"
    }
```
* aws --endpoint-url http://localhost:4100 sqs list-queues  
'''
    {
        "QueueUrls": [
            "http://localhost:4100/test1"
        ]
    }
'''


