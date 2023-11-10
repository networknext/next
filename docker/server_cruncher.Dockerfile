# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/server_cruncher/ /app/cmd/server_cruncher/
COPY /envs/docker.bin /app/database.bin

RUN go build -o server_cruncher /app/cmd/server_cruncher/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/server_cruncher" ]
