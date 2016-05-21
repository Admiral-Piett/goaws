FROM scratch
COPY ./goaws_linux_amd64 /
ENTRYPOINT ["/goaws_linux_amd64"]
