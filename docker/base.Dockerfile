# syntax=docker/dockerfile:1

FROM ubuntu:22.04

WORKDIR /app

RUN apt update -y && apt install libsodium-dev ca-certificates build-essential libcurl4-openssl-dev wget pkg-config -y

RUN wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin"

RUN go version

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

RUN go mod download
