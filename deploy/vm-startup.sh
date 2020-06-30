#!/bin/bash

apt-get install libsodium23

mkdir -p /app
cd /app
gsutil cp gs://prod_artifacts/vm-update-app.sh .
chmod +x vm-update-app.sh
./vm-update-app.sh -a gs://prod_artifacts/ARTIFACT.tar.gz