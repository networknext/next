# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY sdk/ /app/sdk/
COPY cmd/raspberry_server/ /app/cmd/raspberry_server/

RUN g++ -o libnext.so -Isdk/include sdk/source/*.cpp -shared  -fPIC -lsodium -lcurl -lpthread -lm -DNEXT_DEVELOPMENT=1
RUN g++ -o raspberry_server -Isdk/include cmd/raspberry_server/raspberry_server.cpp libnext.so -lcurl -lpthread -lm -DNEXT_DEVELOPMENT=1
RUN mv /app/libnext.so /usr/local/lib && ldconfig

EXPOSE 40000/udp

ENV RASPBERRY_FAKE_LATENCY 1

CMD [ "/app/raspberry_server" ]
