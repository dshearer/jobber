FROM golang:1.12.4-alpine3.9

RUN apk update && \
    apk upgrade && \
    apk add --no-cache make alpine-sdk rsync openssh-client && \
    adduser -D builder && \
    addgroup builder abuild

USER builder
