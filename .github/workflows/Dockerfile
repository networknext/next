FROM golang:1.13

RUN apt-get -y update && \
    apt-get -y install libsodium-dev libcurl4-gnutls-dev valgrind

ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]