#!/bin/bash

mig=
bucket=
env=

print_usage() {
    printf "Usage: deploy-portal.sh\n\n"
    # printf "b [string]\tBucket name on GCP Storage\n"

    printf "Example:\n\n"
    printf "> deploy-portal.sh\n"
}

while getopts 'm:b:e:h' flag; do
  case "${flag}" in
    m) mig="${OPTARG}" ;;
    b) bucket="${OPTARG}" ;;
    e) env="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

# ssh into all the VMs under the specified MIG

VMs=$(gcloud compute --project network-next-v3-${env} instance-groups managed list-instances ${mig} --zone us-central1-a --format "value(instance)")
COMMAND="cd /portal && sudo gsutil cp ${bucket}/portal-dist.${env}.tar.gz artifact.tar.gz && tar -xvf artifact.tar.gz"

while IFS= read -r vm; do
  printf "Deploying Frontend code to ${vm}... \n"
  gcloud compute --project network-next-v3-${env} ssh --zone us-central1-a ${vm} -- ${COMMAND}
  printf "done\n"
done <<< "$VMs"
