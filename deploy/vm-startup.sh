#!/bin/bash

mkdir -p /app
cd /app
gsutil cp gs://artifacts.network-next-v3-dev.appspot.com/vm-update-app.sh .
chmod +x vm-update-app.sh
./vm-update-app.sh -a gs://artifacts.network-next-v3-dev.appspot.com/ARTIFACT.tar.gz