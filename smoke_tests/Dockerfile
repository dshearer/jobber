FROM centos:7.2.1511

RUN yum -y install epel-release
RUN yum -y install rpm-build
RUN yum -y install golang
RUN yum -y install tar
RUN yum -y install systemd
RUN yum -y install make
RUN yum -y install python-pip
RUN yum clean all
RUN pip install robotframework

ENV ROBOT_TAR   robot_tests.tar
ENV SRC_TAR     jobber.tgz

COPY ${ROBOT_TAR}   /
COPY ${SRC_TAR}     /

CMD tar xzf /jobber.tgz && \
    make -C /jobber-*/packaging/centos_7 pkg-local DESTDIR=/ && \
    yum install -y /*.rpm && \
    (/usr/local/libexec/jobberd >/dev/null 2>&1 &) && \
    tar xf /robot_tests.tar && \
    robot robot_tests/test.robot