FROM debian:8.5

RUN apt-get update
RUN apt-get install -y init-system-helpers

# This must come after all 'apt-get' lines.
# It allows services to be started.
RUN echo exit 0 > /usr/sbin/policy-rc.d

ENV PKG_DIR      /pkgs
ENV TEST_SCRIPTS /test_scripts.tar

COPY *.deb "${PKG_DIR}/"
COPY test_scripts.tar "${TEST_SCRIPTS}"

CMD dpkg -i "${PKG_DIR}"/*.deb && \
    (systemctl is-enabled jobber || exit 1) && \
    sleep 2 && \
    tar xf "${TEST_SCRIPTS}" && \
    test_scripts/run_tests