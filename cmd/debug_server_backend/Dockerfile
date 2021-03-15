FROM golang:1.13

RUN apt-get -y update && \
    apt-get -y install libsodium-dev libcurl4-gnutls-dev libzmq5 libzmq3-dev

RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN make build-server-backend

CMD ["./dist/server_backend"]