FROM centos:7

ENV GOLANG_URL https://dl.google.com/go/go1.12.4.linux-amd64.tar.gz
ENV GOLANG_HASH d7d1f1f88ddfe55840712dc1747f37a790cbcaa448f6c9cf51bbe10aa65442f5

RUN yum update -y \
 && yum install -y \
      make \
      rpm-build \
      rsync \
      gcc \
      wget \
 && wget "${GOLANG_URL}" -O golang.tar.gz \
 && echo "${GOLANG_HASH}  golang.tar.gz" | sha256sum -cw \
 && tar -C /usr/local -xzf golang.tar.gz \
 && ln -s /usr/local/go/bin/go /usr/local/bin/go \
 && rm -f golang.tar.gz \
 && yum remove -y wget \
 && yum clean all \
 && rm -rf /var/cache/yum