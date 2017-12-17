FROM centos:6

RUN yum -y install epel-release
RUN yum -y install rpm-build
RUN yum -y install golang
RUN yum -y install tar
RUN yum -y install make
RUN yum -y install python-pip
RUN yum -y install sudo
RUN yum -y install checkpolicy
RUN yum -y install selinux-policy-devel
RUN yum clean all
RUN pip install robotframework

ENV TEST_TAR   platform_tests.tar
ENV SRC_TAR     jobber.tgz

COPY ${TEST_TAR}   /
COPY ${SRC_TAR}     /

CMD tar xzf /jobber.tgz && \
    make -C /jobber-*/packaging/centos_6 pkg-local DESTDIR=/ && \
    yum install -y /*.rpm && \
    useradd normuser --create-home && \
    tar xf /platform_tests.tar && \
    robot platform_tests/test.robot