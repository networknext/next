/*
    Network Next XDP Relay
*/

#ifndef RELAY_SET_H
#define RELAY_SET_H

#include "relay.h"

struct relay_set
{
    int num_relays;
    uint64_t id[MAX_RELAYS];
    uint32_t address[MAX_RELAYS];
    uint16_t port[MAX_RELAYS];
    uint8_t internal[MAX_RELAYS];
};

#endif // #ifndef RELAY_SET_H
