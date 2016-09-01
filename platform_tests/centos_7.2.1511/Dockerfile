FROM centos:7.2.1511

ENV container docker

RUN yum -y install epel-release
RUN yum -y install tar
RUN yum -y install systemd
RUN (cd /lib/systemd/system/sysinit.target.wants/; for i in *; do [ $i == systemd-tmpfiles-setup.service ] || rm -f $i; done); \
rm -f /lib/systemd/system/multi-user.target.wants/*;\
rm -f /etc/systemd/system/*.wants/*;\
rm -f /lib/systemd/system/local-fs.target.wants/*; \
rm -f /lib/systemd/system/sockets.target.wants/*udev*; \
rm -f /lib/systemd/system/sockets.target.wants/*initctl*; \
rm -f /lib/systemd/system/basic.target.wants/*;\
rm -f /lib/systemd/system/anaconda.target.wants/*;

ENV RPM_DIR      /rpms
ENV TEST_SCRIPTS /test_scripts.tar

COPY *.rpm "${RPM_DIR}/"
COPY test_scripts.tar "${TEST_SCRIPTS}"

VOLUME /sys/fs/cgroup /run /tmp

CMD yum -y install "${RPM_DIR}"/*.rpm && \
    systemctl is-enabled jobber.service && \
    (/usr/local/libexec/jobberd >/dev/null 2>&1 &) && \
    tar xf "${TEST_SCRIPTS}" && \
    test_scripts/run_tests
