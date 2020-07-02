#!/bin/bash

mkdir -p /app
cd /app
gsutil cp gs://prod_artifacts/bootstrap.sh .
chmod +x bootstrap.sh
./bootstrap.sh -a gs://prod_artifacts/ARTIFACT.tar.gz