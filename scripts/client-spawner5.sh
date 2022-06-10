#!/bin/bash

num_clients=1

print_usage() {
    printf "Usage: client-spawner5.sh -n number\n\n"
    printf "n [number]\tNumber of clients to spawn\n"

    printf "Example:\n\n"
    printf "> client-spawner5.sh -n 5\n"
}

while getopts 'n:h' flag; do
  case "${flag}" in
    n) num_clients="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

trap "kill 0" EXIT

for ((r=0 ; r<${num_clients} ; r++)); do
  ./client5 &
  pid="$!"
  printf "PID ${pid}: Client opened\n"
done

printf "\nHit CTRL-C to exit and kill all spawned clients\n"

wait
