# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/api/ /app/cmd/api/
COPY /envs/docker.bin /app/database.bin

RUN go build -o api /app/cmd/api/*.go

EXPOSE 80

ENV ENV docker
ENV ALLOWED_ORIGIN *
ENV DEBUG_LOGS 1

CMD [ "/app/api" ]
