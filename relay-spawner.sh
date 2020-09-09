#!/bin/bash

export RELAY_PUBLIC_KEY=9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=
export RELAY_PRIVATE_KEY=lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=
export RELAY_ROUTER_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=
export RELAY_BACKEND_HOSTNAME=http://127.0.0.1:30000

num_relays=1
start_port=10000

print_usage() {
    printf "Usage: relay-spawner.sh -n number -p port\n\n"
    printf "n [number]\tNumber of relays to spawn\n"
    printf "p [port]\tStarting port\n\n"

    printf "Example:\n\n"
    printf "> relay-spawner.sh -n 5 -p 20000\n"
    printf "PID 23665: Relay socket opened on port 20000\n"
    printf "PID 23666: Relay socket opened on port 20001\n"
    printf "PID 23667: Relay socket opened on port 20002\n"
    printf "PID 23668: Relay socket opened on port 20003\n"
    printf "PID 23669: Relay socket opened on port 20004\n"

    print_env
}

print_env() {
  printf "\nShared environment\n"
  printf -- "------------------\n"
  printf "RELAY_PUBLIC_KEY: ${RELAY_PUBLIC_KEY}\n"
  printf "RELAY_PRIVATE_KEY: ${RELAY_PRIVATE_KEY}\n"
  printf "RELAY_ROUTER_PUBLIC_KEY: ${RELAY_ROUTER_PUBLIC_KEY}\n"
  printf "RELAY_BACKEND_HOSTNAME: ${RELAY_BACKEND_HOSTNAME}\n"
}

while getopts 'n:p:h' flag; do
  case "${flag}" in
    n) num_relays="${OPTARG}" ;;
    p) start_port="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

trap "kill 0" EXIT

for ((r=0 ; r<${num_relays} ; r++)); do
port="$start_port"
if [[ ! "$start_port" -eq 0 ]]; then
	port=$((start_port+r))
fi
RELAY_ADDRESS=127.0.0.1:${port} ./bin/relay &
pid="$!"
printf "PID ${pid}: Relay socket opened on port ${port}\n"
done

print_env

printf "\nHit CTRL-C to exit and kill all spawned relays\n"

wait
