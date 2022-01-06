# syntax=docker/dockerfile:1

FROM semaphoreci/golang:1.13

# Install gcloud sdk, pubsub emulator, bigtable emulator, libsodium, and libzmq
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg  add - && \
	apt-get --allow-releaseinfo-change update && \
	apt-get -y install google-cloud-sdk google-cloud-sdk-pubsub-emulator google-cloud-sdk-bigtable-emulator libsodium-dev libzmq3-dev && \
	rm -rf /var/lib/apt/lists/*


# Install go modules
# Requires the go.mod and go.sum files to be copied in order to download them
# Docker image must be built from the root of the repo to have access to the go.mod and go.sum files
# Example: docker build -f ./.semaphore/Dockerfile -t nbopardi/networknext:backend .

# Set the current working directory to root
WORKDIR /

COPY go.mod .
COPY go.sum .

RUN go mod download
