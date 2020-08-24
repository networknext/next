#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

TIMESTAMP=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA=$(git rev-parse --short HEAD)
RELEASE=$(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE=$(git log -1 --pretty=%B | tr '\n' ' ')
DIST_DIR="${DIR}/../dist"

ARTIFACT_BUCKET=
SERVICE=

publish() {
  if [ "$SERVICE" = "staging_servers.txt" ]; then
    printf "Publishing ${SERVICE} to ${ARTIFACT_BUCKET} \n\n"
    gsutil cp ${DIR}/${SERVICE} ${ARTIFACT_BUCKET}/${SERVICE} || exit 1
  else
    printf "Publishing ${SERVICE} artifact to ${ARTIFACT_BUCKET} \n\n"
    gsutil cp ${DIST_DIR}/${SERVICE}.tar.gz ${ARTIFACT_BUCKET}/${SERVICE}.tar.gz || exit 1
  fi
  printf "done\n"
}

print_usage() {
  printf "Usage: publish.sh -e environment -s service -b bucket\n\n"
  printf "s [string]\tService being published [portal, portal_cruncher, server_backend, etc]\n"
  printf "b [string]\tBucket name on GCP Storage\n"

  printf "Example:\n\n"
  printf "> publish.sh -e dev -s portal -b gs://development_artifacts\n"
}

if [ ! $# -eq 4 ]
then
  print_usage
  exit 1
fi

while getopts 'e:s:b:h' flag; do
  case "${flag}" in
    b) ARTIFACT_BUCKET="${OPTARG}" ;;
    s) SERVICE="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

publish
