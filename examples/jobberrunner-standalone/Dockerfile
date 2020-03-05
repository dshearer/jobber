FROM debian:latest

# download jobberrunner
RUN set -xe \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates wget \
    && wget -O /usr/local/bin/jobberrunner https://github.com/AndiDittrich/jobber/releases/download/v1.4.0-p261/jobberrunner \
    && chmod 0755 /usr/local/bin/jobberrunner

# copy files
COPY --chown=www-data:www-data fs/ /

# user downgrade
USER www-data

# new entrypoint
ENTRYPOINT [ "/usr/local/bin/jobberrunner" , "-s", "/etc/cron.yaml" ]