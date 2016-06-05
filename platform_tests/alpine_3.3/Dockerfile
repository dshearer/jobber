FROM alpine:3.3
RUN apk upgrade --update

ENV PKG_DIR     /pkgs
ENV TEST_SCRIPT /test_installation

COPY *.apk "${PKG_DIR}/"
COPY test_installation "${TEST_SCRIPT}"

# NOTE: OpenRC doesn't work in a container
CMD (apk add --allow-untrusted "${PKG_DIR}"/*.apk ||:) && \
    (/usr/libexec/jobberd&) && \
    "${TEST_SCRIPT}"