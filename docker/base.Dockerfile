# syntax=docker/dockerfile:1

FROM ubuntu:22.04

WORKDIR /app

RUN apt update -y && apt upgrade -y && apt install libsodium-dev ca-certificates build-essential curl libcurl4-openssl-dev wget pkg-config -y

RUN if lscpu | grep -q x86_64 ; then wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz ; else wget https://go.dev/dl/go1.21.4.linux-arm64.tar.gz ; fi

RUN tar -C /usr/local -xzf go1.21.4.linux-*.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin"

RUN go version

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

RUN go mod download
