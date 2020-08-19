#!/bin/bash

trap "remove_containers" EXIT

function remove_containers() {
    if [ "$(docker ps -aq --filter name="test_*")" ]; then
        docker rm --force $(docker ps -aq --filter name="test_*") > /dev/null
    fi
}

tests=(
    test_direct_default
    test_direct_upgrade
    test_network_next
    test_fallback_to_direct
    test_disable_network_next_on_server
    test_disable_network_next_on_client
    test_server_under_load
    test_reconnect_direct
    test_reconnect_next
    test_connect_to_another_server_direct
    test_connect_to_another_server_next
    test_route_switching
    test_on_off
    test_on_on_off
    test_multipath
    test_multipath_next_packet_loss
    test_multipath_fallback_to_direct
    test_uncommitted
    test_uncommitted_to_committed
    test_user_flags
    test_packet_loss_direct
    test_packet_loss_next
    test_idempotent
)

for test in "${tests[@]}"; do
    docker run --name "$test" --env test="$test" func_tests:latest &
done

exitCode="0"
processes="$(ps --ppid $$ | grep 'docker' | awk '{print $1}')"

while read process; do
    wait $process
    processExitCode=$?
    if [ "$processExitCode" != "0" ]; then
        exitCode=$processExitCode
    fi
done <<< "$processes"

wait
exit $exitCode