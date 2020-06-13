FROM golang:1.13

RUN apt-get -y update && \
    apt-get -y install libsodium-dev libcurl4-gnutls-dev

RUN mkdir /app 
ADD . /app/
WORKDIR /app

RUN make build-test-func

ENV test=test_direct_default

CMD go run ./cmd/tools/functional/tests/func_tests.go ${test}