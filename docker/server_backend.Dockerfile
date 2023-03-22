# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/server_backend /app/cmd/server_backend

RUN go build -o server_backend /app/cmd/server_backend/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/server_backend" ]
