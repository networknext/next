#!/bin/bash

bucket=
artifact=

print_usage() {
    printf "Usage: bootstrap.sh -a artifact\n\n"
    printf "b [string]\tBucket name on GCP Storage\n"
    printf "a [string]\tArtifact name on GCP Storage\n"

    printf "Example:\n\n"
    printf "> bootstrap.sh -b gs://development_artifacts -a server_backend.dev.tar.gz\n"
}

while getopts 'b:a:h' flag; do
  case "${flag}" in
    b) bucket="${OPTARG}" ;;
    a) artifact="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

# Copy libsodium from GCP Storage
gsutil cp "$bucket/libsodium.so" '/usr/local/lib'

# Copy libnorm1 from GCP Storage
gsutil cp "$bucket/libnorm.so.1" '/usr/local/lib'

# Copy libpgm from GCP Storage
gsutil cp "$bucket/libpgm-5.2.so.0" '/usr/local/lib'

# Copy libzmq from GCP Storage
gsutil cp "$bucket/libzmq.so" '/usr/local/lib'

# Refresh the known libs on the system
ldconfig

# Copy the required files for the service from GCP Storage
gsutil cp "$bucket/$artifact" artifact.tar.gz

# Stop the service
systemctl stop app.service

# Uncompress the artifact files into /app
tar -xvf artifact.tar.gz

# Set the app service binary to executable
chmod +x app

# Copy the Systemd service definition to the right location
cp app.service /etc/systemd/system/app.service

# Bump up the max socket read and write buffer sizes
sudo sysctl -w net.core.rmem_max=1000000000
sudo sysctl -w net.core.wmem_max=1000000000

# Start the service
systemctl daemon-reload
systemctl start app.service
