# syntax=docker/dockerfile:1

FROM ubuntu:22.10

WORKDIR /app

RUN apt update -y && apt install golang-go libsodium-dev ca-certificates -y

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
COPY /modules/ /app/modules/
COPY /cmd/ /app/cmd/

RUN go mod download

RUN go build -o portal_cruncher /app/cmd/portal_cruncher/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/portal_cruncher" ]
