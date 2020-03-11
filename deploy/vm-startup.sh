#!/bin/bash

# When this is copied to the VM you need to specify the tar file for the relay or server backend
artifact=gs://artifacts.network-next-v3-dev.appspot.com/ARTIFACT.tar.gz

# Set up an /app directory in the root of the VM
mkdir -p /app
cd /app

# Copy the required files for the service from GCP Storage
gsutil cp gs://artifacts.network-next-v3-dev.appspot.com/GeoLite2-City.mmdb .
gsutil cp ${artifact} artifact.tar.gz

# Uncompress the artifact files into /app
tar -xvf artifact.tar.gz

# Set the app service binary to executable
chmod +x app

# Copy the Systemd service definition to the right location
cp /app/app.service /etc/systemd/system/app.service

# Finally start the service
systemctl start app.service