FROM centos:6.7

RUN yum -y install epel-release
RUN yum -y install daemonize
RUN yum -y install golang
RUN yum -y install initscripts
RUN yum -y install tar
RUN yum clean all

ENV SRC_ARCHIVE     src.tar
ENV GOPATH          /go_workspace
ENV JOBBER_PKG      github.com/dshearer/jobber
ENV TEST_SCRIPT_DIR /test_script
ENV SRCDIR          ${GOPATH}/src/${JOBBER_PKG}

COPY ${SRC_ARCHIVE} /
COPY test_script ${TEST_SCRIPT_DIR}

RUN mkdir -p "${SRCDIR}" && \
    tar -C "${SRCDIR}" -xf "/${SRC_ARCHIVE}"

CMD ${TEST_SCRIPT_DIR}/test

