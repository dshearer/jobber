FROM ubuntu:trusty

RUN apt-get update -y \
# next is for add-apt-repository
 && apt-get install -y --no-install-recommends software-properties-common \
 && add-apt-repository -y ppa:gophers/archive \
 && apt-get remove -y software-properties-common \
 && apt-get autoremove -y \
 && apt-get update -y \
 && apt-get install -y --no-install-recommends golang-1.9-go dpkg-dev debhelper dh-systemd rsync build-essential \
 && ln -s /usr/lib/go-1.9/bin/go /usr/bin/go \
 && rm -rf /var/lib/apt/lists/*
