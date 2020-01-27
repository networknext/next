#!/bin/bash

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "Script must be sourced"
    exit 1
fi

kill_relays=0
flush_redis=0
use_relay_count=0

# because it must be sourced, reset OPTIND to 1
OPTIND=1
while getopts ":n:khr" opt; do
    case "$opt" in
	(n)
	    use_relay_count=1
	    relay_count="$OPTARG"
	    ;;
	(k) 
	    kill_relays=1
	    ;;
	(h) 
	    echo "Usage: relay-spawner.sh [-k] [-s port] [starting port] [ending port]"
	    printf "k\tKill all relays that were spawned using this script\n"
	    printf "h\tPrint this help menu\n"
	    printf "r\tReset redis before any other operation\n"
	    return 0
	    ;;
	(r) 
	    flush_redis=1
	    ;;
	(\?) 
	    echo "Bad option '$OPTARG'" && return 1 
	    ;;
	(:)
	    echo "Bad param: '$OPTARG'" && return 1
	    ;;
    esac
done

shift $(( OPTIND - 1 ))

if [[ "$flush_redis" -eq 1 ]]; then
    echo "Clearing Redis: $(redis-cli FLUSHALL)"
fi

if [[ "$kill_relays" == 1 ]]; then
    IFS=':' read -ra RELAYS <<< "$RUNNING_RELAYS"
    for relay in ${RELAYS[@]}; do
	kill "$relay" > /dev/null
	echo "killed $relay"
    done

    export RUNNING_RELAYS=""
else
    if [[ "$#" == 0 ]]; then
	if [[ "$flush_redis" == 0 ]]; then
	    echo "You must supply a port number"
	    return 1
	else
	    return 0
	fi
    fi

    begin_port="$1"
    end_port="$2"

    if [[ "$use_relay_count" == 1 ]]; then
	end_port="$relay_count"
    elif [ -z "$end_port" ]; then
	# enable the option to just spawn a single relay
	end_port="$begin_port"
    fi

    if [ "$(( end_port - begin_port ))" -lt 0 ]; then
	echo "The lesser port must be first followed by the greater port"
	return
    fi

    echo "Spawning $(( end_port - begin_port + 1 )) relays between $begin_port and $end_port"

    for ((port=${begin_port} ; port<=${end_port} ; port++)); do
	RELAY_ADDRESS="127.0.0.1:$port" make dev-relay-v2 > /dev/null &
	pid="$!"
	export RUNNING_RELAYS="$RUNNING_RELAYS:$pid"
	echo "Started Relay on port $port with pid: $pid"
    done
fi

