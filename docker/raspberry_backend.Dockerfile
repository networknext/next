# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/raspberry_backend/ /app/cmd/raspberry_backend/

RUN go build -o raspberry_backend /app/cmd/raspberry_backend/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/raspberry_backend" ]
