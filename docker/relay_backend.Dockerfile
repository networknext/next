# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/relay_backend /app/cmd/relay_backend/
COPY /envs/docker.bin /app/database.bin

RUN go build -o relay_backend /app/cmd/relay_backend/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/relay_backend" ]
