# syntax=docker/dockerfile:1

FROM ubuntu:22.10

WORKDIR /app

RUN apt update -y && apt install golang-go libsodium-dev ca-certificates -y

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
COPY /modules/ /app/modules/
COPY /cmd/relay_backend /app/cmd/relay_backend/

RUN go mod download

RUN go build -o relay_backend /app/cmd/relay_backend/*.go

COPY ./envs/docker.bin /app/database.bin

EXPOSE 80

ENV ENV docker

CMD [ "/app/relay_backend" ]
