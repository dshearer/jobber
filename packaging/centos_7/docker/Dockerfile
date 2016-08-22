FROM centos:7.2.1511

RUN yum -y install epel-release
RUN yum -y install rpm-build
RUN yum -y install golang
RUN yum -y install tar
RUN yum -y install systemd
RUN yum -y install make
RUN yum clean all

ENV SRC_ARCHIVE     src.tgz

COPY ${SRC_ARCHIVE} /

RUN tar -xzmf "${SRC_ARCHIVE}"

CMD cd /jobber-* && \
    make -C packaging/centos_7 pkg-local DESTDIR=/rpms && \
    cd /rpms && tar -cf /rpms.tar ./*
