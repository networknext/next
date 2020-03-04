#!/bin/bash

trap "kill 0" EXIT

docker run --env test=test_direct_default func_tests:latest &
docker run --env test=test_direct_upgrade func_tests:latest &
docker run --env test=test_direct_no_upgrade func_tests:latest &
docker run --env test=test_direct_with_backend func_tests:latest &
docker run --env test=test_fallback_to_direct_without_backend func_tests:latest &
docker run --env test=test_fallback_to_direct_is_not_sticky func_tests:latest &
docker run --env test=test_packets_over_next_with_relay_and_backend func_tests:latest &
docker run --env test=test_idempotent func_tests:latest &
docker run --env test=test_fallback_to_direct_when_backend_goes_down func_tests:latest &
docker run --env test=test_network_next_disabled_server func_tests:latest &
docker run --env test=test_network_next_disabled_client func_tests:latest &
docker run --env test=test_server_under_load func_tests:latest &
docker run --env test=test_reconnect_direct func_tests:latest &
docker run --env test=test_reconnect_next func_tests:latest &
docker run --env test=test_connect_to_another_server_direct func_tests:latest &
docker run --env test=test_connect_to_another_server_next func_tests:latest &
docker run --env test=test_route_switching func_tests:latest &
docker run --env test=test_on_off func_tests:latest &
docker run --env test=test_multipath func_tests:latest &
docker run --env test=test_uncommitted func_tests:latest &
docker run --env test=test_uncommitted_to_committed func_tests:latest &
docker run --env test=test_user_flags func_tests:latest &
docker run --env test=test_packet_loss_direct func_tests:latest &

wait
