#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

DIST_DIR="${DIR}/../dist"

ENV=
ARTIFACT_BUCKET=
SERVICE=

print_usage() {
  printf "Usage: build-artifacts.sh -e environment -b artifact bucket -s service\n\n"
  printf "e [string]\Building environment [dev, staging, prod]\n"
  printf "b [string]\Bucket for portal dist folder\n"
  printf "s [string]\Service being built [portal, admin]\n"

  printf "Example:\n\n"
  printf "> build-artifacts.sh -e dev -b gs://development_artifacts\n -s portal"
}


build-artifacts() {
  printf "Building ${ENV} artifact... \n"
  npm run build-${ENV}
  tar -zcf ${DIST_DIR}/../${SERVICE}-dist.${ENV}.tar.gz ${DIST_DIR}
  gsutil cp ${DIST_DIR}/../${SERVICE}-dist.${ENV}.tar.gz ${ARTIFACT_BUCKET}/${SERVICE}-dist.${ENV}.tar.gz
  printf "Done building and artifact was published\n"
}

if [ ! $# -eq 6 ]
then
  print_usage
  exit 1
fi

while getopts 'e:b:s:c:h' flag; do
  case "${flag}" in
    e) ENV="${OPTARG}" ;;
    b) ARTIFACT_BUCKET="${OPTARG}" ;;
    s) SERVICE="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

build-artifacts