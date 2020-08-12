#!/bin/bash

num_servers=1

print_usage() {
    printf "Usage: server-spawner.sh -n number\n\n"
    printf "n [number]\tNumber of servers to spawn\n"

    printf "Example:\n\n"
    printf "> server-spawner.sh -n 5\n"

    print_env
}

while getopts 'n:h' flag; do
  case "${flag}" in
    n) num_servers="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

trap "kill 0" EXIT

for ((r=0 ; r<${num_servers} ; r++)); do
./dist/server &
pid="$!"
printf "PID ${pid}: Server opened\n"
done

printf "\nHit CTRL-C to exit and kill all spawned servers\n"

wait