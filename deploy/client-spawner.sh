#!/bin/bash

export NEXT_CUSTOMER_PUBLIC_KEY=leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==
export NEXT_CUSTOMER_PRIVATE_KEY=leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn

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

ips=$(gcloud compute instances list --filter='load-test-server-staging AND status=RUNNING' --format='get(networkInterfaces[0].networkIP)')

for ((r=0 ; r<${num_clients} ; r++)); do
export NEXT_LOG_LEVEL=1
export CORES=1
echo $ips | /app/app &
pid="$!"
printf "PID ${pid}: Client opened\n"
done

print_env

printf "\nHit CTRL-C to exit and kill all spawned clients\n"

wait
