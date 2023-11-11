# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY sdk/ /app/sdk/
COPY cmd/server/ /app/cmd/server/

RUN g++ -o libnext.so -Isdk/include sdk/source/*.cpp -shared  -fPIC -lsodium -lcurl -lpthread -lm -DNEXT_DEVELOPMENT=1
RUN g++ -o server -Isdk/include cmd/server/server.cpp libnext.so -lcurl -lpthread -lm -DNEXT_DEVELOPMENT=1
RUN mv /app/libnext.so /usr/local/lib && ldconfig

EXPOSE 30000/udp

CMD [ "/app/server" ]
