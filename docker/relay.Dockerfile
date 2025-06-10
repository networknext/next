# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY relay/reference /app

RUN g++ -o relay *.cpp -lsodium -lcurl -lpthread -lm

EXPOSE 40000/udp

ENV RELAY_NUM_THREADS="1"

CMD [ "/app/relay" ]
