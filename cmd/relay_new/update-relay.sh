#!/bin/bash

artifact=

print_usage() {
	printf "Usage: update-relay.sh -a artifact\n\n"
	printf "a [string]\tPath to artifact on the relay server\n"

	printf "Example:\n\n"
	printf "> update-relay.sh -a /app/relay.tar.gz\n"
}

while getopts 'a:h' flag; do
	case "${flag}" in
		a) 
			artifact="${OPTARG}" ;;
		h) 
			print_usage
			exit 1 
			;;
		*)
			print_usage
			exit 1 
			;;
	esac
done

if [ $(id -u) = 0 ]; then

else
fi

