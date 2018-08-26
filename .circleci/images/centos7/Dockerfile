FROM centos:7

RUN yum update -y \
 && yum install -y \
      golang \
      make \
      rpm-build \
      rsync \
 && yum clean all \
 && rm -rf /var/cache/yum
