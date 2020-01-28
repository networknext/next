/*
    Network Next Relay.
    Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

#include <stdint.h>

#ifndef RELAY_H
#define RELAY_H

#define RELAY_OK 0
#define RELAY_ERROR -1

#define RELAY_ADDRESS_NONE 0
#define RELAY_ADDRESS_IPV4 1
#define RELAY_ADDRESS_IPV6 2

#define RELAY_MAX_ADDRESS_STRING_LENGTH 256

#if RELAY_PLATFORM == RELAY_PLATFORM_WINDOWS
#include "relay_windows.h"
#elif RELAY_PLATFORM == RELAY_PLATFORM_MAC
#include "relay_mac.h"
#elif RELAY_PLATFORM == RELAY_PLATFORM_LINUX
#include "relay_linux.h"
#endif

struct relay_address_t
{
    union
    {
        uint8_t ipv4[4];
        uint16_t ipv6[8];
    } data;
    uint16_t port;
    uint8_t type;
};

void relay_printf(const char* format, ...);

int relay_address_parse(relay_address_t* address, const char* address_string_in);

const char* relay_address_to_string(const relay_address_t* address, char* buffer);

int relay_address_equal(const relay_address_t* a, const relay_address_t* b);
#endif  // #ifndef RELAY_H
