FROM debian:8.5

ENV PKG_DIR      /pkgs
ENV TEST_SCRIPTS /test_scripts.tar

COPY *.deb "${PKG_DIR}/"
COPY test_scripts.tar "${TEST_SCRIPTS}"

CMD dpkg -i "${PKG_DIR}"/*.deb && \
    service jobber start && \
    sleep 2 && \
    tar xf "${TEST_SCRIPTS}" && \
    test_scripts/run_tests