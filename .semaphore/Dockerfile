# syntax=docker/dockerfile:1

FROM semaphoreci/node:14

WORKDIR /

COPY package.json .
COPY package-lock.json .

RUN npm i
