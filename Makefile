DEFAULT_GOAL: all
VERSION=0.0.1
BUILD_TIME=`date +%FT%T%z`
BINARY=./build/distributions/native/goaws
DOCKER_BINARY=./build/distributions/docker/goaws

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

clean:
	rm -Rf build

format:
	@echo "Build started at $(BUILD_TIME)"
	go fmt ./app/...

test:
	go test -cover ./app/...

all: $(SOURCES) format test
	go build -v -o ${BINARY} ./app/cmd/goaws.go


docker:
	GOOS=linux GOARCH=amd64 go build -v -o ${DOCKER_BINARY} app/cmd/goaws.go
	docker build -t pafortin/goaws .