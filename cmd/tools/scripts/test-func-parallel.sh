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
    test_direct_no_upgrade
    test_direct_with_backend
    test_fallback_to_direct_without_backend
    test_fallback_to_direct_is_not_sticky
    test_packets_over_next_with_relay_and_backend
    test_idempotent
    test_fallback_to_direct_when_backend_goes_down
    test_network_next_disabled_server
    test_network_next_disabled_client
    test_server_under_load
    test_reconnect_direct
    test_reconnect_next
    test_connect_to_another_server_direct
    test_connect_to_another_server_next
    test_route_switching
    test_on_off
    test_multipath
    test_uncommitted
    test_uncommitted_to_committed
    test_user_flags
    test_packet_loss_direct
)

for test in "${tests[@]}"; do
  docker run --name "$test" --env test="$test" func_tests:latest &
done

wait
