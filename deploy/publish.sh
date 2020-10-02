#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

TIMESTAMP=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA=$(git rev-parse --short HEAD)
RELEASE=$(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE=$(git log -1 --pretty=%B | tr '\n' ' ')
DIST_DIR="${DIR}/../dist"

ENV=
ARTIFACT_BUCKET=
SERVICE=
CUSTOMER=

publish() {
  if { [ "$SERVICE" = "server_backend" ] || [ "$SERVICE" = "server_backend4" ]; } && [ -n "$CUSTOMER" ]; then
    printf "Publishing ${SERVICE}-${CUSTOMER} ${ENV} artifact to ${ARTIFACT_BUCKET} \n\n"
    gsutil cp ${DIST_DIR}/${SERVICE}-${CUSTOMER}.${ENV}.tar.gz ${ARTIFACT_BUCKET}/${SERVICE}-${CUSTOMER}.${ENV}.tar.gz
    gsutil setmeta -h "x-goog-meta-build-time:${TIMESTAMP}" -h "x-goog-meta-sha:${SHA}" -h "x-goog-meta-release:${RELEASE}" -h "x-goog-meta-commitMessage:${COMMITMESSAGE}" ${ARTIFACT_BUCKET}/${SERVICE}-${CUSTOMER}.${ENV}.tar.gz
  else
    printf "Publishing ${SERVICE} ${ENV} artifact to ${ARTIFACT_BUCKET} \n\n"
    gsutil cp ${DIST_DIR}/${SERVICE}.${ENV}.tar.gz ${ARTIFACT_BUCKET}/${SERVICE}.${ENV}.tar.gz
    gsutil setmeta -h "x-goog-meta-build-time:${TIMESTAMP}" -h "x-goog-meta-sha:${SHA}" -h "x-goog-meta-release:${RELEASE}" -h "x-goog-meta-commitMessage:${COMMITMESSAGE}" ${ARTIFACT_BUCKET}/${SERVICE}.${ENV}.tar.gz
  fi

  if [ "$SERVICE" = "relay" ]; then
		gsutil acl set public-read ${ARTIFACT_BUCKET}/${SERVICE}.${ENV}.tar.gz
		gsutil setmeta -h 'Content-Type:application/xtar' -h 'Cache-Control:no-cache, max-age=0' ${ARTIFACT_BUCKET}/${SERVICE}.${ENV}.tar.gz
	fi
	printf "done\n"
}

print_usage() {
  printf "Usage: publish.sh -e environment -s service -b bucket\n\n"
  printf "e [string]\tPublishing environment [dev, staging, prod]\n"
  printf "s [string]\tService being published [portal, portal_cruncher, server_backend, etc]\n"
  printf "b [string]\tBucket name on GCP Storage\n"
  printf "c [string][optional]\tCustomer server backend name [esl-22dr]\n"

  printf "Example:\n\n"
  printf "> publish.sh -e dev -s portal -b gs://development_artifacts\n"
}

if [ ! $# -ge 6 ]
then
  print_usage
  exit 1
fi

while getopts 'e:s:b:c:h' flag; do
  case "${flag}" in
    b) ARTIFACT_BUCKET="${OPTARG}" ;;
    s) SERVICE="${OPTARG}" ;;
    e) ENV="${OPTARG}" ;;
    c) CUSTOMER="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

publish
