FROM centos:7.2.1511

RUN yum -y install epel-release
RUN yum -y install tar
RUN yum -y install systemd

ENV RPM_DIR      /rpms
ENV TEST_SCRIPTS /test_scripts.tar

COPY *.rpm "${RPM_DIR}/"
COPY test_scripts.tar "${TEST_SCRIPTS}"

CMD yum -y install "${RPM_DIR}"/*.rpm && \
    systemctl is-enabled jobber.service && \
    (/usr/local/libexec/jobbermaster >/dev/null 2>&1 &) && \
    tar xf "${TEST_SCRIPTS}" && \
    test_scripts/run_tests
