FROM golang:1.13

RUN apt-get -y update && \
    apt-get -y install libsodium-dev libcurl4-gnutls-dev

RUN mkdir /app 
ADD . /app/
WORKDIR /app

RUN make build-relay-backend

CMD ["./dist/relay_backend"]