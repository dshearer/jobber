FROM alpine:3.6
RUN apk add --update make
RUN apk add --update alpine-sdk
RUN apk add --update 'go>=1.8.4'
RUN apk update

RUN adduser -D -G abuild builder
COPY jobber.tgz /home/builder/
RUN chown builder:abuild /home/builder/jobber.tgz
RUN mkdir -p /output
RUN chown builder:abuild /output

USER builder
WORKDIR /home/builder
RUN tar -xzmf jobber.tgz
RUN mv jobber.tgz jobber/packaging/alpine_3.6/

#USER root
#CMD /bin/sh -i

CMD make -C jobber/packaging/alpine_3.6 pkg-local "DESTDIR=${HOME}/" && \
    cd "${HOME}" && tar -czf /output/pkgs.tgz ./*.apk