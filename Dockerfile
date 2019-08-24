FROM alpine:3.7

RUN apk -U add ca-certificates

ADD build/linux-amd64 /bin/

ENTRYPOINT [ "go-starter" ]
