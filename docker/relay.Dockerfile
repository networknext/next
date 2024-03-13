# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY relay/reference /app

RUN g++ -o relay *.cpp -lsodium -lcurl -lpthread -lm

EXPOSE 40000/udp

CMD [ "/app/relay" ]
