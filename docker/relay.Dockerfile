# syntax=docker/dockerfile:1

FROM ubuntu:22.10

WORKDIR /app

RUN apt update -y && apt install libsodium-dev build-essential libcurl4-openssl-dev -y

COPY relay/ /app

RUN g++ -o relay *.cpp -lsodium -lcurl -lpthread -lm

EXPOSE 40000/udp

CMD [ "/app/relay" ]
