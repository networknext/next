/*
    Network Next Relay XDP program

    USAGE:

        clang -Ilibbpf/src -g -O2 -target bpf -c relay_xdp.c -o relay_xdp.o
        sudo ip link set dev enp4s0 xdp obj relay_xdp.o sec relay_xdp
        sudo cat /sys/kernel/debug/tracing/trace_pipe
        sudo ip link set dev enp4s0 xdp off
*/

#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/if_vlan.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/udp.h>
#include <linux/bpf.h>
#include <linux/string.h>
#include <bpf/bpf_helpers.h>

#include "relay_shared.h"

#if defined(__BYTE_ORDER__) && defined(__ORDER_LITTLE_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_ntohs(x)        __builtin_bswap16(x)
#define bpf_htons(x)        __builtin_bswap16(x)
#elif defined(__BYTE_ORDER__) && defined(__ORDER_BIG_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define bpf_ntohs(x)        (x)
#define bpf_htons(x)        (x)
#else
# error "Endianness detection needs to be set up for your compiler?!"
#endif

struct {
    __uint( type, BPF_MAP_TYPE_ARRAY );
    __type( key, __u32 );
    __type( value, struct relay_config );
    __uint( max_entries, 1 );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} config_map SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_ARRAY );
    __type( key, __u32 );
    __type( value, struct relay_state );
    __uint( max_entries, 1 );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} state_map SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_PERCPU_ARRAY );
    __type( key, __u32 );
    __type( value, struct relay_stats );
    __uint( max_entries, 1 );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} stats_map SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_LRU_HASH );
    __type( key, __u64 );
    __type( value, __u64 );
    __uint( max_entries, MAX_RELAYS * 2 );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} relay_map SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_LRU_HASH );
    __type( key, struct session_key );
    __type( value, struct session_data );
    __uint( max_entries, MAX_SESSIONS * 2 );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} session_map SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_LRU_HASH );
    __type( key, struct whitelist_key );
    __type( value, struct whitelist_value );
    __uint( max_entries, MAX_SESSIONS * 2 );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
} whitelist_map SEC(".maps");

#define INCREMENT_COUNTER(counter_index) __sync_fetch_and_add( &stats->counters[counter_index], 1 )

#define DECREMENT_COUNTER(counter_index) __sync_fetch_and_sub( &stats->counters[counter_index], 1 )

#define ADD_COUNTER(counter_index, value) __sync_fetch_and_add( &stats->counters[counter_index], ( value) )

#define SUB_COUNTER(counter_index, value) __sync_fetch_and_sub( &stats->counters[counter_index], ( value) )

#define XCHACHA20POLY1305_NONCE_SIZE 24

#define CHACHA20POLY1305_KEY_SIZE 32

struct chacha20poly1305_crypto
{
    __u8 nonce[XCHACHA20POLY1305_NONCE_SIZE];
    __u8 key[CHACHA20POLY1305_KEY_SIZE];
};

int bpf_relay_sha256( void * data, int data__sz, void * output, int output__sz ) __ksym;

int bpf_relay_xchacha20poly1305_decrypt( void * data, int data__sz, struct chacha20poly1305_crypto * crypto ) __ksym;

#ifndef RELAY_DEBUG
#define RELAY_DEBUG 0
#endif // #ifndef RELAY_DEBUG

#if RELAY_DEBUG
#define relay_printf bpf_printk
#else // #if RELAY_DEBUG
#define relay_printf(...) do { } while (0)
#endif // #if RELAY_DEBUG

static int relay_decrypt_route_token( struct decrypt_route_token_data * data, void * route_token, int route_token__sz )
{
    __u8 * nonce = route_token;
    __u8 * encrypted = route_token + 24;
    struct chacha20poly1305_crypto crypto_data;
    memcpy( crypto_data.nonce, nonce, XCHACHA20POLY1305_NONCE_SIZE );
    memcpy( crypto_data.key, data->relay_secret_key, CHACHA20POLY1305_KEY_SIZE );
    if ( !bpf_relay_xchacha20poly1305_decrypt( encrypted, RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES - 24, &crypto_data ) )
        return 0;
    return 1;
}

static int relay_decrypt_continue_token( struct decrypt_continue_token_data * data, void * continue_token, int continue_token__sz )
{
    __u8 * nonce = continue_token;
    __u8 * encrypted = continue_token + 24;
    struct chacha20poly1305_crypto crypto_data;
    memcpy( crypto_data.nonce, nonce, XCHACHA20POLY1305_NONCE_SIZE );
    memcpy( crypto_data.key, data->relay_secret_key, CHACHA20POLY1305_KEY_SIZE );
    if ( !bpf_relay_xchacha20poly1305_decrypt( encrypted, RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES - 24, &crypto_data ) )
        return 0;
    return 1;
}

static void relay_reflect_packet( void * data, int payload_bytes, __u8 * magic )
{
    struct ethhdr * eth = data;
    struct iphdr  * ip  = data + sizeof( struct ethhdr );
    struct udphdr * udp = (void*) ip + sizeof( struct iphdr );

    __u16 a = udp->source;
    udp->source = udp->dest;
    udp->dest = a;
    udp->check = 0;
    udp->len = bpf_htons( sizeof(struct udphdr) + payload_bytes );

    __u32 b = ip->saddr;
    ip->saddr = ip->daddr;
    ip->daddr = b;
    ip->tot_len = bpf_htons( sizeof(struct iphdr) + sizeof(struct udphdr) + payload_bytes );
    ip->check = 0;

    char c[ETH_ALEN];
    memcpy( c, eth->h_source, ETH_ALEN );
    memcpy( eth->h_source, eth->h_dest, ETH_ALEN );
    memcpy( eth->h_dest, c, ETH_ALEN );

    __u16 * p = (__u16*) ip;
    __u32 checksum = p[0];
    checksum += p[1];
    checksum += p[2];
    checksum += p[3];
    checksum += p[4];
    checksum += p[5];
    checksum += p[6];
    checksum += p[7];
    checksum += p[8];
    checksum += p[9];
    checksum = ~ ( ( checksum & 0xFFFF ) + ( checksum >> 16 ) );
    ip->check = checksum;

    __u8 * packet_data = (void*) udp + sizeof( struct udphdr );

    __u32 from = ip->saddr;
    __u32 to   = ip->daddr;

    unsigned short sum = 0;

    sum += ( from >> 24 );
    sum += ( from >> 16 ) & 0xFF;
    sum += ( from >> 8  ) & 0xFF;
    sum += ( from       ) & 0xFF;

    sum += ( to >> 24 );
    sum += ( to >> 16 ) & 0xFF;
    sum += ( to >> 8  ) & 0xFF;
    sum += ( to       ) & 0xFF;

    sum += ( payload_bytes >> 8 );
    sum += ( payload_bytes      ) & 0xFF;

    char * sum_data = (char*) &sum;

    __u8 sum_0 = ( sum      ) & 0xFF;
    __u8 sum_1 = ( sum >> 8 );

    __u8 pittle[2];
    pittle[0] = 1 | ( sum_0 ^ sum_1 ^ 193 );
    pittle[1] = 1 | ( ( 255 - pittle[0] ) ^ 113 );

    packet_data[1] = pittle[0];
    packet_data[2] = pittle[1];

    __u64 hash = 0xCBF29CE484222325;

    hash ^= magic[0];
    hash *= 0x00000100000001B3;

    hash ^= magic[1];
    hash *= 0x00000100000001B3;

    hash ^= magic[2];
    hash *= 0x00000100000001B3;

    hash ^= magic[3];
    hash *= 0x00000100000001B3;

    hash ^= magic[4];
    hash *= 0x00000100000001B3;

    hash ^= magic[5];
    hash *= 0x00000100000001B3;

    hash ^= magic[6];
    hash *= 0x00000100000001B3;

    hash ^= magic[7];
    hash *= 0x00000100000001B3;

    hash ^= ( from       ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( from >> 8  ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( from >> 16 ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( from >> 24 );
    hash *= 0x00000100000001B3;

    hash ^= ( to       ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( to >> 8  ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( to >> 16 ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( to >> 24 );
    hash *= 0x00000100000001B3;

    hash ^= ( payload_bytes      ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( payload_bytes >> 8 );
    hash *= 0x00000100000001B3;

    __u8 hash_0 = ( hash       ) & 0xFF;
    __u8 hash_1 = ( hash >> 8  ) & 0xFF;
    __u8 hash_2 = ( hash >> 16 ) & 0xFF;
    __u8 hash_3 = ( hash >> 24 ) & 0xFF;
    __u8 hash_4 = ( hash >> 32 ) & 0xFF;
    __u8 hash_5 = ( hash >> 40 ) & 0xFF;
    __u8 hash_6 = ( hash >> 48 ) & 0xFF;
    __u8 hash_7 = ( hash >> 56 );

    __u8 chonkle[15];

    chonkle[0] = ( ( hash_6 & 0xC0 ) >> 6 ) + 42;
    chonkle[1] = ( hash_3 & 0x1F ) + 200;
    chonkle[2] = ( ( hash_2 & 0xFC ) >> 2 ) + 5;
    chonkle[3] = hash_0;
    chonkle[4] = ( hash_2 & 0x03 ) + 78;
    chonkle[5] = ( hash_4 & 0x7F ) + 96;
    chonkle[6] = ( ( hash_1 & 0xFC ) >> 2 ) + 100;
    if ( ( hash_7 & 1 ) == 0 ) 
    {
        chonkle[7] = 79;
    } 
    else 
    {
        chonkle[7] = 7;
    }
    if ( ( hash_4 & 0x80 ) == 0 )
    {
        chonkle[8] = 37;
    } 
    else 
    {
        chonkle[8] = 83;
    }
    chonkle[9] = ( hash_5 & 0x07 ) + 124;
    chonkle[10] = ( ( hash_1 & 0xE0 ) >> 5 ) + 175;
    chonkle[11] = ( hash_6 & 0x3F ) + 33;
    __u8 value = ( hash_1 & 0x03 );
    if ( value == 0 )
    {
        chonkle[12] = 97;
    } 
    else if ( value == 1 )
    {
        chonkle[12] = 5;
    } 
    else if ( value == 2 )
    {
        chonkle[12] = 43;
    } 
    else 
    {
        chonkle[12] = 13;
    }
    chonkle[13] = ( ( hash_5 & 0xF8 ) >> 3 ) + 210;
    chonkle[14] = ( ( hash_7 & 0xFE ) >> 1 ) + 17;

    packet_data[3]  = chonkle[0];
    packet_data[4]  = chonkle[1];
    packet_data[5]  = chonkle[2];
    packet_data[6]  = chonkle[3];
    packet_data[7]  = chonkle[4];
    packet_data[8]  = chonkle[5];
    packet_data[9] = chonkle[6];
    packet_data[10] = chonkle[7];
    packet_data[11] = chonkle[8];
    packet_data[12] = chonkle[9];
    packet_data[13] = chonkle[10];
    packet_data[14] = chonkle[11];
    packet_data[15] = chonkle[12];
    packet_data[16] = chonkle[13];
    packet_data[17] = chonkle[14];
}

static int relay_redirect_packet( void * data, int payload_bytes, __u32 dest_address, __u16 dest_port, __u8 * magic )
{
    struct ethhdr * eth = data;
    struct iphdr  * ip  = data + sizeof( struct ethhdr );
    struct udphdr * udp = (void*) ip + sizeof( struct iphdr );

    udp->source = udp->dest;
    udp->dest = dest_port;
    udp->check = 0;
    udp->len = bpf_htons( sizeof(struct udphdr) + payload_bytes );

    ip->saddr = ip->daddr;
    ip->daddr = dest_address;
    ip->tot_len = bpf_htons( sizeof(struct iphdr) + sizeof(struct udphdr) + payload_bytes );
    ip->check = 0;

    struct whitelist_key key;
    key.address = dest_address;
    key.port = dest_port;
    
    struct whitelist_value * whitelist_value = (struct whitelist_value*) bpf_map_lookup_elem( &whitelist_map, &key );
    if ( whitelist_value == NULL )
    {
        relay_printf( "redirect address not in whitelist" );
        return XDP_DROP;
    }

    memcpy( eth->h_source, whitelist_value->dest_address, 6 );
    memcpy( eth->h_dest, whitelist_value->source_address, 6 );

    __u16 * p = (__u16*) ip;
    __u32 checksum = p[0];
    checksum += p[1];
    checksum += p[2];
    checksum += p[3];
    checksum += p[4];
    checksum += p[5];
    checksum += p[6];
    checksum += p[7];
    checksum += p[8];
    checksum += p[9];
    checksum = ~ ( ( checksum & 0xFFFF ) + ( checksum >> 16 ) );
    ip->check = checksum;

    __u8 * packet_data = (void*) udp + sizeof( struct udphdr );

    __u32 from = ip->saddr;
    __u32 to   = ip->daddr;

    unsigned short sum = 0;

    sum += ( from >> 24 );
    sum += ( from >> 16 ) & 0xFF;
    sum += ( from >> 8  ) & 0xFF;
    sum += ( from       ) & 0xFF;

    sum += ( to >> 24 );
    sum += ( to >> 16 ) & 0xFF;
    sum += ( to >> 8  ) & 0xFF;
    sum += ( to       ) & 0xFF;

    sum += ( payload_bytes >> 8 );
    sum += ( payload_bytes      ) & 0xFF;

    char * sum_data = (char*) &sum;

    __u8 sum_0 = ( sum      ) & 0xFF;
    __u8 sum_1 = ( sum >> 8 );

    __u8 pittle[2];
    pittle[0] = 1 | ( sum_0 ^ sum_1 ^ 193 );
    pittle[1] = 1 | ( ( 255 - pittle[0] ) ^ 113 );

    packet_data[1] = pittle[0];
    packet_data[2] = pittle[1];

    __u64 hash = 0xCBF29CE484222325;

    hash ^= magic[0];
    hash *= 0x00000100000001B3;

    hash ^= magic[1];
    hash *= 0x00000100000001B3;

    hash ^= magic[2];
    hash *= 0x00000100000001B3;

    hash ^= magic[3];
    hash *= 0x00000100000001B3;

    hash ^= magic[4];
    hash *= 0x00000100000001B3;

    hash ^= magic[5];
    hash *= 0x00000100000001B3;

    hash ^= magic[6];
    hash *= 0x00000100000001B3;

    hash ^= magic[7];
    hash *= 0x00000100000001B3;

    hash ^= ( from       ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( from >> 8  ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( from >> 16 ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( from >> 24 );
    hash *= 0x00000100000001B3;

    hash ^= ( to       ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( to >> 8  ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( to >> 16 ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( to >> 24 );
    hash *= 0x00000100000001B3;

    hash ^= ( payload_bytes      ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( payload_bytes >> 8 );
    hash *= 0x00000100000001B3;

    __u8 hash_0 = ( hash       ) & 0xFF;
    __u8 hash_1 = ( hash >> 8  ) & 0xFF;
    __u8 hash_2 = ( hash >> 16 ) & 0xFF;
    __u8 hash_3 = ( hash >> 24 ) & 0xFF;
    __u8 hash_4 = ( hash >> 32 ) & 0xFF;
    __u8 hash_5 = ( hash >> 40 ) & 0xFF;
    __u8 hash_6 = ( hash >> 48 ) & 0xFF;
    __u8 hash_7 = ( hash >> 56 );

    __u8 chonkle[15];

    chonkle[0] = ( ( hash_6 & 0xC0 ) >> 6 ) + 42;
    chonkle[1] = ( hash_3 & 0x1F ) + 200;
    chonkle[2] = ( ( hash_2 & 0xFC ) >> 2 ) + 5;
    chonkle[3] = hash_0;
    chonkle[4] = ( hash_2 & 0x03 ) + 78;
    chonkle[5] = ( hash_4 & 0x7F ) + 96;
    chonkle[6] = ( ( hash_1 & 0xFC ) >> 2 ) + 100;
    if ( ( hash_7 & 1 ) == 0 ) 
    {
        chonkle[7] = 79;
    } 
    else 
    {
        chonkle[7] = 7;
    }
    if ( ( hash_4 & 0x80 ) == 0 )
    {
        chonkle[8] = 37;
    } 
    else 
    {
        chonkle[8] = 83;
    }
    chonkle[9] = ( hash_5 & 0x07 ) + 124;
    chonkle[10] = ( ( hash_1 & 0xE0 ) >> 5 ) + 175;
    chonkle[11] = ( hash_6 & 0x3F ) + 33;
    __u8 value = ( hash_1 & 0x03 );
    if ( value == 0 )
    {
        chonkle[12] = 97;
    } 
    else if ( value == 1 )
    {
        chonkle[12] = 5;
    } 
    else if ( value == 2 )
    {
        chonkle[12] = 43;
    } 
    else 
    {
        chonkle[12] = 13;
    }
    chonkle[13] = ( ( hash_5 & 0xF8 ) >> 3 ) + 210;
    chonkle[14] = ( ( hash_7 & 0xFE ) >> 1 ) + 17;

    packet_data[3]  = chonkle[0];
    packet_data[4]  = chonkle[1];
    packet_data[5]  = chonkle[2];
    packet_data[6]  = chonkle[3];
    packet_data[7]  = chonkle[4];
    packet_data[8]  = chonkle[5];
    packet_data[9] = chonkle[6];
    packet_data[10] = chonkle[7];
    packet_data[11] = chonkle[8];
    packet_data[12] = chonkle[9];
    packet_data[13] = chonkle[10];
    packet_data[14] = chonkle[11];
    packet_data[15] = chonkle[12];
    packet_data[16] = chonkle[13];
    packet_data[17] = chonkle[14];

    return XDP_TX;
}

SEC("relay_xdp") int relay_xdp_filter( struct xdp_md *ctx ) 
{ 
    void * data = (void*) (long) ctx->data; 

    void * data_end = (void*) (long) ctx->data_end; 

    struct ethhdr * eth = data;

    int key = 0;
    struct relay_stats * stats = (struct relay_stats*) bpf_map_lookup_elem( &stats_map, &key );
    if ( stats == NULL )
        return XDP_PASS;

    struct relay_config * config = (struct relay_config*) bpf_map_lookup_elem( &config_map, &key );
    if ( config == NULL )
        return XDP_PASS;

    if ( (void*)eth + sizeof(struct ethhdr) <= data_end )
    {
        if ( eth->h_proto == __constant_htons(ETH_P_IP) ) // IPV4
        {
            struct iphdr * ip = data + sizeof(struct ethhdr);

            if ( (void*)ip + sizeof(struct iphdr) > data_end )
            {
                relay_printf( "smaller than ipv4 header" );
                INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                return XDP_DROP;
            }

            if ( ip->protocol == IPPROTO_UDP ) // UDP only
            {
                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_RECEIVED );
                ADD_COUNTER( RELAY_COUNTER_BYTES_RECEIVED, data_end - data );

                // Drop UDP packets with IPv4 headers not equal to 20 bytes

                if ( ip->ihl != 5 )
                {
                    relay_printf( "ip header is not 20 bytes" );
                    INCREMENT_COUNTER( RELAY_COUNTER_DROP_LARGE_IP_HEADER );
                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                    return config->dedicated ? XDP_DROP : XDP_PASS;
                }

                struct udphdr * udp = (void*) ip + sizeof(struct iphdr);

                if ( (void*)udp + sizeof(struct udphdr) <= data_end )
                {
                    if ( udp->dest == config->relay_port && ( ip->daddr == config->relay_public_address || ip->daddr == config->relay_internal_address ) )
                    {
                        struct relay_state * state;
                        __u8 * packet_data = (unsigned char*) (void*)udp + sizeof(struct udphdr);

                        // Drop packets that are too small to be valid

                        if ( (void*)packet_data + 18 > data_end )
                        {
                            relay_printf( "packet is too small" );
                            INCREMENT_COUNTER( RELAY_COUNTER_PACKET_TOO_SMALL );
                            INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                            ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                            return XDP_DROP;
                        }

                        // Drop packets that are too large to be valid

                        int packet_bytes = data_end - (void*)udp - sizeof(struct udphdr);

                        if ( packet_bytes > 1400 )
                        {
                            relay_printf( "packet is too large" );
                            INCREMENT_COUNTER( RELAY_COUNTER_PACKET_TOO_LARGE );
                            INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                            ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                            return XDP_DROP;
                        }

                        __u8 packet_type = packet_data[0];

                        switch ( packet_type )
                        {
                            case RELAY_PING_PACKET:
                            {
                                relay_printf( "relay ping packet" ); // from %x:%d", ip->saddr, udp->source );

                                INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED );

                                // IMPORTANT: for verifier
                                if ( (void*) packet_data + 18 + 8 + 8 + 1 + RELAY_PING_TOKEN_BYTES > data_end )
                                {
                                    relay_printf( "relay ping packet is wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u8 * payload = packet_data + 18;

                                __u64 expire_timestamp;
                                expire_timestamp  = payload[8];
                                expire_timestamp |= ( ( (__u64)( payload[8 + 1] ) ) << 8  );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 2] ) ) << 16 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 3] ) ) << 24 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 4] ) ) << 32 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 5] ) ) << 40 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 6] ) ) << 48 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 7] ) ) << 56 );

                                if ( expire_timestamp < state->current_timestamp )
                                {
                                    // relay_printf( "ping token expired: %lld < %lld", expire_timestamp, state->current_timestamp );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_EXPIRED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 relay_map_key = ( ( (__u64)ip->saddr ) << 32 ) | udp->source;
                                void * relay_map_value = bpf_map_lookup_elem( &relay_map, &relay_map_key );
                                if ( relay_map_value == NULL )
                                {
                                    // relay_printf( "unknown relay %x:%d", ip->saddr, bpf_ntohs( udp->source ) );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_UNKNOWN_RELAY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                /*
                                struct ping_token_data verify_data;
                                verify_data.source_address = ip->saddr;
                                verify_data.source_port = udp->source;
                                verify_data.dest_address = ip->daddr;
                                verify_data.dest_port = udp->dest;
                                verify_data.expire_timestamp = expire_timestamp;
                                memcpy( verify_data.ping_key, state->ping_key, RELAY_PING_KEY_BYTES );

                                __u8 hash[32];
                                bpf_relay_sha256( &verify_data, sizeof(struct ping_token_data), hash, 32 );
                                __u8 * ping_token = packet_data + 8 + 8 + 1;
                                if ( relay_memcmp( hash, ping_token, 32 ) != 0 )
                                {
                                    relay_printf( "did not verify: %x:%d -> %x:%d", ip->saddr, bpf_ntohs( udp->source ), ip->daddr, bpf_ntohs( udp->dest ) );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }
                                */

                                struct whitelist_key key;
                                key.address = ip->saddr;
                                key.port = udp->source;
                                
                                struct whitelist_value value;
                                value.expire_timestamp = state->current_timestamp + 30;
                                memcpy( value.source_address, eth->h_source, 6 );
                                memcpy( value.dest_address, eth->h_dest, 6 );

                                bpf_map_update_elem( &whitelist_map, &key, &value, BPF_ANY );

                                packet_data[0] = RELAY_PONG_PACKET;

                                const int payload_bytes = 18 + 8;

                                relay_reflect_packet( data, payload_bytes, state->current_magic );

                                bpf_xdp_adjust_tail( ctx, -( 8 + 1 + RELAY_PING_TOKEN_BYTES ) );

                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PONG_PACKET_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, payload_bytes + sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr) );
        
                                // relay_printf( "sent relay pong packet to %x:%d", ip->saddr, udp->source );

                                return XDP_TX;
                            }

                            default:
                                return XDP_DROP;
                        }
                    }
                    else
                    {
                        // drop UDP packets not sent to the relay address and port in dedicated mode

                        if ( config->dedicated )
                        {
                            INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                            ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                            return XDP_DROP;
                        }
                    }
                }
                else
                {
                    // drop non-UDP IPv4 packets in dedicated mode

                    if ( config->dedicated )
                    {
                        INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                        ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                        return XDP_DROP;
                    }
                }
            }
        }
        else if ( eth->h_proto == __constant_htons(ETH_P_IPV6) )
        {
            // drop IPv6 packets in dedicated mode

            if ( config->dedicated )
            {
                INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                return XDP_DROP;
            }
        }
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
