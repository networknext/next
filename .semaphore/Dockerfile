# syntax=docker/dockerfile:1

FROM google/cloud-sdk:alpine

# Install go via copy
COPY --from=golang:1.13-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

# Install necessary packages
RUN apk add --no-cache coreutils build-base pkgconfig libsodium-dev zeromq-dev musl-dev openjdk8-jre && \
	gcloud components install --quiet beta pubsub-emulator bigtable

# Install go modules
# Requires the go.mod and go.sum files to be copied in order to download them
# Docker image must be built from the root of the repo to have access to the go.mod and go.sum files
# Example: docker build -f ./.semaphore/Dockerfile -t nbopardi/networknext:backend .

# Set the current working directory to root
WORKDIR /

COPY go.mod .
COPY go.sum .

RUN go mod download
