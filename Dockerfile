FROM alpine

EXPOSE 4100

RUN mkdir /conf

COPY ./goaws_linux_amd64 /
COPY ./goaws.yaml /conf/
ENTRYPOINT ["/goaws_linux_amd64"]
