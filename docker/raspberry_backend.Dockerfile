# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/magic_backend/ /app/cmd/magic_backend/

RUN go build -o magic_backend /app/cmd/magic_backend/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/magic_backend" ]
