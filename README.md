# GoAws

Written in Go this is a clone of AWS for Development purposes.

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




