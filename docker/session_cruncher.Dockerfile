# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/session_cruncher/ /app/cmd/session_cruncher/
COPY /envs/docker.bin /app/database.bin

RUN go build -o session_cruncher /app/cmd/session_cruncher/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/session_cruncher" ]
