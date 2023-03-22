# syntax=docker/dockerfile:1

FROM ubuntu:22.10

WORKDIR /app

RUN apt update -y && apt install golang-go libsodium-dev ca-certificates -y

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
COPY /modules/ /app/modules/
COPY /cmd/map_cruncher/ /app/cmd/map_cruncher/

RUN go mod download

RUN go build -o map_cruncher /app/cmd/map_cruncher/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/map_cruncher" ]
