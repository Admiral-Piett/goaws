# build image
FROM golang:alpine as build

WORKDIR /go/src/github.com/p4tin/goaws

COPY . .
RUN CGO_ENABLED=0 go test ./app/...
RUN go build -o goaws app/cmd/goaws.go

# release image
FROM alpine

COPY --from=build /go/src/github.com/p4tin/goaws/goaws /goaws

COPY app/conf/goaws.yaml /conf/

EXPOSE 4100
ENTRYPOINT ["/goaws"]
