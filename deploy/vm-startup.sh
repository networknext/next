#!/bin/bash

artifact=

print_usage() {
    printf "Usage: vm-startup.sh -a artifact\n\n"
    printf "a [path]\tPath to a valid gs:// path to copy from\n"

    printf "Example:\n\n"
    printf "> vm-startup.sh -a gs://path/to/artifact.tar.gz\n"

    print_env
}

while getopts 'a:h' flag; do
  case "${flag}" in
    a) artifact="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

mkdir -p ./app
cd ./app
gsutil cp gs://artifacts.network-next-v3-dev.appspot.com/GeoLite2-City.mmdb .
gsutil cp ${artifact} artifact.tar.gz
tar -xvf artifact.tar.gz
source app.env
chmod +x app
./app
