FROM centos:6.7

RUN yum -y install epel-release
RUN yum -y install rpm-build
RUN yum -y install golang
RUN yum -y install tar
RUN yum -y install checkpolicy
RUN yum -y install policycoreutils
RUN yum -y install selinux-policy-devel
RUN yum clean all

ENV SRC_ARCHIVE     src.tgz

COPY ${SRC_ARCHIVE} /

RUN tar -xzf "${SRC_ARCHIVE}"

CMD cd /jobber-* && \
    make -C packaging/centos_6.7 rpm-local DESTDIR=/rpms && \
    cd /rpms && tar -cf /rpms.tar ./*
