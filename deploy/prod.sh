#!/bin/bash

INSTANCE_GROUP=${1}
ARTIFACT=${2}

printf "Deploying ${ARTIFACT} to ${INSTANCE_GROUP}:\n"
for INSTANCE in $(gcloud compute --project "network-next-v3-prod" instance-groups managed list-instances "${INSTANCE_GROUP}" --zone "us-central1-a" --format "value(instance)")
do
    printf "\n\nInstance: ${INSTANCE}...\n\n"
    gcloud compute --project "network-next-v3-prod" ssh ${INSTANCE} -- "cd /app && sudo ./vm-update-app.sh -a ${ARTIFACT}"
done
printf "\ndone\n"