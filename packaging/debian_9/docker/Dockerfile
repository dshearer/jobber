FROM debian:9.3

RUN apt-get update
RUN apt-get install -y make
RUN apt-get install -y dpkg-dev
RUN apt-get install -y debhelper
RUN apt-get install -y dh-systemd
RUN apt-get install -y apt-utils 
RUN apt-get install -y golang-1.8

COPY src.tgz /

RUN tar -xzmf /src.tgz

CMD export PATH=/usr/lib/go-1.8/bin:$PATH && \
    go version && \
    cd /jobber-* && \
    make -C packaging/debian_9 pkg-local DESTDIR=/pkgs && \
    cd /pkgs && tar -cf /pkgs.tar ./*
