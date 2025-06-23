/*
    Network Next XDP Relay
*/

#ifndef RELAY_BPF_H
#define RELAY_BPF_H

#include "relay.h"

#include <bpf/bpf.h>
#include <bpf/libbpf.h>
#include <xdp/libxdp.h>

struct bpf_t
{
#ifdef COMPILE_WITH_BPF
    int interface_index;
    struct xdp_program * program;
    bool attached_native;
    bool attached_skb;
    int config_fd;
    int state_fd;
    int stats_fd;
    int relay_map_fd;
    int session_map_fd;
    int whitelist_map_fd;
#endif // #ifdef COMPILE_WITH_BPF
};

int bpf_init( struct bpf_t * bpf, uint32_t relay_public_address, uint32_t relay_internal_address );

void bpf_shutdown( struct bpf_t * bpf );

#endif // #ifndef RELAY_BPF_H
