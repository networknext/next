# syntax=docker/dockerfile:1

FROM node

WORKDIR /app

COPY portal/package*.json ./

RUN npm install

COPY portal/*.js ./
COPY portal/*.json ./
COPY portal/.env.* ./
COPY portal/src ./src
COPY portal/public ./public

EXPOSE 8080

CMD [ "yarn", "serve-local" ]
