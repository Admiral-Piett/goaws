FROM scratch
EXPOSE 4100
COPY ./goaws_linux_amd64 /
COPY ./goaws.yaml /
ENTRYPOINT ["/goaws_linux_amd64"]
