FROM alpine

COPY goaws /
COPY app/conf/goaws.yaml /app/conf/goaws.yaml
EXPOSE 4100

ENTRYPOINT ["./goaws"]
