# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/map_cruncher/ /app/cmd/map_cruncher/

RUN go build -o map_cruncher /app/cmd/map_cruncher/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/map_cruncher" ]
