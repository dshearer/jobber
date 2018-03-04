FROM centos:6

RUN yum -y install tar make sudo epel-release rpm-build \
    selinux-policy-devel checkpolicy rsync
RUN yum clean all
RUN yum -y install golang python-pip lsof
RUN yum clean all
RUN pip install robotframework

ENV TEST_TAR   platform_tests.tar
ENV SRC_TAR    jobber.tgz

COPY ${TEST_TAR}   /
COPY ${SRC_TAR}    /

CMD tar xzf /jobber.tgz && \
    make -C /jobber-*/packaging/centos_6 pkg-local DESTDIR=/ && \
    yum install -y /*.rpm && \
    useradd normuser --create-home && \
    tar xf /platform_tests.tar && \
    robot --pythonpath platform_tests/keywords platform_tests/suites
