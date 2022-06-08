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
ARTIFACT_BUCKET=

build-artifacts() {
  printf "Building ${SERVICE} ${ENV} artifact... \n"
	mkdir -p ${DIST_DIR}/artifact/${SERVICE}
	if [ "$SERVICE" = "relay" ]; then
		cp ${DIST_DIR}/${SERVICE} ${DIST_DIR}/artifact/${SERVICE}/${SERVICE}
		cp ${DIR}/${SERVICE}/${SERVICE}.service ${DIST_DIR}/artifact/${SERVICE}/${SERVICE}.service
		cp ${DIR}/${SERVICE}/install.sh ${DIST_DIR}/artifact/${SERVICE}/install.sh
		cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz ${SERVICE} ${SERVICE}.service install.sh && cd ../..
  elif [ "$SERVICE" = "debug_server_backend" ]; then
		cp ${DIST_DIR}/server_backend ${DIST_DIR}/artifact/${SERVICE}/app
		cp ${DIR}/../cmd/server_backend/${ENV}_debug.env ${DIST_DIR}/artifact/${SERVICE}/app.env
		cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
		cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
  elif [ "$SERVICE" = "debug_relay_backend" ]; then
		cp ${DIST_DIR}/relay_backend ${DIST_DIR}/artifact/${SERVICE}/app
		cp ${DIR}/../cmd/relay_backend/${ENV}_debug.env ${DIST_DIR}/artifact/${SERVICE}/app.env
		cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
		cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
  elif [ "$SERVICE" = "debug_billing" ]; then
        cp ${DIST_DIR}/billing ${DIST_DIR}/artifact/${SERVICE}/app
        cp ${DIR}/../cmd/billing/${ENV}_debug.env ${DIST_DIR}/artifact/${SERVICE}/app.env
        cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
        cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
  elif [ "$SERVICE" = "test_server4" ]; then
  	    cp ${DIST_DIR}/libnext4.so ${DIST_DIR}/artifact/${SERVICE}/libnext4.so
        cp ${DIST_DIR}/test_server4 ${DIST_DIR}/artifact/${SERVICE}/app
        cp ${DIR}/../cmd/test_server4/${ENV}.env ${DIST_DIR}/artifact/${SERVICE}/app.env
        cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
        cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz * ${SYSTEMD_SERVICE_FILE} && cd ../..
  elif [ "$SERVICE" = "test_server5" ]; then
  	    cp ${DIST_DIR}/libnext5.so ${DIST_DIR}/artifact/${SERVICE}/libnext5.so
        cp ${DIST_DIR}/test_server5 ${DIST_DIR}/artifact/${SERVICE}/app
        cp ${DIR}/../cmd/test_server5/${ENV}.env ${DIST_DIR}/artifact/${SERVICE}/app.env
        cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
        cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz * && cd ../..
  else
		cp ${DIST_DIR}/${SERVICE} ${DIST_DIR}/artifact/${SERVICE}/app
		cp ${DIR}/../cmd/${SERVICE}/${ENV}.env ${DIST_DIR}/artifact/${SERVICE}/app.env
		cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
		cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
	fi

  printf "${DIST_DIR}/${SERVICE}.${ENV}.tar.gz\n"
	printf "done\n"
}

print_usage() {
  printf "Usage: build-artifacts.sh -e environment -s service -b artifact bucket\n\n"
  printf "e [string]\tBuilding environment [dev, staging, prod]\n"
  printf "s [string]\tService being built [portal, portal_cruncher, server_backend, etc]\n"
  printf "b [string][optional]\tBucket for portal dist folder\n"

  printf "Example:\n\n"
  printf "> build-artifacts.sh -e dev -s portal -b gs://development_artifacts\n"
}

if [ ! $# -ge 4 ]
then
  print_usage
  exit 1
fi

while getopts 'e:b:s:h' flag; do
  case "${flag}" in
    e) ENV="${OPTARG}" ;;
    s) SERVICE="${OPTARG}" ;;
    b) ARTIFACT_BUCKET="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

build-artifacts
