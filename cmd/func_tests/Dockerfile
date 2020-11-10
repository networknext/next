FROM golang:1.13

RUN apt-get -y update && \
    apt-get -y install libsodium-dev libcurl4-gnutls-dev libzmq5 libzmq3-dev

RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN make build-test-func

ENV test=test_direct_default

CMD go run ./cmd/func_tests/func_tests.go ${test}