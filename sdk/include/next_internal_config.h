/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_INTERNAL_CONFIG_H
#define NEXT_INTERNAL_CONFIG_H

#include "next_crypto.h"

struct next_internal_config_t
{
    char server_backend_hostname[256];
    uint64_t client_buyer_id;
    uint64_t server_buyer_id;
    uint8_t buyer_public_key[NEXT_CRYPTO_SIGN_PUBLICKEYBYTES];
    uint8_t buyer_private_key[NEXT_CRYPTO_SIGN_SECRETKEYBYTES];
    bool valid_buyer_private_key;
    bool valid_buyer_public_key;
    int socket_send_buffer_size;
    int socket_receive_buffer_size;
    bool disable_network_next;
    bool disable_autodetect;
};

#endif // #ifndef NEXT_H
