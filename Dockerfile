FROM node:latest as build-stage
ARG ENVIRONMENT
ENV ENVIRONMENT $ENVIRONMENT
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY ./ .
RUN /bin/bash -c '([[ ${ENVIRONMENT} == "local" ]] && npm run build-local) || ([[ ${ENVIRONMENT} == "dev" ]] && npm run build-dev) || ([[ ${ENVIRONMENT} == "prod" ]] && npm run build-prod) || ([[ ${ENVIRONMENT} == "staging" ]] && npm run build-staging)'

FROM nginx as production-stage
RUN mkdir /app
COPY --from=build-stage /app/dist /app
COPY nginx.conf /etc/nginx/nginx.conf