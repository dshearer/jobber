FROM golang:1.12-stretch

RUN apt-get update -y && \
    apt-get install -y --no-install-recommends dpkg-dev debhelper dh-systemd \
      rsync build-essential && \
    rm -rf /var/lib/apt/lists/*
