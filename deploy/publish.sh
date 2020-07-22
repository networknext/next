#!/bin/bash


TIMESTAMP=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA=$(git rev-parse --short HEAD)
RELEASE=$(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE=$(git log -1 --pretty=%B | tr '\n' ' ')
DIST_DIR="./dist"

ENV=
ARTIFACT_BUCKET=

publish() {
  printf "Publishing relay backend ${ENV} artifact to ${ARTIFACT_BUCKET} \n\n"
	#gsutil cp ${DIST_DIR}/relay_backend.${ENV}.tar.gz ${ARTIFACT_BUCKET}/relay_backend.${ENV}.tar.gz
	#gsutil setmeta -h "x-goog-meta-build-time:${TIMESTAMP}" -h "x-goog-meta-sha:${SHA}" -h "x-goog-meta-release:${RELEASE}" -h "x-goog-meta-commitMessage:${COMMITMESSAGE}" ${ARTIFACT_BUCKET}/relay_backend.${ENV}.tar.gz
	printf "done\n"
  printf "Publishing portal ${ENV} artifact... \n\n"
  #gsutil cp ${DIST_DIR}/portal.${ENV}.tar.gz ${ARTIFACT_BUCKET}/portal.${ENV}.tar.gz
  #gsutil setmeta -h "x-goog-meta-build-time:${TIMESTAMP}" -h "x-goog-meta-sha:${SHA}" -h "x-goog-meta-release:${RELEASE}" -h "x-goog-meta-commitMessage:${COMMITMESSAGE}" ${ARTIFACT_BUCKET}/portal.${ENV}.tar.gz
  printf "done\n"
  printf "Publishing billing ${ENV} artifact... \n\n"
	#gsutil cp ${DIST_DIR}/billing.${ENV}.tar.gz ${ARTIFACT_BUCKET}/billing.${ENV}.tar.gz
	#gsutil setmeta -h "x-goog-meta-build-time:${TIMESTAMP}" -h "x-goog-meta-sha:${SHA}" -h "x-goog-meta-release:${RELEASE}" -h "x-goog-meta-commitMessage:${COMMITMESSAGE}" ${ARTIFACT_BUCKET}/billing.${ENV}.tar.gz
	printf "done\n"
  printf "Publishing server backend ${ENV} artifact... \n\n"
	#gsutil cp ${DIST_DIR}/server_backend.${ENV}.tar.gz ${ARTIFACT_BUCKET}/server_backend.${ENV}.tar.gz
	#gsutil setmeta -h "x-goog-meta-build-time:${TIMESTAMP}" -h "x-goog-meta-sha:${SHA}" -h "x-goog-meta-release:${RELEASE}" -h "x-goog-meta-commitMessage:${COMMITMESSAGE}" ${ARTIFACT_BUCKET}/server_backend.${ENV}.tar.gz
	printf "done\n"
  printf "Publishing relay artifact... \n\n"
	#gsutil cp ${DIST_DIR}/relay.${ENV}.tar.gz ${ARTIFACT_BUCKET}/relay.${ENV}.tar.gz
	#gsutil acl set public-read ${ARTIFACT_BUCKET}/relay.${ENV}.tar.gz
	#gsutil setmeta -h 'Content-Type:application/xtar' -h 'Cache-Control:no-cache, max-age=0' ${ARTIFACT_BUCKET}/relay.${ENV}.tar.gz
	printf "done\n"
	printf "Publishing bootstrap script... \n\n"
	#gsutil cp ${DIST_DIR}/bootstrap.sh ${ARTIFACT_BUCKET}/bootstrap.sh
	printf "done\n"
}

print_usage() {
  printf "Usage: publish.sh -e environment -b bucket\n\n"
  printf "e [string]\tPublishing environment [dev, staging, prod]\n"
  printf "b [string]\tBucket name on GCP Storage\n"

  printf "Example:\n\n"
  printf "> publish.sh -e dev -b gs://development_artifacts\n"
}

if [ ! $# -eq 4 ]
then
  print_usage
  exit 1
fi

while getopts 'e:b:h' flag; do
  case "${flag}" in
    b) ARTIFACT_BUCKET="${OPTARG}" ;;
    e) ENV="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

publish
