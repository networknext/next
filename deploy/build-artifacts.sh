#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

TIMESTAMP=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA=$(git rev-parse --short HEAD)
RELEASE=$(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE=$(git log -1 --pretty=%B | tr '\n' ' ')
SYSTEMD_SERVICE_FILE="app.service"
DIST_DIR="${DIR}/../dist"

ENV=

build-artifacts() {
  printf "Building relay backend ${ENV} artifact... \n"
	mkdir -p ${DIST_DIR}/artifact/relay_backend
	cp ${DIST_DIR}/relay_backend ${DIST_DIR}/artifact/relay_backend/app
	cp ${DIR}/../cmd/relay_backend/${ENV}.env ${DIST_DIR}/artifact/relay_backend/app.env
	cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/relay_backend/${SYSTEMD_SERVICE_FILE}
	cd ${DIST_DIR}/artifact/relay_backend && tar -zcf ../../relay_backend.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
	printf "${DIST_DIR}/relay_backend.${ENV}.tar.gz\n"
	printf "done\n"
	printf "Building portal ${ENV} artifact ... \n"
	mkdir -p ${DIST_DIR}/artifact/portal
	cp ${DIST_DIR}/portal ${DIST_DIR}/artifact/portal/app
	cp -r ${DIR}/../cmd/portal/public ${DIST_DIR}/artifact/portal
	cp ${DIR}/../cmd/portal/${ENV}.env ${DIST_DIR}/artifact/portal/app.env
	cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/portal/${SYSTEMD_SERVICE_FILE}
	cd ${DIST_DIR}/artifact/portal && tar -zcf ../../portal.${ENV}.tar.gz public app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
	printf "${DIST_DIR}/portal.${ENV}.tar.gz\n"
	printf "done\n"
	printf "Building billing ${ENV} artifact ... \n"
	mkdir -p ${DIST_DIR}/artifact/billing
	cp ${DIST_DIR}/billing ${DIST_DIR}/artifact/billing/app
	cp ${DIR}/../cmd/billing/${ENV}.env ${DIST_DIR}/artifact/billing/app.env
	cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/billing/${SYSTEMD_SERVICE_FILE}
	cd ${DIST_DIR}/artifact/billing && tar -zcf ../../billing.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
	printf "${DIST_DIR}/billing.${ENV}.tar.gz\n"
	printf "done\n"
	printf "Building server backend ${ENV} artifact ... \n"
	mkdir -p ${DIST_DIR}/artifact/server_backend
	cp ${DIST_DIR}/server_backend ${DIST_DIR}/artifact/server_backend/app
	cp ${DIR}/../cmd/server_backend/${ENV}.env ${DIST_DIR}/artifact/server_backend/app.env
	cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/server_backend/${SYSTEMD_SERVICE_FILE}
	cd ${DIST_DIR}/artifact/server_backend && tar -zcf ../../server_backend.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
	printf "${DIST_DIR}/server_backend.${ENV}.tar.gz\n"
	printf "done\n"
  printf "Building relay artifact ... \n"
	mkdir -p ${DIST_DIR}/artifact/relay
	cp ${DIST_DIR}/relay ${DIST_DIR}/artifact/relay/relay
	cp ${DIR}/relay/relay.service ${DIST_DIR}/artifact/relay/relay.service
	cp ${DIR}/relay/install.sh ${DIST_DIR}/artifact/relay/install.sh
	cd ${DIST_DIR}/artifact/relay && tar -zcf ../../relay.${ENV}.tar.gz relay relay.service install.sh && cd ../..
	printf "${DIST_DIR}/relay.${ENV}.tar.gz\n"
	printf "done\n"
	printf "Building portal_cruncher ${ENV} artifact... \n"
	mkdir -p ${DIST_DIR}/artifact/portal_cruncher
	cp ${DIST_DIR}/portal_cruncher ${DIST_DIR}/artifact/portal_cruncher/app
	cp ${DIR}/../cmd/portal_cruncher/${ENV}.env ${DIST_DIR}/artifact/portal_cruncher/app.env
	cp ${DIR}/${SYSTEMD_SERVICE_FILE} ${DIST_DIR}/artifact/portal_cruncher/${SYSTEMD_SERVICE_FILE}
	cd ${DIST_DIR}/artifact/portal_cruncher && tar -zcf ../../portal_cruncher.${ENV}.tar.gz app app.env ${SYSTEMD_SERVICE_FILE} && cd ../..
	printf "${DIST_DIR}/portal_cruncher.${ENV}.tar.gz\n"
	printf "done\n"
}

print_usage() {
  printf "Usage: publish.sh -e environment\n\n"
  printf "e [string]\tPublishing environment [dev, staging, prod]\n"

  printf "Example:\n\n"
  printf "> publish.sh -e dev\n"
}

if [ ! $# -eq 2 ]
then
  print_usage
  exit 1
fi

while getopts 'e:b:h' flag; do
  case "${flag}" in
    e) ENV="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

build-artifacts
