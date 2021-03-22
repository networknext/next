#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

TIMESTAMP=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA=$(git rev-parse --short HEAD)
RELEASE=$(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE=$(git log -1 --pretty=%B | tr '\n' ' ')
DIST_DIR="${DIR}/../dist"

ENV=
CUSTOMER=
ARTIFACT_BUCKET=
ARTIFACT_NAME=
TYPE=

deploy-backend() {
  bootstrap='bootstrap.sh'
  if [ "$ARTIFACT_NAME" = 'ghost_army' ]; then
    bootstrap='ghost_army_bootstrap.sh'
  fi

  full_env_name='network-next-v3-'${ENV}
  if [ "$ENV" = 'nrb' ]; then
    full_env_name='network-next-new-relay-backend'
  fi

  COMMAND="cd /app && sudo gsutil cp ${ARTIFACT_BUCKET}/$bootstrap ./bootstrap.sh && sudo chmod +x ./bootstrap.sh && sudo ./bootstrap.sh -b ${ARTIFACT_BUCKET} -a ${ARTIFACT_NAME}.${ENV}.tar.gz"
  printf "Deploying ${CUSTOMER} ${TYPE}... \n"
  gcloud compute --project ${full_env_name} ssh ${TYPE}-${CUSTOMER} -- ${COMMAND}
  printf "done\n"
}

print_usage() {
  printf "Usage: deploy.sh -e environment -t type -c customer -b bucket\n\n"
  printf "e [string]\tDeployment environment [dev, staging, prod]\n"
  printf "t [string]\tBackend type [relay/server]\n"
  printf "c [string]\tCustomer\n"
  printf "b [string]\tBucket name on GCP Storage\n"
  printf "n [string]\tArtifact name on GCP Storage\n"

  printf "Example:\n\n"
  printf "> deploy.sh -e prod -c psyonix -t server-backend -n server_backend -b gs://prod_artifacts\n"
}

if [ ! $# -eq 10 ]
then
  print_usage
  exit 1
fi

while getopts 'e:c:t:b:n:h' flag; do
  case "${flag}" in
    b) ARTIFACT_BUCKET="${OPTARG}" ;;
    n) ARTIFACT_NAME="${OPTARG}" ;;
    t) TYPE="${OPTARG}" ;;
    c) CUSTOMER="${OPTARG}" ;;
    e) ENV="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

deploy-backend
