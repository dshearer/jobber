FROM alpine:3.3
RUN apk upgrade --update

ENV PKG_DIR      /pkgs
ENV TEST_SCRIPTS /test_scripts.tar

COPY *.apk "${PKG_DIR}/"
COPY test_scripts.tar "${TEST_SCRIPTS}"

# NOTE: OpenRC doesn't work in a container
CMD (apk add --allow-untrusted "${PKG_DIR}"/*.apk ||:) && \
    (/usr/libexec/jobbermaster > /dev/null&) && \
    sleep 2 && \
    #/bin/sh
    tar xf "${TEST_SCRIPTS}" && \
    test_scripts/run_tests