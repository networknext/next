#!/bin/bash

# When this is copied to the VM you need to specify the tar file for the relay or server backend
artifact=gs://artifacts.network-next-v3-dev.appspot.com/ARTIFACT.tar.gz

# Set up an /app directory in the root of the VM
cd /app

# Copy the required files for the service from GCP Storage
gsutil cp ${artifact} artifact.tar.gz

# Stop the service
systemctl stop app.service

# Uncompress the artifact files into /app
tar -xvf artifact.tar.gz

# Set the app service binary to executable
chmod +x app

# Copy the Systemd service definition to the right location
cp /app/app.service /etc/systemd/system/app.service

# Start the service
systemctl daemon-reload
systemctl start app.service