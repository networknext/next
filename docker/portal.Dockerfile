FROM node:lts-alpine

WORKDIR /app
COPY portal/*.js ./
COPY portal/*.json ./
COPY portal/.env.* ./
COPY portal/src ./src
COPY portal/public ./public
RUN yarn install

EXPOSE 8080

CMD [ "yarn", "serve-local" ]
