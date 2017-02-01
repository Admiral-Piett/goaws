FROM alpine

EXPOSE 4100

COPY ./build/distributions/docker/goaws /
COPY ./app/conf/goaws.yaml /conf/
ENTRYPOINT ["/goaws"]
