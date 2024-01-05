# build image
FROM golang:alpine as build

WORKDIR /go/src/github.com/Admiral-Piett/goaws

COPY ./app/ ./app/
COPY ./go.mod .
COPY ./go.sum .

RUN ls -la
RUN CGO_ENABLED=0 go test ./app/...
RUN go build -o goaws app/cmd/goaws.go

# release image
FROM alpine

WORKDIR /app

COPY --from=build /go/src/github.com/Admiral-Piett/goaws/goaws ./goaws

COPY --from=build /go/src/github.com/Admiral-Piett/goaws/app/conf/goaws.yaml ./conf/

EXPOSE 4100
ENTRYPOINT ["./app/goaws"]
