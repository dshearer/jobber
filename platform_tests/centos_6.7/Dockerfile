FROM centos:6.7

RUN yum -y install epel-release
RUN yum -y install tar

ENV RPM_DIR      /rpms
ENV TEST_SCRIPTS /test_scripts.tar

COPY *.rpm "${RPM_DIR}/"
COPY test_scripts.tar "${TEST_SCRIPTS}"

CMD yum -y install "${RPM_DIR}"/*.rpm && \
    tar xf "${TEST_SCRIPTS}" && \
    test_scripts/run_tests
