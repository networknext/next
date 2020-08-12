#!/bin/bash

export NEXT_CUSTOMER_PUBLIC_KEY=RBXUB1j4m2ielbODxa+nXivbAfhca69IrpCtRfWH9HAzkawKwnzSqA==
export NEXT_CUSTOMER_PRIVATE_KEY=RBXUB1j4m2gOwE2SrrbLAsFob6qCUkaIEfiOEkA453a1VgccIvr16Z6Vs4PFr6deK9sB+Fxrr0iukK1F9Yf0cDORrArCfNKo

num_clients=1

print_usage() {
    printf "Usage: client-spawner.sh -n number\n\n"
    printf "n [number]\tNumber of clients to spawn\n"

    printf "Example:\n\n"
    printf "> client-spawner.sh -n 5\n"

    print_env
}

print_env() {
  printf "\nShared environment\n"
  printf -- "------------------\n"
  printf "NEXT_CUSTOMER_PUBLIC_KEY: ${NEXT_CUSTOMER_PUBLIC_KEY}\n"
  printf "NEXT_CUSTOMER_PRIVATE_KEY: ${NEXT_CUSTOMER_PRIVATE_KEY}\n"
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
./dist/client &
pid="$!"
printf "PID ${pid}: Client opened\n"
done

print_env

printf "\nHit CTRL-C to exit and kill all spawned clients\n"

wait