#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

TIMESTAMP=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
SHA=$(git rev-parse --short HEAD)
RELEASE=$(shell git describe --tags --exact-match 2> /dev/null)
COMMITMESSAGE=$(git log -1 --pretty=%B | tr '\n' ' ')
SYSTEMD_SERVICE_FILE="app.service"
DIST_DIR="${DIR}/../dist"

SERVICE=

build-artifacts() {
  printf "Building ${SERVICE} artifact... \n"
	mkdir -p ${DIST_DIR}/artifact/${SERVICE}
	if [ "$SERVICE" = "load_test_server" ]; then
		cp ${DIST_DIR}/${SERVICE} ${DIST_DIR}/artifact/${SERVICE}/app
    cp ${DIR}/server-spawner.service ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
    cp ${DIR}/server-spawner.sh ${DIST_DIR}/artifact/${SERVICE}/server-spawner.sh
    cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.tar.gz app server-spawner.sh ${SYSTEMD_SERVICE_FILE} && cd ../..
  elif [ "$SERVICE" = "load_test_client" ]; then
		cp ${DIST_DIR}/${SERVICE} ${DIST_DIR}/artifact/${SERVICE}/app
    cp ${DIR}/client-spawner.service ${DIST_DIR}/artifact/${SERVICE}/${SYSTEMD_SERVICE_FILE}
    cp ${DIR}/client-spawner.sh ${DIST_DIR}/artifact/${SERVICE}/client-spawner.sh
    cd ${DIST_DIR}/artifact/${SERVICE} && tar -zcf ../../${SERVICE}.tar.gz app client-spawner.sh ${SYSTEMD_SERVICE_FILE} && cd ../..
	fi
	printf "${DIST_DIR}/${SERVICE}.tar.gz\n"
	printf "done\n"
}

print_usage() {
  printf "Usage: build-load-test-artifacts.sh -s service\n\n"
  printf "e [string]\tService being built [load_test_client, load_test_server, etc]\n"

  printf "Example:\n\n"
  printf "> build-load-test-artifacts.sh -s load_test_client\n"
}

if [ ! $# -eq 2 ]
then
  print_usage
  exit 1
fi

while getopts 's:h' flag; do
  case "${flag}" in
    s) SERVICE="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

build-artifacts
