#!/bin/bash

artifact=

print_usage() {
    printf "Usage: vm-update-app.sh -a artifact\n\n"
    printf "a [string]\tPath to artifact on GCP Storage\n"

    printf "Example:\n\n"
    printf "> vm-update-app.sh -a gs://artifacts.network-next-v3-dev.appspot.com/server_backend.dev.tar.gz\n"
}

while getopts 'a:h' flag; do
  case "${flag}" in
    a) artifact="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

# Copy the required files for the service from GCP Storage
gsutil cp ${artifact} artifact.tar.gz

# Stop the service
systemctl stop app.service

# Uncompress the artifact files into /app
tar -xvf artifact.tar.gz

# Set the app service binary to executable
chmod +x app

# Copy the Systemd service definition to the right location
cp app.service /etc/systemd/system/app.service

# Start the service
systemctl daemon-reload
systemctl start app.service

# Ensure the file descriptor limit for the process is set
# --> Moved to ExecStartPost in the service script PR 830
# prlimit --nofile=200000:200000 --pid $(pidof app)