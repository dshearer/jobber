FROM centos:7

RUN yum -y install epel-release
RUN yum -y install rpm-build
RUN yum -y install tar
RUN yum -y install systemd
RUN yum -y install make
RUN yum clean all

RUN yum -y install wget
RUN wget https://archive.fedoraproject.org/pub/epel/6/x86_64/Packages/g/golang-bin-1.9.2-1.el6.x86_64.rpm
RUN wget https://archive.fedoraproject.org/pub/epel/6/x86_64/Packages/g/golang-src-1.9.2-1.el6.noarch.rpm
RUN wget https://archive.fedoraproject.org/pub/epel/6/x86_64/Packages/g/golang-1.9.2-1.el6.x86_64.rpm
RUN yum -y install ./golang-*

ENV SRC_ARCHIVE     src.tgz

COPY ${SRC_ARCHIVE} /

RUN tar -xzmf "${SRC_ARCHIVE}"

CMD cd /jobber-* && \
    make -C packaging/centos_7 pkg-local DESTDIR=/rpms && \
    cd /rpms && tar -cf /rpms.tar ./*
