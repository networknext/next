FROM postgres:14-alpine

ENV POSTGRES_DB postgres

COPY sql/create.sql /docker-entrypoint-initdb.d/a.sql
COPY sql/docker.sql /docker-entrypoint-initdb.d/b.sql