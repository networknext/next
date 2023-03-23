# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY sdk5/ /app/sdk5/
COPY cmd/client/ /app/cmd/client/

RUN g++ -o libnext5.so -Isdk5/include sdk5/source/*.cpp -shared  -fPIC -lsodium -lcurl -lpthread -lm
RUN g++ -o client -Isdk5/include cmd/client/client.cpp libnext5.so -lcurl -lpthread -lm
RUN mv /app/libnext5.so /usr/local/lib && ldconfig

CMD [ "/app/client" ]
