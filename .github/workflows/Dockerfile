FROM golang:1.13

ADD entrypoint.sh /entrypoint.sh

RUN apt-get -y update && \
    apt-get -y install libsodium-dev libcurl4-gnutls-dev valgrind && \
    chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]