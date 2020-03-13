FROM docker.io/library/golang:1.14-alpine as builder

WORKDIR /go/src/github.com/p4tin/goaws

RUN apk add --no-cache --no-progress \
    ca-certificates \
    make \
    git \
    openssh \
    gcc \
    libc-dev

COPY . .

RUN GOOS=linux  GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags "-w -s" app/cmd/goaws.go

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/p4tin/goaws/app/conf/goaws.yaml /conf/goaws.yaml
COPY --from=builder /go/src/github.com/p4tin/goaws/goaws /
COPY --from=builder /go/src/github.com/p4tin/goaws/Dockerfile /

EXPOSE 4100

USER nobody

ENTRYPOINT ["/goaws"]
