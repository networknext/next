#!/bin/bash

function start-relays() {
    export RUNNING_RELAYS=()

    begin_port="$1"
    end_port="$2"

    for ((i = $(($begin_port)) ; i <= $(($end_port)) ; i++ )); do
        RELAY_ADDRESS="127.0.0.1:$i" make dev-relay-v2 &
        pid="$!"
        RUNNING_RELAYS+=("$pid")
        echo "Started Relay with pid: $pid"
    done
}

export -f start-relays

function kill-relays() {
    for relay in ${RUNNING_RELAYS[@]}; do
        kill "$relay"
        echo "killed $relay"
    done
}

export -f kill-relays
