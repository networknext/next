/*
 * Network Next Relay.
 * Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
 */

#include <cassert>
#include <string.h>
#include <stdio.h>
#include <cinttypes>
#include <stdarg.h>
#include <sodium.h>
#include <math.h>
#include <map>
#include <float.h>
#include <signal.h>
#include <curl/curl.h>

#include "sysinfo.hpp"
#include "config.hpp"
#include "test/test.hpp"
#include "bench/bench.hpp"
#include "util.hpp"

#include "encoding/base64.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"

#include "relay/relay.hpp"
#include "net/address.hpp"
#include "relay/relay_continue_token.hpp"
#include "relay/relay_ping_history.hpp"
#include "relay/relay_platform.hpp"
#include "relay/relay_replay_protection.hpp"
#include "relay/relay_route_token.hpp"

namespace
{
    volatile uint64_t quit = 0;

    void interrupt_handler(int signal)
    {
        (void)signal;
        quit = 1;
    }
    static relay::relay_platform_thread_return_t RELAY_PLATFORM_THREAD_FUNC receive_thread_function(void* context)
    {
        relay::relay_t* relay = (relay::relay_t*)context;

        uint8_t packet_data[RELAY_MAX_PACKET_BYTES];

        while (!quit) {
            legacy::relay_address_t from;
            const int packet_bytes =
                relay_platform_socket_receive_packet(relay->socket, &from, packet_data, sizeof(packet_data));
            if (packet_bytes == 0)
                continue;
            if (packet_data[0] == RELAY_PING_PACKET && packet_bytes == 9) {
                packet_data[0] = RELAY_PONG_PACKET;
                relay_platform_socket_send_packet(relay->socket, &from, packet_data, 9);
            } else if (packet_data[0] == RELAY_PONG_PACKET && packet_bytes == 9) {
                relay_platform_mutex_acquire(relay->mutex);
                const uint8_t* p = packet_data + 1;
                uint64_t sequence = encoding::read_uint64(&p);
                relay_manager_process_pong(relay->relay_manager, &from, sequence);
                relay_platform_mutex_release(relay->mutex);
            } else if (packet_data[0] == RELAY_ROUTE_REQUEST_PACKET) {
                if (packet_bytes < int(1 + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES * 2)) {
                    relay_printf("ignoring route request. bad packet size (%d)", packet_bytes);
                    continue;
                }
                uint8_t* p = &packet_data[1];
                relay::relay_route_token_t token;
                if (relay::relay_read_encrypted_route_token(&p, &token, relay->router_public_key, relay->relay_private_key) !=
                    RELAY_OK) {
                    relay_printf("ignoring route request. could not read route token");
                    continue;
                }
                if (token.expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                uint64_t hash = token.session_id ^ token.session_version;
                if (relay->sessions->find(hash) == relay->sessions->end()) {
                    relay::relay_session_t* session = (relay::relay_session_t*)malloc(sizeof(relay::relay_session_t));
                    assert(session);
                    session->expire_timestamp = token.expire_timestamp;
                    session->session_id = token.session_id;
                    session->session_version = token.session_version;
                    session->client_to_server_sequence = 0;
                    session->server_to_client_sequence = 0;
                    session->kbps_up = token.kbps_up;
                    session->kbps_down = token.kbps_down;
                    session->prev_address = from;
                    session->next_address = token.next_address;
                    memcpy(session->private_key, token.private_key, crypto_box_SECRETKEYBYTES);
                    relay_replay_protection_reset(&session->replay_protection_client_to_server);
                    relay_replay_protection_reset(&session->replay_protection_server_to_client);
                    relay->sessions->insert(std::make_pair(hash, session));
                    printf("session created: %" PRIx64 ".%d\n", token.session_id, token.session_version);
                }
                packet_data[RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES] = RELAY_ROUTE_REQUEST_PACKET;
                relay_platform_socket_send_packet(relay->socket,
                    &token.next_address,
                    packet_data + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES,
                    packet_bytes - RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES);
            } else if (packet_data[0] == RELAY_ROUTE_RESPONSE_PACKET) {
                if (packet_bytes != RELAY_HEADER_BYTES) {
                    continue;
                }
                uint8_t type;
                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if (relay::relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
                        &type,
                        &sequence,
                        &session_id,
                        &session_version,
                        packet_data,
                        packet_bytes) != RELAY_OK) {
                    continue;
                }
                uint64_t hash = session_id ^ session_version;
                relay::relay_session_t* session = (*(relay->sessions))[hash];
                if (!session) {
                    continue;
                }
                if (session->expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
                if (clean_sequence <= session->server_to_client_sequence) {
                    continue;
                }
                session->server_to_client_sequence = clean_sequence;
                if (relay::relay_verify_header(
                        RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet_data, packet_bytes) != RELAY_OK) {
                    continue;
                }
                relay_platform_socket_send_packet(relay->socket, &session->prev_address, packet_data, packet_bytes);
            } else if (packet_data[0] == RELAY_CONTINUE_REQUEST_PACKET) {
                if (packet_bytes < int(1 + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES * 2)) {
                    relay_printf("ignoring continue request. bad packet size (%d)", packet_bytes);
                    continue;
                }
                uint8_t* p = &packet_data[1];
                relay::relay_continue_token_t token;
                if (relay_read_encrypted_continue_token(&p, &token, relay->router_public_key, relay->relay_private_key) !=
                    RELAY_OK) {
                    relay_printf("ignoring continue request. could not read continue token");
                    continue;
                }
                if (token.expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                uint64_t hash = token.session_id ^ token.session_version;
                relay::relay_session_t* session = (*(relay->sessions))[hash];
                if (!session) {
                    continue;
                }
                if (session->expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                if (session->expire_timestamp != token.expire_timestamp) {
                    printf("session continued: %" PRIx64 ".%d\n", token.session_id, token.session_version);
                }
                session->expire_timestamp = token.expire_timestamp;
                packet_data[RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES] = RELAY_CONTINUE_REQUEST_PACKET;
                relay_platform_socket_send_packet(relay->socket,
                    &session->next_address,
                    packet_data + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES,
                    packet_bytes - RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES);
            } else if (packet_data[0] == RELAY_CONTINUE_RESPONSE_PACKET) {
                if (packet_bytes != RELAY_HEADER_BYTES) {
                    continue;
                }
                uint8_t type;
                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if (relay::relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
                        &type,
                        &sequence,
                        &session_id,
                        &session_version,
                        packet_data,
                        packet_bytes) != RELAY_OK) {
                    continue;
                }
                uint64_t hash = session_id ^ session_version;
                relay::relay_session_t* session = (*(relay->sessions))[hash];
                if (!session) {
                    continue;
                }
                if (session->expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
                if (clean_sequence <= session->server_to_client_sequence) {
                    continue;
                }
                session->server_to_client_sequence = clean_sequence;
                if (relay::relay_verify_header(
                        RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet_data, packet_bytes) != RELAY_OK) {
                    continue;
                }
                relay_platform_socket_send_packet(relay->socket, &session->prev_address, packet_data, packet_bytes);
            } else if (packet_data[0] == RELAY_CLIENT_TO_SERVER_PACKET) {
                if (packet_bytes <= RELAY_HEADER_BYTES || packet_bytes > RELAY_HEADER_BYTES + RELAY_MTU) {
                    continue;
                }
                uint8_t type;
                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if (relay::relay_peek_header(RELAY_DIRECTION_CLIENT_TO_SERVER,
                        &type,
                        &sequence,
                        &session_id,
                        &session_version,
                        packet_data,
                        packet_bytes) != RELAY_OK) {
                    continue;
                }
                uint64_t hash = session_id ^ session_version;
                relay::relay_session_t* session = (*(relay->sessions))[hash];
                if (!session) {
                    continue;
                }
                if (session->expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
                if (relay_replay_protection_already_received(&session->replay_protection_client_to_server, clean_sequence)) {
                    continue;
                }
                relay_replay_protection_advance_sequence(&session->replay_protection_client_to_server, clean_sequence);
                if (relay::relay_verify_header(
                        RELAY_DIRECTION_CLIENT_TO_SERVER, session->private_key, packet_data, packet_bytes) != RELAY_OK) {
                    continue;
                }
                relay_platform_socket_send_packet(relay->socket, &session->next_address, packet_data, packet_bytes);
            } else if (packet_data[0] == RELAY_SERVER_TO_CLIENT_PACKET) {
                if (packet_bytes <= RELAY_HEADER_BYTES || packet_bytes > RELAY_HEADER_BYTES + RELAY_MTU) {
                    continue;
                }
                uint8_t type;
                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if (relay::relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
                        &type,
                        &sequence,
                        &session_id,
                        &session_version,
                        packet_data,
                        packet_bytes) != RELAY_OK) {
                    continue;
                }
                uint64_t hash = session_id ^ session_version;
                relay::relay_session_t* session = (*(relay->sessions))[hash];
                if (!session) {
                    continue;
                }
                if (session->expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
                if (relay_replay_protection_already_received(&session->replay_protection_server_to_client, clean_sequence)) {
                    continue;
                }
                relay_replay_protection_advance_sequence(&session->replay_protection_server_to_client, clean_sequence);
                if (relay::relay_verify_header(
                        RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet_data, packet_bytes) != RELAY_OK) {
                    continue;
                }
                relay_platform_socket_send_packet(relay->socket, &session->prev_address, packet_data, packet_bytes);
            } else if (packet_data[0] == RELAY_SESSION_PING_PACKET) {
                if (packet_bytes > RELAY_HEADER_BYTES + 32) {
                    continue;
                }
                uint8_t type;
                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if (relay::relay_peek_header(RELAY_DIRECTION_CLIENT_TO_SERVER,
                        &type,
                        &sequence,
                        &session_id,
                        &session_version,
                        packet_data,
                        packet_bytes) != RELAY_OK) {
                    continue;
                }
                uint64_t hash = session_id ^ session_version;
                relay::relay_session_t* session = (*(relay->sessions))[hash];
                if (!session) {
                    continue;
                }
                if (session->expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
                if (clean_sequence <= session->client_to_server_sequence) {
                    continue;
                }
                session->client_to_server_sequence = clean_sequence;
                if (relay::relay_verify_header(
                        RELAY_DIRECTION_CLIENT_TO_SERVER, session->private_key, packet_data, packet_bytes) != RELAY_OK) {
                    continue;
                }
                relay_platform_socket_send_packet(relay->socket, &session->next_address, packet_data, packet_bytes);
            } else if (packet_data[0] == RELAY_SESSION_PONG_PACKET) {
                if (packet_bytes > RELAY_HEADER_BYTES + 32) {
                    continue;
                }
                uint8_t type;
                uint64_t sequence;
                uint64_t session_id;
                uint8_t session_version;
                if (relay::relay_peek_header(RELAY_DIRECTION_SERVER_TO_CLIENT,
                        &type,
                        &sequence,
                        &session_id,
                        &session_version,
                        packet_data,
                        packet_bytes) != RELAY_OK) {
                    continue;
                }
                uint64_t hash = session_id ^ session_version;
                relay::relay_session_t* session = (*(relay->sessions))[hash];
                if (!session) {
                    continue;
                }
                if (session->expire_timestamp < relay_timestamp(relay)) {
                    continue;
                }
                uint64_t clean_sequence = relay::relay_clean_sequence(sequence);
                if (clean_sequence <= session->server_to_client_sequence) {
                    continue;
                }
                session->server_to_client_sequence = clean_sequence;
                if (relay::relay_verify_header(
                        RELAY_DIRECTION_SERVER_TO_CLIENT, session->private_key, packet_data, packet_bytes) != RELAY_OK) {
                    continue;
                }
                relay_platform_socket_send_packet(relay->socket, &session->prev_address, packet_data, packet_bytes);
            } else if (packet_data[0] == RELAY_NEAR_PING_PACKET) {
                if (packet_bytes != 1 + 8 + 8 + 8 + 8) {
                    continue;
                }
                packet_data[0] = RELAY_NEAR_PONG_PACKET;
                relay_platform_socket_send_packet(relay->socket, &from, packet_data, packet_bytes - 16);
            }
        }

        RELAY_PLATFORM_THREAD_RETURN();
    }

    static relay::relay_platform_thread_return_t RELAY_PLATFORM_THREAD_FUNC ping_thread_function(void* context)
    {
        relay::relay_t* relay = (relay::relay_t*)context;

        while (!quit) {
            relay::relay_platform_mutex_acquire(relay->mutex);

            if (relay->relays_dirty) {
                relay::relay_manager_update(relay->relay_manager, relay->num_relays, relay->relay_ids, relay->relay_addresses);
                relay->relays_dirty = false;
            }

            double current_time = relay::relay_platform_time();

            struct ping_data_t
            {
                uint64_t sequence;
                legacy::relay_address_t address;
            };

            int num_pings = 0;
            ping_data_t pings[MAX_RELAYS];

            for (int i = 0; i < relay->relay_manager->num_relays; ++i) {
                if (relay->relay_manager->relay_last_ping_time[i] + RELAY_PING_TIME <= current_time) {
                    pings[num_pings].sequence =
                        relay_ping_history_ping_sent(relay->relay_manager->relay_ping_history[i], current_time);
                    pings[num_pings].address = relay->relay_manager->relay_addresses[i];
                    relay->relay_manager->relay_last_ping_time[i] = current_time;
                    num_pings++;
                }
            }

            relay_platform_mutex_release(relay->mutex);

            for (int i = 0; i < num_pings; ++i) {
                uint8_t packet_data[9];
                packet_data[0] = RELAY_PING_PACKET;
                uint8_t* p = packet_data + 1;
                encoding::write_uint64(&p, pings[i].sequence);
                relay_platform_socket_send_packet(relay->socket, &pings[i].address, packet_data, 9);
            }

            relay::relay_platform_sleep(1.0 / 100.0);
        }

        RELAY_PLATFORM_THREAD_RETURN();
    }
}  // namespace

int main(int argc, const char** argv)
{
    if (argc == 2 && strcmp(argv[1], "test") == 0) {
        testing::relay_test();
        return 0;
    }

    if (argc == 2 && strcmp(argv[1], "bench") == 0) {
        benchmarking::Benchmark::Run();
        return 0;
    }

    printf("\nNetwork Next Relay\n");

    printf("\nEnvironment:\n\n");

    const char* relay_address_env = relay::relay_platform_getenv("RELAY_ADDRESS");
    if (!relay_address_env) {
        printf("\nerror: RELAY_ADDRESS not set\n\n");
        return 1;
    }

    legacy::relay_address_t relay_address;
    if (relay_address_parse(&relay_address, relay_address_env) != RELAY_OK) {
        printf("\nerror: invalid relay address '%s'\n\n", relay_address_env);
        return 1;
    }

    {
        legacy::relay_address_t address_without_port = relay_address;
        address_without_port.port = 0;
        char address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
        printf("    relay address is '%s'\n", legacy::relay_address_to_string(&address_without_port, address_buffer));
    }

    uint16_t relay_bind_port = relay_address.port;

    printf("    relay bind port is %d\n", relay_bind_port);

    const char* relay_private_key_env = relay::relay_platform_getenv("RELAY_PRIVATE_KEY");
    if (!relay_private_key_env) {
        printf("\nerror: RELAY_PRIVATE_KEY not set\n\n");
        return 1;
    }

    uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];
    if (encoding::base64_decode_data(relay_private_key_env, relay_private_key, RELAY_PRIVATE_KEY_BYTES) !=
        RELAY_PRIVATE_KEY_BYTES) {
        printf("\nerror: invalid relay private key\n\n");
        return 1;
    }

    printf("    relay private key is '%s'\n", relay_private_key_env);

    const char* relay_public_key_env = relay::relay_platform_getenv("RELAY_PUBLIC_KEY");
    if (!relay_public_key_env) {
        printf("\nerror: RELAY_PUBLIC_KEY not set\n\n");
        return 1;
    }

    uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
    if (encoding::base64_decode_data(relay_public_key_env, relay_public_key, RELAY_PUBLIC_KEY_BYTES) !=
        RELAY_PUBLIC_KEY_BYTES) {
        printf("\nerror: invalid relay public key\n\n");
        return 1;
    }

    printf("    relay public key is '%s'\n", relay_public_key_env);

    const char* router_public_key_env = relay::relay_platform_getenv("RELAY_ROUTER_PUBLIC_KEY");
    if (!router_public_key_env) {
        printf("\nerror: RELAY_ROUTER_PUBLIC_KEY not set\n\n");
        return 1;
    }

    uint8_t router_public_key[crypto_sign_PUBLICKEYBYTES];
    if (encoding::base64_decode_data(router_public_key_env, router_public_key, crypto_sign_PUBLICKEYBYTES) !=
        crypto_sign_PUBLICKEYBYTES) {
        printf("\nerror: invalid router public key\n\n");
        return 1;
    }

    printf("    router public key is '%s'\n", router_public_key_env);

    const char* backend_hostname = relay::relay_platform_getenv("RELAY_BACKEND_HOSTNAME");
    if (!backend_hostname) {
        printf("\nerror: RELAY_BACKEND_HOSTNAME not set\n\n");
        return 1;
    }

    printf("    backend hostname is '%s'\n", backend_hostname);

    if (relay::relay_initialize() != RELAY_OK) {
        printf("\nerror: failed to initialize relay\n\n");
        return 1;
    }

    relay::relay_platform_socket_t* socket =
        relay::relay_platform_socket_create(&relay_address, RELAY_PLATFORM_SOCKET_BLOCKING, 0.1f, 100 * 1024, 100 * 1024);
    if (socket == NULL) {
        printf("\ncould not create socket\n\n");
        relay::relay_term();
        return 1;
    }

    printf("\nRelay socket opened on port %d\n", relay_address.port);
    char relay_address_buffer[RELAY_MAX_ADDRESS_STRING_LENGTH];
    const char* relay_address_string = relay_address_to_string(&relay_address, relay_address_buffer);

    CURL* curl = curl_easy_init();
    if (!curl) {
        printf("\nerror: could not initialize curl\n\n");
        relay_platform_socket_destroy(socket);
        curl_easy_cleanup(curl);
        relay::relay_term();
        return 1;
    }

    uint8_t relay_token[RELAY_TOKEN_BYTES];

    printf("\nInitializing relay\n");

    bool relay_initialized = false;

    uint64_t router_timestamp = 0;

    for (int i = 0; i < 60; ++i) {
        if (relay::relay_init(curl,
                backend_hostname,
                relay_token,
                relay_address_string,
                router_public_key,
                relay_private_key,
                &router_timestamp) == RELAY_OK) {
            printf("\n");
            relay_initialized = true;
            break;
        }

        printf(".");
        fflush(stdout);

        relay::relay_platform_sleep(1.0);
    }

    if (!relay_initialized) {
        printf("\nerror: could not initialize relay\n\n");
        relay_platform_socket_destroy(socket);
        curl_easy_cleanup(curl);
        relay::relay_term();
        return 1;
    }

    relay::relay_t relay;
    memset(&relay, 0, sizeof(relay));
    relay.initialize_time = relay::relay_platform_time();
    relay.initialize_router_timestamp = router_timestamp;
    relay.sessions = new std::map<uint64_t, relay::relay_session_t*>();
    memcpy(relay.relay_public_key, relay_public_key, RELAY_PUBLIC_KEY_BYTES);
    memcpy(relay.relay_private_key, relay_private_key, RELAY_PRIVATE_KEY_BYTES);
    memcpy(relay.router_public_key, router_public_key, crypto_sign_PUBLICKEYBYTES);

    relay.socket = socket;
    relay.mutex = relay::relay_platform_mutex_create();
    if (!relay.mutex) {
        printf("\nerror: could not create ping thread\n\n");
        quit = 1;
    }

    relay.relay_manager = relay::relay_manager_create();
    if (!relay.relay_manager) {
        printf("\nerror: could not create relay manager\n\n");
        quit = 1;
    }

    relay::relay_platform_thread_t* receive_thread = relay_platform_thread_create(receive_thread_function, &relay);
    if (!receive_thread) {
        printf("\nerror: could not create receive thread\n\n");
        quit = 1;
    }

    relay::relay_platform_thread_t* ping_thread = relay_platform_thread_create(ping_thread_function, &relay);
    if (!ping_thread) {
        printf("\nerror: could not create ping thread\n\n");
        quit = 1;
    }

    printf("Relay initialized\n\n");

    signal(SIGINT, interrupt_handler);

    uint8_t* update_response_memory = (uint8_t*)malloc(RESPONSE_MAX_BYTES);

    while (!quit) {
        bool updated = false;

        for (int i = 0; i < 10; ++i) {
            if (relay_update(curl, backend_hostname, relay_token, relay_address_string, update_response_memory, &relay) ==
                RELAY_OK) {
                updated = true;
                break;
            }
        }

        if (!updated) {
            printf("error: could not update relay\n\n");
            quit = 1;
            break;
        }

        relay::relay_platform_sleep(1.0);
    }

    printf("Cleaning up\n");

    if (receive_thread) {
        relay_platform_thread_join(receive_thread);
        relay_platform_thread_destroy(receive_thread);
    }

    if (ping_thread) {
        relay_platform_thread_join(ping_thread);
        relay_platform_thread_destroy(ping_thread);
    }

    free(update_response_memory);

    for (std::map<uint64_t, relay::relay_session_t*>::iterator itor = relay.sessions->begin(); itor != relay.sessions->end();
         ++itor) {
        delete itor->second;
    }

    delete relay.sessions;

    relay_manager_destroy(relay.relay_manager);

    relay_platform_mutex_destroy(relay.mutex);

    relay_platform_socket_destroy(socket);

    curl_easy_cleanup(curl);

    relay::relay_term();

    return 0;
}
