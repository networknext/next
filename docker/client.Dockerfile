# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY sdk/ /app/sdk/
COPY cmd/client/ /app/cmd/client/

RUN g++ -o libnext.so -Isdk/include sdk/source/*.cpp -shared  -fPIC -lsodium -lcurl -lpthread -lm -DNEXT_DEVELOPMENT=1
RUN g++ -o client -Isdk/include cmd/client/client.cpp libnext.so -lcurl -lpthread -lm -DNEXT_DEVELOPMENT=1
RUN mv /app/libnext.so /usr/local/lib && ldconfig

CMD [ "/app/client" ]
