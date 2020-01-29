#ifndef RELAY_RELAY_HPP
#define RELAY_RELAY_HPP

#include <sodium.h>
#include <curl/curl.h>

#include <cinttypes>
#include <map>  // TODO replace with unordered_map, map is not constant lookup

#include "relay_address.hpp"
#include "relay_manager.hpp"
#include "relay_replay_protection.hpp"
#include "relay_platform_socket.hpp"
#include "relay_platform.hpp"

namespace relay
{
    struct relay_session_t
    {
        uint64_t expire_timestamp;
        uint64_t session_id;
        uint8_t session_version;
        uint64_t client_to_server_sequence;
        uint64_t server_to_client_sequence;
        int kbps_up;
        int kbps_down;
        relay_address_t prev_address;
        relay_address_t next_address;
        uint8_t private_key[crypto_box_SECRETKEYBYTES];
        relay_replay_protection_t replay_protection_server_to_client;
        relay_replay_protection_t replay_protection_client_to_server;
    };

    struct relay_t
    {
        relay_manager_t* relay_manager;
        relay_platform_socket_t* socket;
        relay_platform_mutex_t* mutex;
        double initialize_time;
        uint64_t initialize_router_timestamp;
        uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
        uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];

        uint8_t router_public_key[RELAY_PUBLIC_KEY_BYTES];
        std::map<uint64_t, relay_session_t*>* sessions;
        bool relays_dirty;
        int num_relays;
        uint64_t relay_ids[MAX_RELAYS];
        relay::relay_address_t relay_addresses[MAX_RELAYS];
    };

    int relay_initialize();

    void relay_term(

    );

    int relay_init(CURL* curl,
        const char* hostname,
        uint8_t* relay_token,
        const char* relay_address,
        const uint8_t* router_public_key,
        const uint8_t* relay_private_key,
        uint64_t* router_timestamp);

    int relay_update(CURL* curl,
        const char* hostname,
        const uint8_t* relay_token,
        const char* relay_address,
        uint8_t* update_response_memory,
        relay_t* relay);

    int relay_write_header(int direction,
        uint8_t type,
        uint64_t sequence,
        uint64_t session_id,
        uint8_t session_version,
        const uint8_t* private_key,
        uint8_t* buffer,
        int buffer_length);

    int relay_peek_header(int direction,
        uint8_t* type,
        uint64_t* sequence,
        uint64_t* session_id,
        uint8_t* session_version,
        const uint8_t* buffer,
        int buffer_length);

    int relay_verify_header(int direction, const uint8_t* private_key, uint8_t* buffer, int buffer_length);

    uint64_t relay_timestamp(relay_t* relay);

    uint64_t relay_clean_sequence(uint64_t sequence);
}  // namespace relay
#endif