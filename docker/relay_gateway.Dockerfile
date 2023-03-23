# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/relay_gateway /app/cmd/relay_gateway
COPY /envs/docker.bin /app/database.bin

RUN go build -o relay_gateway /app/cmd/relay_gateway/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/relay_gateway" ]
