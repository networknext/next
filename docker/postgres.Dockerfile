FROM postgres:14-alpine

ENV POSTGRES_DB postgres

COPY schemas/sql/create.sql /docker-entrypoint-initdb.d/a.sql
COPY schemas/sql/docker.sql /docker-entrypoint-initdb.d/b.sql
