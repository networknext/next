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

build-artifacts() {
  printf "Building ${SERVICE} ${ENV} artifact... \n"
	mkdir -p ${DIST_DIR}/artifact/${SERVICE}
	if [ "$SERVICE" = "relay" ]; then
		cp ${DIST_DIR}/${SERVICE} ${DIST_DIR}/artifact/${SERVICE}/${SERVICE}
		cp ${DIR}/${SERVICE}/${SERVICE}.service ${DIST_DIR}/artifact/${SERVICE}/${SERVICE}.service
		cp ${DIR}/${SERVICE}/install.sh ${DIST_DIR}/artifact/${SERVICE}/install.sh
		cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.${ENV}.tar.gz ${SERVICE} ${SERVICE}.service install.sh && cd ../..
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
  printf "Usage: build-artifacts.sh -e environment -s service\n\n"
  printf "s [string]\Building environment [dev, staging, prod]\n"
  printf "e [string]\tService being built [portal, portal_cruncher, server_backend, etc]\n"

  printf "Example:\n\n"
  printf "> build-artifacts.sh -e dev -s portal\n"
}

if [ ! $# -eq 4 ]
then
  print_usage
  exit 1
fi

while getopts 'e:s:h' flag; do
  case "${flag}" in
    e) ENV="${OPTARG}" ;;
    s) SERVICE="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

build-artifacts
