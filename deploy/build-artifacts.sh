#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

TIMESTAMP=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA=$(git rev-parse --short HEAD)
RELEASE=$(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE=$(git log -1 --pretty=%B | tr '\n' ' ')
SYSTEMD_SERVICE_FILE="app.service"
DIST_DIR="${DIR}/../dist"

ENV=
SERVICE=
CUSTOMER=
ARTIFACT_BUCKET=

build-artifacts() {
  printf "Building ${SERVICE} ${ENV} artifact... \n"
	mkdir -p ${DIST_DIR}/artifact/${SERVICE}
	if [ "$SERVICE" = "relay" ]; then
		cp ${DIST_DIR}/${SERVICE} ${DIST_DIR}/artifact/${SERVICE}/${SERVICE}
		cp ${DIR}/${SERVICE}/${SERVICE}.service ${DIST_DIR}/artifact/${SERVICE}/${SERVICE}.service
		cp ${DIR}/${SERVICE}/install.sh ${DIST_DIR}/artifact/${SERVICE}/install.sh
		cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz ${SERVICE} ${SERVICE}.service install.sh && cd ../..
  elif [ "$SERVICE" = "portal" ] || [ "$SERVICE" = "portal-test" ]; then
    gsutil cp ${ARTIFACT_BUCKET}/${SERVICE}-dist.${ENV}.tar.gz ${DIST_DIR}/artifact/${SERVICE}/.
    tar -xvf ${DIST_DIR}/artifact/${SERVICE}/${SERVICE}-dist.${ENV}.tar.gz --directory ${DIST_DIR}/artifact/${SERVICE}
    cp ${DIST_DIR}/portal ${DIST_DIR}/artifact/${SERVICE}/app
    cp ./cmd/portal/${ENV}.env ${DIST_DIR}/artifact/${SERVICE}/app.env
    cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
    cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz dist app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
  elif [ "$SERVICE" = "server_backend" ] && [ -n "$CUSTOMER" ]; then
		cp ${DIST_DIR}/${SERVICE} ${DIST_DIR}/artifact/${SERVICE}/app
		cp ${DIR}/../cmd/${SERVICE}/${ENV}.env ${DIST_DIR}/artifact/${SERVICE}/app.env
		cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
		cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}-${CUSTOMER}.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
	else
		cp ${DIST_DIR}/${SERVICE} ${DIST_DIR}/artifact/${SERVICE}/app
		cp ${DIR}/../cmd/${SERVICE}/${ENV}.env ${DIST_DIR}/artifact/${SERVICE}/app.env
		cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
		cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
	fi

  if [ "$SERVICE" = "server_backend" ] && [ -n "$CUSTOMER" ]; then
	  printf "${DIST_DIR}/${SERVICE}-${CUSTOMER}.${ENV}.tar.gz\n"
  else
    printf "${DIST_DIR}/${SERVICE}.${ENV}.tar.gz\n"
  fi
	printf "done\n"
}

print_usage() {
  printf "Usage: build-artifacts.sh -e environment -s service -b artifact bucket\n\n"
  printf "e [string]\tBuilding environment [dev, staging, prod]\n"
  printf "s [string]\tService being built [portal, portal_cruncher, server_backend, etc]\n"
  printf "c [string][optional]\tCustomer server backend name [esl-22dr]\n"
  printf "b [string][optional]\tBucket for portal dist folder\n"

  printf "Example:\n\n"
  printf "> build-artifacts.sh -e dev -s portal\n"
}

if [ ! $# -ge 4 ]
then
  print_usage
  exit 1
fi

while getopts 'e:b:s:c:h' flag; do
  case "${flag}" in
    e) ENV="${OPTARG}" ;;
    s) SERVICE="${OPTARG}" ;;
    c) CUSTOMER="${OPTARG}" ;;
    b) ARTIFACT_BUCKET="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

build-artifacts
