# syntax=docker/dockerfile:1

FROM ubuntu:22.10

WORKDIR /app

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

COPY /modules/ /app/modules/

COPY /cmd/magic_backend/ /app/cmd/magic_backend/

RUN apt update -y && apt install golang-go libsodium-dev ca-certificates -y

RUN go mod download

RUN go build -o magic_backend /app/cmd/magic_backend/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/magic_backend" ]
