#!/bin/bash

tag=
bucket=
artifact=

print_usage() {
    printf "Usage: bootstrap.sh -t tag -b bucket_name -a artifact\n\n"
    printf "t [string]\tGit tag of release\n"
    printf "b [string]\tBucket name on GCP Storage\n"
    printf "a [string]\tArtifact name on GCP Storage\n"

    printf "Example:\n\n"
    printf "> bootstrap.sh -t dev-007 -b gs://test_network_next_artifacts -a server_backend.tar.gz\n"
}

while getopts 't:b:a:h' flag; do
  case "${flag}" in
    t) tag="${OPTARG}" ;;
    b) bucket="${OPTARG}" ;;
    a) artifact="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

# Install the Google Cloud OS Agent
sudo su -c "echo 'deb http://packages.cloud.google.com/apt google-compute-engine-focal-stable main' > /etc/apt/sources.list.d/google-compute-engine.list"
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
sudo apt update
sudo apt -y install google-osconfig-agent

# Install the Google Ops Agent
curl -sSO https://dl.google.com/cloudagents/add-google-cloud-ops-agent-repo.sh
sudo bash add-google-cloud-ops-agent-repo.sh --also-install

# Create /app dir
rm -rf /app
mkdir -p /app
cd /app

# Copy libsodium from GCP Storage
gsutil cp "$bucket/$tag/libsodium.so" '/usr/local/lib'

# Refresh the known libs on the system
ldconfig

# Copy the required files for the service from GCP Storage
gsutil cp "$bucket/$tag/$artifact" artifact.tar.gz

# Uncompress the artifact files
tar -xvf artifact.tar.gz

# Set the app service binary to executable
chmod +x app

# Copy the Systemd service definition to the right location
cp app.service /etc/systemd/system/app.service

# Bump up the max socket read and write buffer sizes
sysctl -w net.core.rmem_max=1000000000
sysctl -w net.core.wmem_max=1000000000

# Reload services
systemctl daemon-reload
