FROM centos:6.7

RUN yum -y install epel-release && \
    yum -y install daemonize \
                   git \
                   golang \
                   sudo && \
    yum clean all

ENV GOPATH        /go_workspace
ENV JOBBER_BRANCH docker
ENV SCRIPTDIR     /test_script

COPY test_script ${SCRIPTDIR}

RUN mkdir -p "${GOPATH}" "${SCRIPTDIR}"

CMD ${SCRIPTDIR}/test

