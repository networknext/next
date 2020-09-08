#!/bin/bash

export NEXT_CUSTOMER_PUBLIC_KEY=leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==
export NEXT_CUSTOMER_PRIVATE_KEY=leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn

num_servers=1
next_hostname='10.128.0.3'

print_usage() {
    printf "Usage: server-spawner.sh -n number\n\n"
    printf "n [number]\tNumber of servers to spawn\n"
    printf "s [string][Optional]\tThe IP address of the server backend\n"

    printf "Example:\n\n"
    printf "> server-spawner.sh -n 5 -s 10.128.0.75\n"

    print_env
}

while getopts 'n:s:h' flag; do
  case "${flag}" in
    n) num_servers="${OPTARG}" ;;
    s) next_hostname="${OPTARG}" ;;
    h) print_usage
       exit 1 ;;
    *) print_usage
       exit 1 ;;
  esac
done

trap "kill 0" EXIT

for ((r=0 ; r<${num_servers} ; r++)); do
port=$((50000 + r))
export NEXT_HOSTNAME="${next_hostname}"
export NEXT_LOG_LEVEL=0
export SERVER_IP="$(hostname -I) | awk '{print $1}'"
export SERVER_PORT="${port}"
/app/app &
pid="$!"
printf "PID ${pid}: Server opened\n"
done

printf "\nHit CTRL-C to exit and kill all spawned servers\n"

wait
