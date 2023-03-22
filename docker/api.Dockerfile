# syntax=docker/dockerfile:1

FROM ubuntu:22.10

WORKDIR /app

RUN apt update -y && apt install golang-go libsodium-dev ca-certificates -y

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
COPY /modules/ /app/modules/
COPY /cmd/ /app/cmd/

RUN go mod download

RUN go build -o api /app/cmd/api/*.go

COPY ./envs/local.bin /app/database.bin

EXPOSE 80

ENV ENV docker

CMD [ "/app/api" ]
