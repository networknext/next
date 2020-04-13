#!/bin/bash

VERSION=288.0.0

export PATH=$PATH:$SEMAPHORE_CACHE_DIR/google-cloud-sdk/bin

if [ ! -f "$SEMAPHORE_CACHE_DIR/google-cloud-sdk/bin/gcloud" ]
then
  curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-$VERSION-linux-x86_64.tar.gz
  tar -x -C $SEMAPHORE_CACHE_DIR -f google-cloud-sdk-$VERSION-linux-x86_64.tar.gz
  gcloud --quiet components update --version $VERSION
  gcloud --quiet components update --version $VERSION beta
  gcloud --quiet components install --version $VERSION cloud-firestore-emulator
fi
