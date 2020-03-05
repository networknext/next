#!/bin/bash

trap "remove_containers" EXIT

function remove_containers() {
    if [ "$(docker ps -aq --filter name="test_*")" ]; then
        docker rm --force $(docker ps -aq --filter name="test_*") > /dev/null
    fi
}

docker run --name test_direct_default --env test=test_direct_default func_tests:latest &
docker run --name test_direct_upgrade --env test=test_direct_upgrade func_tests:latest &
docker run --name test_direct_no_upgrade --env test=test_direct_no_upgrade func_tests:latest &
docker run --name test_direct_with_backend --env test=test_direct_with_backend func_tests:latest &
docker run --name test_fallback_to_direct_without_backend --env test=test_fallback_to_direct_without_backend func_tests:latest &
docker run --name test_fallback_to_direct_is_not_sticky --env test=test_fallback_to_direct_is_not_sticky func_tests:latest &
docker run --name test_packets_over_next_with_relay_and_backend --env test=test_packets_over_next_with_relay_and_backend func_tests:latest &
docker run --name test_idempotent --env test=test_idempotent func_tests:latest &
docker run --name test_fallback_to_direct_when_backend_goes_down --env test=test_fallback_to_direct_when_backend_goes_down func_tests:latest &
docker run --name test_network_next_disabled_server --env test=test_network_next_disabled_server func_tests:latest &
docker run --name test_network_next_disabled_client --env test=test_network_next_disabled_client func_tests:latest &
docker run --name test_server_under_load --env test=test_server_under_load func_tests:latest &
docker run --name test_reconnect_direct --env test=test_reconnect_direct func_tests:latest &
docker run --name test_reconnect_next --env test=test_reconnect_next func_tests:latest &
docker run --name test_connect_to_another_server_direct --env test=test_connect_to_another_server_direct func_tests:latest &
docker run --name test_connect_to_another_server_next --env test=test_connect_to_another_server_next func_tests:latest &
docker run --name test_route_switching --env test=test_route_switching func_tests:latest &
docker run --name test_on_off --env test=test_on_off func_tests:latest &
docker run --name test_multipath --env test=test_multipath func_tests:latest &
docker run --name test_uncommitted --env test=test_uncommitted func_tests:latest &
docker run --name test_uncommitted_to_committed --env test=test_uncommitted_to_committed func_tests:latest &
docker run --name test_user_flags --env test=test_user_flags func_tests:latest &
docker run --name test_packet_loss_direct --env test=test_packet_loss_direct func_tests:latest &
docker run --name test_packet_loss_next --env test=test_packet_loss_next func_tests:latest &

wait
