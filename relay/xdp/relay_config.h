#ifndef RELAY_CONFIG_H
#define RELAY_CONFIG_H

#include "relay.h"

struct config_t
{
    char relay_name[256];
    uint16_t relay_port;
    uint32_t relay_public_address;
    uint32_t relay_internal_address;
    uint8_t relay_public_key[RELAY_PUBLIC_KEY_BYTES];
    uint8_t relay_private_key[RELAY_PRIVATE_KEY_BYTES];
    uint8_t relay_backend_public_key[RELAY_BACKEND_PUBLIC_KEY_BYTES];
    uint8_t relay_secret_key[RELAY_SECRET_KEY_BYTES];
    uint8_t gateway_ethernet_address[RELAY_ETHERNET_ADDRESS_BYTES];
    uint8_t use_gateway_ethernet_address;
    char relay_backend_url[256];
};

int read_config( struct config_t * config );

#endif // #ifndef RELAY_CONFIG_H
