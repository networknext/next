FROM golang:1.14-buster

# Add GCP key to apt keyring
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -

# # Add package sources to lists
RUN echo "deb http://security.debian.org/debian-security stretch/updates main" | tee -a /etc/apt/sources.list && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

# # Install GCP tools
RUN apt-get update && \
    apt-get -y install \
        libsodium-dev \
        apt-transport-https \
        ca-certificates \
        software-properties-common \
        gnupg \
        openjdk-8-jdk \
        google-cloud-sdk-firestore-emulator \
        google-cloud-sdk-pubsub-emulator