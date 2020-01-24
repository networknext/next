#!/bin/bash

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "Script must be sourced"
    exit 1
fi

kill_relays=0

print_help() {
    echo "Usage: relay-spawner.sh [-k] [-s port] [starting port] [ending port]"
    printf "k\tKill all relays that were spawned using this script\n"
    printf "h\tPrint this help menu\n"
    return 0
}

while getopts "kh" opt
do
    case $opt in
	(k) ((single_relay)) && printf "Cannot use -s with -k\n" && exit 1; kill_relays=1; shift ;;
	(h) print_help && exit 0 ;;
	(*) printf "Bad option '%s'\n" "$opt" && exit 1 ;;
    esac
done

start_relay() {
    port="$1"
    RELAY_ADDRESS="127.0.0.1:$port" make dev-relay-v2 > /dev/null &
    pid="$!"
    export RUNNING_RELAYS="$RUNNING_RELAYS:$pid"
    echo "Started Relay on port $port with pid: $pid"
}

start_relays() {
    begin_port="$1"
    end_port="$2"

    # enable the option to just spawn a single relay
    if [ -z "$end_port" ]; then
	end_port="$begin_port"
    fi

    if [ "$(( end_port - begin_port ))" -lt 0 ]; then
	echo "The lesser port must be first followed by the greater port"
	return
    fi

    echo "Spawning $(( end_port - begin_port + 1 )) relays between $begin_port and $end_port"

    for ((port = $(($begin_port)) ; port <= $(($end_port)) ; port++ )); do
	start_relay "$port"
    done
}

kill_relays() {
    IFS=':' read -ra RELAYS <<< "$RUNNING_RELAYS"
    for relay in ${RELAYS[@]}; do
        kill "$relay" > /dev/null
        echo "killed $relay"
    done

    export RUNNING_RELAYS=""
}

if [[ "$kill_relays" -eq 1 ]]; then
    kill_relays
else
    start_relays "$1" "$2"
fi

