FROM alpine:3.7

RUN apk update && \
    apk upgrade && \
    apk add make && \
    apk add alpine-sdk && \
    apk add gcc && \
    apk add rsync && \
    apk add go && \
    apk add openssh-client

RUN adduser -D builder && \
    addgroup builder abuild

USER builder
