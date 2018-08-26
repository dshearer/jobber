FROM golang:1.10-alpine3.7

RUN apk update && \
    apk upgrade && \
    apk add --no-cache make alpine-sdk rsync openssh-client && \
    adduser -D builder && \
    addgroup builder abuild

USER builder
