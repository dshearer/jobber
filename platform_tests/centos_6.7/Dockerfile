FROM centos:6.7

RUN yum -y install epel-release

ENV RPM_DIR     /rpms
ENV TEST_SCRIPT /test_installation

COPY *.rpm "${RPM_DIR}/"
COPY test_installation "${TEST_SCRIPT}"

CMD yum -y install "${RPM_DIR}"/*.rpm && \
   "${TEST_SCRIPT}"
