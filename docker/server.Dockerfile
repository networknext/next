# syntax=docker/dockerfile:1

FROM ubuntu:22.10

WORKDIR /app

RUN apt update -y && apt install libsodium-dev build-essential libcurl4-openssl-dev -y

COPY sdk5/ /app/sdk5/

COPY cmd/server/ /app/cmd/server/

RUN g++ -o libnext5.so -Isdk5/include sdk5/source/*.cpp -shared  -fPIC -lsodium -lcurl -lpthread -lm

RUN g++ -o server -Isdk5/include cmd/server/server.cpp libnext5.so -lcurl -lpthread -lm

RUN mv /app/libnext5.so /usr/local/lib && ldconfig

EXPOSE 40000/udp

CMD [ "/app/server" ]
