FROM alpine

COPY goaws /
EXPOSE 4100

ENTRYPOINT ["./goaws"]
