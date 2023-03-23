# syntax=docker/dockerfile:1

FROM network_next_base

WORKDIR /app

COPY /modules/ /app/modules/
COPY /cmd/portal_cruncher /app/cmd/portal_cruncher/

RUN go build -o portal_cruncher /app/cmd/portal_cruncher/*.go

EXPOSE 80

ENV ENV docker

CMD [ "/app/portal_cruncher" ]
