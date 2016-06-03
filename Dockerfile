FROM alpine

EXPOSE 4100

COPY ./goaws_linux_amd64 /
COPY ./conf/goaws.yaml /conf/
ENTRYPOINT ["/goaws_linux_amd64"]
