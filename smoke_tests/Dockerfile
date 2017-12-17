FROM centos:7.2.1511

RUN yum -y install epel-release rpm-build golang tar systemd make sudo
RUN yum clean all
RUN yum -y install python-pip
RUN pip install robotframework

ENV TEST_TAR   platform_tests.tar
ENV SRC_TAR     jobber.tgz

COPY ${TEST_TAR}   /
COPY ${SRC_TAR}     /

CMD tar xzf /jobber.tgz && \
    make -C /jobber-*/packaging/centos_7 pkg-local DESTDIR=/ && \
    yum install -y /*.rpm && \
    (/usr/local/libexec/jobberd >/dev/null 2>&1 &) && \
    useradd normuser --create-home && \
    tar xf /platform_tests.tar && \
    robot platform_tests/test.robot