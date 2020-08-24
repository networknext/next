#!/bin/bash

# This script is only used in the staging environment for clients

bucket='gs://staging_artifacts'

# Copy libsodium from GCP Storage
gsutil cp "$bucket/libsodium.so" '/usr/local/lib' || exit 1

# Create the dist dir under app so the linker can find libnext
mkdir -p '/app/dist'

# Copy libnext from GCP Storage
gsutil cp "$bucket/libnext3.so" '/app/dist' || exit 1

# Copy the list of servers from GCP Storage
gsutil cp "$bucket/staging_servers.txt" . || exit 1

# Refresh the known libs on the system
ldconfig

# Copy the required files for the service from GCP Storage
gsutil cp "$bucket/load_test_client.tar.gz" 'artifact.tar.gz' || exit 1

# Stop the service
systemctl stop app.service

# Uncompress the artifact files into /app
tar -xvf artifact.tar.gz

# Set the app service binary to executable
chmod +x app

# Copy the Systemd service definition to the right location
cp app.service /etc/systemd/system/app.service

# Bump up the max socket read and write buffer sizes
sysctl -w net.core.rmem_max=1000000000
sysctl -w net.core.wmem_max=1000000000

# Start the service
systemctl daemon-reload
systemctl start app.service
