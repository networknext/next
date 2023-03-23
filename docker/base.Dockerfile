# syntax=docker/dockerfile:1

FROM ubuntu:22.10

WORKDIR /app

RUN apt update -y && apt install golang-go libsodium-dev ca-certificates build-essential libcurl4-openssl-dev -y

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

RUN go mod download
