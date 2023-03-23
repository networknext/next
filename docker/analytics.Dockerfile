# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/analytics/ /app/cmd/analytics/

RUN go build -o analytics /app/cmd/analytics/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/analytics" ]
