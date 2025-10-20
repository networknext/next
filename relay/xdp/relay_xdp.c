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

#define RELAY_ADVANCED_PACKET_FILTER 0

#include "relay_shared.h"

#if defined(__BYTE_ORDER__) && defined(__ORDER_LITTLE_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_ntohl(x)        __builtin_bswap32(x)
#define bpf_htonl(x)        __builtin_bswap32(x)
#define bpf_ntohs(x)        __builtin_bswap16(x)
#define bpf_htons(x)        __builtin_bswap16(x)
#elif defined(__BYTE_ORDER__) && defined(__ORDER_BIG_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define bpf_ntohl(x)        (x)
#define bpf_htonl(x)        (x)
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

#define INCREMENT_COUNTER(counter_index)  __sync_fetch_and_add( &stats->counters[counter_index], 1 )

#define DECREMENT_COUNTER(counter_index)  __sync_fetch_and_sub( &stats->counters[counter_index], 1 )

#define ADD_COUNTER(counter_index, value) __sync_fetch_and_add( &stats->counters[counter_index], ( value) )

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

static void relay_reflect_packet( void * data, int payload_bytes, __u8 * magic, __u8 * gateway_ethernet_address )
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

    // IMPORTANT: Sometimes reflecting the ethernet address doesn't work
    // In this case let the config specify the gateway ethernet address
    if ( gateway_ethernet_address != NULL )
    {
        memcpy( eth->h_dest, gateway_ethernet_address, ETH_ALEN );
    }

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
    packet_data[9]  = chonkle[6];
    packet_data[10] = chonkle[7];
    packet_data[11] = chonkle[8];
    packet_data[12] = chonkle[9];
    packet_data[13] = chonkle[10];
    packet_data[14] = chonkle[11];
    packet_data[15] = chonkle[12];
    packet_data[16] = chonkle[13];
    packet_data[17] = chonkle[14];
}

struct redirect_args_t
{
    void * data;
    int payload_bytes;
    __u32 source_address;
    __u32 dest_address;
    __u16 source_port;
    __u16 dest_port;
    __u8 * magic;
    __u8 * gateway_ethernet_address;
};

static int relay_redirect_packet( struct redirect_args_t * args )
{
    struct ethhdr * eth = args->data;
    struct iphdr  * ip  = args->data + sizeof( struct ethhdr );
    struct udphdr * udp = (void*) ip + sizeof( struct iphdr );

    udp->source = args->source_port;
    udp->dest = args->dest_port;
    udp->check = 0;
    udp->len = bpf_htons( sizeof(struct udphdr) + args->payload_bytes );

    ip->saddr = args->source_address;
    ip->daddr = args->dest_address;
    ip->tot_len = bpf_htons( sizeof(struct iphdr) + sizeof(struct udphdr) + args->payload_bytes );
    ip->check = 0;

    struct whitelist_key key;
    key.address = args->dest_address;
    key.port = args->dest_port;
    
    struct whitelist_value * whitelist_value = (struct whitelist_value*) bpf_map_lookup_elem( &whitelist_map, &key );
    if ( whitelist_value == NULL )
    {
        return XDP_DROP;
    }

    memcpy( eth->h_source, whitelist_value->dest_address, 6 );
    memcpy( eth->h_dest, whitelist_value->source_address, 6 );

    // IMPORTANT: Sometimes reflecting the ethernet address doesn't work
    // In this case let the config specify the gateway ethernet address
    if ( args->gateway_ethernet_address != NULL )
    {
        memcpy( eth->h_dest, args->gateway_ethernet_address, ETH_ALEN );
    }

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

    sum += ( args->payload_bytes >> 8 );
    sum += ( args->payload_bytes      ) & 0xFF;

    char * sum_data = (char*) &sum;

    __u8 sum_0 = ( sum      ) & 0xFF;
    __u8 sum_1 = ( sum >> 8 );

    __u8 pittle[2];
    pittle[0] = 1 | ( sum_0 ^ sum_1 ^ 193 );
    pittle[1] = 1 | ( ( 255 - pittle[0] ) ^ 113 );

    packet_data[1] = pittle[0];
    packet_data[2] = pittle[1];

    __u64 hash = 0xCBF29CE484222325;

    hash ^= args->magic[0];
    hash *= 0x00000100000001B3;

    hash ^= args->magic[1];
    hash *= 0x00000100000001B3;

    hash ^= args->magic[2];
    hash *= 0x00000100000001B3;

    hash ^= args->magic[3];
    hash *= 0x00000100000001B3;

    hash ^= args->magic[4];
    hash *= 0x00000100000001B3;

    hash ^= args->magic[5];
    hash *= 0x00000100000001B3;

    hash ^= args->magic[6];
    hash *= 0x00000100000001B3;

    hash ^= args->magic[7];
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

    hash ^= ( args->payload_bytes      ) & 0xFF;
    hash *= 0x00000100000001B3;

    hash ^= ( args->payload_bytes >> 8 );
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

                        // Basic packet filter

                        if ( packet_data[0] < 0x01 || packet_data[0] > 0x0E                                                           ||
                             packet_data[2] != ( 1 | ( ( 255 - packet_data[1] ) ^ 113 ) )                                             ||
                             packet_data[3] < 0x2A || packet_data[3] > 0x2D                                                           ||
                             packet_data[4] < 0xC8 || packet_data[4] > 0xE7                                                           ||
                             packet_data[5] < 0x05 || packet_data[5] > 0x44                                                           ||
                             packet_data[7] < 0x4E || packet_data[7] > 0x51                                                           ||
                             packet_data[8] < 0x60 || packet_data[8] > 0xDF                                                           ||
                             packet_data[9] < 0x64 || packet_data[9] > 0xE3                                                           ||
                             packet_data[10] != 0x07 && packet_data[10] != 0x4F                                                       ||
                             packet_data[11] != 0x25 && packet_data[11] != 0x53                                                       ||
                             packet_data[12] < 0x7C || packet_data[12] > 0x83                                                         ||
                             packet_data[13] < 0xAF || packet_data[13] > 0xB6                                                         ||
                             packet_data[14] < 0x21 || packet_data[14] > 0x60                                                         ||
                             packet_data[15] != 0x61 && packet_data[15] != 0x05 && packet_data[15] != 0x2B && packet_data[15] != 0x0D ||
                             packet_data[16] < 0xD2 || packet_data[16] > 0xF1                                                         ||
                             packet_data[17] < 0x11 || packet_data[17] > 0x90 )
                        {
                            relay_printf( "basic packet filter dropped packet [%d]", packet_data[0] );
                            INCREMENT_COUNTER( RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET );
                            INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                            ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                            return XDP_DROP;
                        }

                        // Get relay state

                        state = (struct relay_state*) bpf_map_lookup_elem( &state_map, &key );
                        if ( state == NULL )
                        {
                            relay_printf( "null relay state" );
                            INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                            ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                            return XDP_DROP;
                        }

#if RELAY_ADVANCED_PACKET_FILTER

                        // Advanced packet filter

                        __u32 from = ip->saddr;
                        __u32 to   = config->relay_public_address;

                        unsigned short sum = 0;

                        sum += ( from       ) & 0xFF;
                        sum += ( from >> 8  ) & 0xFF;
                        sum += ( from >> 16 ) & 0xFF;
                        sum += ( from >> 24 );

                        sum += ( to       ) & 0xFF;
                        sum += ( to >> 8  ) & 0xFF;
                        sum += ( to >> 16 ) & 0xFF;
                        sum += ( to >> 24 );

                        sum += ( packet_bytes >> 8 );
                        sum += ( packet_bytes      ) & 0xFF;

                        char * sum_data = (char*) &sum;

                        __u8 sum_0 = ( sum      ) & 0xFF;
                        __u8 sum_1 = ( sum >> 8 );

                        __u8 pittle[2];
                        pittle[0] = 1 | ( sum_0 ^ sum_1 ^ 193 );
                        pittle[1] = 1 | ( ( 255 - pittle[0] ) ^ 113 );

                        if ( pittle[0] != packet_data[1] || pittle[1] != packet_data[2] )
                        {
                            to = config->relay_internal_address;

                            unsigned short sum = 0;

                            sum += ( from       ) & 0xFF;
                            sum += ( from >> 8  ) & 0xFF;
                            sum += ( from >> 16 ) & 0xFF;
                            sum += ( from >> 24 );

                            sum += ( to       ) & 0xFF;
                            sum += ( to >> 8  ) & 0xFF;
                            sum += ( to >> 16 ) & 0xFF;
                            sum += ( to >> 24 );

                            sum += ( packet_bytes >> 8 );
                            sum += ( packet_bytes      ) & 0xFF;

                            char * sum_data = (char*) &sum;

                            __u8 sum_0 = ( sum      ) & 0xFF;
                            __u8 sum_1 = ( sum >> 8 );

                            __u8 pittle[2];
                            pittle[0] = 1 | ( sum_0 ^ sum_1 ^ 193 );
                            pittle[1] = 1 | ( ( 255 - pittle[0] ) ^ 113 );

                            if ( pittle[0] != packet_data[1] || pittle[1] != packet_data[2] )
                            {
                                relay_printf( "advanced packet filter dropped packet (a) [%d]", packet_data[0] );
                                INCREMENT_COUNTER( RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET );
                                INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                return XDP_DROP;
                            }
                        }

                        // current magic

                        int passed = 0;
                        {
                            __u8 * magic = state->current_magic;

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

                            hash ^= ( packet_bytes      ) & 0xFF;
                            hash *= 0x00000100000001B3;

                            hash ^= ( packet_bytes >> 8 );
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

                            if ( chonkle[0] == packet_data[3]   &&
                                 chonkle[1] == packet_data[4]   &&
                                 chonkle[2] == packet_data[5]   &&
                                 chonkle[3] == packet_data[6]   &&
                                 chonkle[4] == packet_data[7]   &&
                                 chonkle[5] == packet_data[8]   &&
                                 chonkle[6] == packet_data[9]   &&
                                 chonkle[7] == packet_data[10]  &&
                                 chonkle[8] == packet_data[11]  &&
                                 chonkle[9] == packet_data[12]  &&
                                 chonkle[10] == packet_data[13] &&
                                 chonkle[11] == packet_data[14] &&
                                 chonkle[12] == packet_data[15] &&
                                 chonkle[13] == packet_data[16] &&
                                 chonkle[14] == packet_data[17] )
                            {
                                passed = 1;
                            }
                        }

                        if ( !passed )
                        {
                            // previous magic

                            __u8 * magic = state->previous_magic;

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

                            hash ^= ( packet_bytes      ) & 0xFF;
                            hash *= 0x00000100000001B3;

                            hash ^= ( packet_bytes >> 8 );
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

                            if ( chonkle[0] == packet_data[3]   &&
                                 chonkle[1] == packet_data[4]   &&
                                 chonkle[2] == packet_data[5]   &&
                                 chonkle[3] == packet_data[6]   &&
                                 chonkle[4] == packet_data[7]   &&
                                 chonkle[5] == packet_data[8]   &&
                                 chonkle[6] == packet_data[9]   &&
                                 chonkle[7] == packet_data[10]  &&
                                 chonkle[8] == packet_data[11]  &&
                                 chonkle[9] == packet_data[12]  &&
                                 chonkle[10] == packet_data[13] &&
                                 chonkle[11] == packet_data[14] &&
                                 chonkle[12] == packet_data[15] &&
                                 chonkle[13] == packet_data[16] &&
                                 chonkle[14] == packet_data[17] )
                            {
                                passed = 1;
                            }

                            if ( !passed )
                            {
                                // next magic

                                __u8 * magic = state->previous_magic;

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

                                hash ^= ( packet_bytes      ) & 0xFF;
                                hash *= 0x00000100000001B3;

                                hash ^= ( packet_bytes >> 8 );
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

                                if ( chonkle[0] == packet_data[3]   &&
                                     chonkle[1] == packet_data[4]   &&
                                     chonkle[2] == packet_data[5]   &&
                                     chonkle[3] == packet_data[6]   &&
                                     chonkle[4] == packet_data[7]   &&
                                     chonkle[5] == packet_data[8]   &&
                                     chonkle[6] == packet_data[9]   &&
                                     chonkle[7] == packet_data[10]  &&
                                     chonkle[8] == packet_data[11]  &&
                                     chonkle[9] == packet_data[12]  &&
                                     chonkle[10] == packet_data[13] &&
                                     chonkle[11] == packet_data[14] &&
                                     chonkle[12] == packet_data[15] &&
                                     chonkle[13] == packet_data[16] &&
                                     chonkle[14] == packet_data[17] )
                                {
                                    passed = 1;
                                }

                                if ( !passed )
                                {
                                    relay_printf( "advanced packet filter dropped packet (b) [%d]", packet_data[0] );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }
                            }
                        }

#endif // #if RELAY_ADVANCED_PACKET_FILTER
                        
                        __u8 packet_type = packet_data[0];

                        switch ( packet_type )
                        {
                            case RELAY_PING_PACKET:
                            {
                                relay_printf( "relay ping packet from %x:%d to %x:%d", bpf_htonl( ip->saddr ), bpf_htons( udp->source ), bpf_htonl( ip->daddr ), bpf_htons( udp->dest ) );

                                INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED );

                                // IMPORTANT: for the verifier, because it's fucking stupid
                                if ( (void*) packet_data + 18 + 8 + 8 + 1 + RELAY_PING_TOKEN_BYTES > data_end )
                                {
                                    relay_printf( "relay ping packet has wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                if ( (void*) packet_data + 18 + 8 + 8 + 1 + RELAY_PING_TOKEN_BYTES != data_end )
                                {
                                    relay_printf( "relay ping packet has wrong size" );
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
                                    relay_printf( "ping token expired: %lld < %lld", expire_timestamp, state->current_timestamp );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_EXPIRED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 relay_map_key = ( ( (__u64)ip->saddr ) << 32 ) | udp->source;
                                void * relay_map_value = bpf_map_lookup_elem( &relay_map, &relay_map_key );
                                if ( relay_map_value == NULL )
                                {
                                    relay_printf( "ping from unknown relay %x:%d", bpf_ntohl( ip->saddr ), bpf_ntohs( udp->source ) );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_UNKNOWN_RELAY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                // first try with relay public address as dest

                                struct ping_token_data verify_data;
                                verify_data.source_address = ip->saddr;
                                verify_data.source_port = udp->source;
                                verify_data.dest_address = config->relay_public_address;
                                verify_data.dest_port = udp->dest;
                                verify_data.expire_timestamp = expire_timestamp;
                                memcpy( verify_data.ping_key, state->ping_key, RELAY_PING_KEY_BYTES );

                                __u8 * ping_token = payload + 8 + 8 + 1;

                                __u8 hash[RELAY_PING_TOKEN_BYTES];
                                bpf_relay_sha256( &verify_data, sizeof(struct ping_token_data), hash, RELAY_PING_TOKEN_BYTES );
                                if ( hash[0] != ping_token[0] || 
                                     hash[1] != ping_token[1] || 
                                     hash[2] != ping_token[2] || 
                                     hash[3] != ping_token[3] || 
                                     hash[4] != ping_token[4] || 
                                     hash[5] != ping_token[5] || 
                                     hash[6] != ping_token[6] || 
                                     hash[7] != ping_token[7] || 
                                     hash[8] != ping_token[8] || 
                                     hash[9] != ping_token[9] || 
                                     hash[10] != ping_token[10] || 
                                     hash[11] != ping_token[11] || 
                                     hash[12] != ping_token[12] || 
                                     hash[13] != ping_token[13] || 
                                     hash[14] != ping_token[14] || 
                                     hash[15] != ping_token[15] || 
                                     hash[16] != ping_token[16] || 
                                     hash[17] != ping_token[17] || 
                                     hash[18] != ping_token[18] || 
                                     hash[19] != ping_token[19] || 
                                     hash[20] != ping_token[20] || 
                                     hash[21] != ping_token[21] || 
                                     hash[22] != ping_token[22] || 
                                     hash[23] != ping_token[23] || 
                                     hash[24] != ping_token[24] || 
                                     hash[25] != ping_token[25] || 
                                     hash[26] != ping_token[26] || 
                                     hash[27] != ping_token[27] || 
                                     hash[28] != ping_token[28] || 
                                     hash[29] != ping_token[29] || 
                                     hash[30] != ping_token[30] || 
                                     hash[31] != ping_token[31] )
                                {
                                    // next try with relay internal address

                                    verify_data.dest_address = config->relay_internal_address;
                                    bpf_relay_sha256( &verify_data, sizeof(struct ping_token_data), hash, RELAY_PING_TOKEN_BYTES );

                                    if ( hash[0] != ping_token[0] || 
                                         hash[1] != ping_token[1] || 
                                         hash[2] != ping_token[2] || 
                                         hash[3] != ping_token[3] || 
                                         hash[4] != ping_token[4] || 
                                         hash[5] != ping_token[5] || 
                                         hash[6] != ping_token[6] || 
                                         hash[7] != ping_token[7] || 
                                         hash[8] != ping_token[8] || 
                                         hash[9] != ping_token[9] || 
                                         hash[10] != ping_token[10] || 
                                         hash[11] != ping_token[11] || 
                                         hash[12] != ping_token[12] || 
                                         hash[13] != ping_token[13] || 
                                         hash[14] != ping_token[14] || 
                                         hash[15] != ping_token[15] || 
                                         hash[16] != ping_token[16] || 
                                         hash[17] != ping_token[17] || 
                                         hash[18] != ping_token[18] || 
                                         hash[19] != ping_token[19] || 
                                         hash[20] != ping_token[20] || 
                                         hash[21] != ping_token[21] || 
                                         hash[22] != ping_token[22] || 
                                         hash[23] != ping_token[23] || 
                                         hash[24] != ping_token[24] || 
                                         hash[25] != ping_token[25] || 
                                         hash[26] != ping_token[26] || 
                                         hash[27] != ping_token[27] || 
                                         hash[28] != ping_token[28] || 
                                         hash[29] != ping_token[29] || 
                                         hash[30] != ping_token[30] || 
                                         hash[31] != ping_token[31] )
                                    {
                                        relay_printf( "relay ping token did not verify" );
                                        INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PING_PACKET_DID_NOT_VERIFY );
                                        INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                        ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                        return XDP_DROP;
                                    }
                                }

                                struct whitelist_key key;
                                key.address = ip->saddr;
                                key.port = udp->source;
                                
                                struct whitelist_value value;
                                value.expire_timestamp = state->current_timestamp + WHITELIST_TIMEOUT;
                                memcpy( value.source_address, eth->h_source, 6 );
                                memcpy( value.dest_address, eth->h_dest, 6 );

                                bpf_map_update_elem( &whitelist_map, &key, &value, BPF_ANY );

                                packet_data[0] = RELAY_PONG_PACKET;

                                const int payload_bytes = 18 + 8;

                                relay_reflect_packet( data, payload_bytes, state->current_magic, config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL );

                                bpf_xdp_adjust_tail( ctx, -( 8 + 1 + RELAY_PING_TOKEN_BYTES ) );

                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PONG_PACKET_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, payload_bytes + sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr) );
        
                                return XDP_TX;
                            }

                            case RELAY_CLIENT_PING_PACKET:
                            {
                                relay_printf( "client ping packet from %x:%d to %x:%d", bpf_htonl( ip->saddr ), bpf_htons( udp->source ), bpf_htonl( ip->daddr ), bpf_htons( udp->dest ) );

                                INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED );

                                // IMPORTANT: for the verifier, because it's fucking stupid
                                if ( (void*) packet_data + 18 + 8 + 8 + 8 + RELAY_PING_TOKEN_BYTES > data_end )
                                {
                                    relay_printf( "client ping packet has wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_PING_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                if ( (void*) packet_data + 18 + 8 + 8 + 8 + RELAY_PING_TOKEN_BYTES != data_end )
                                {
                                    relay_printf( "client ping packet has wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_PING_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u8 * payload = packet_data + 18;

                                __u64 expire_timestamp;
                                expire_timestamp  = payload[8 + 8];
                                expire_timestamp |= ( ( (__u64)( payload[8 + 8 + 1] ) ) << 8  );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 8 + 2] ) ) << 16 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 8 + 3] ) ) << 24 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 8 + 4] ) ) << 32 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 8 + 5] ) ) << 40 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 8 + 6] ) ) << 48 );
                                expire_timestamp |= ( ( (__u64)( payload[8 + 8 + 7] ) ) << 56 );

                                if ( expire_timestamp < state->current_timestamp )
                                {
                                    relay_printf( "client ping token expired" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_PING_PACKET_EXPIRED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct ping_token_data verify_data;
                                verify_data.source_address = ip->saddr;
                                verify_data.source_port = 0; // IMPORTANT: Some NAT change the client port, so it is set to zero in client ping token
                                verify_data.dest_address = config->relay_public_address;
                                verify_data.dest_port = udp->dest;
                                verify_data.expire_timestamp = expire_timestamp;
                                memcpy( verify_data.ping_key, state->ping_key, RELAY_PING_KEY_BYTES );

                                __u8 hash[RELAY_PING_TOKEN_BYTES];
                                bpf_relay_sha256( &verify_data, sizeof(struct ping_token_data), hash, 32 );

                                __u8 * ping_token = payload + 8 + 8 + 8;

                                if ( hash[0] != ping_token[0] || 
                                     hash[1] != ping_token[1] || 
                                     hash[2] != ping_token[2] || 
                                     hash[3] != ping_token[3] || 
                                     hash[4] != ping_token[4] || 
                                     hash[5] != ping_token[5] || 
                                     hash[6] != ping_token[6] || 
                                     hash[7] != ping_token[7] || 
                                     hash[8] != ping_token[8] || 
                                     hash[9] != ping_token[9] || 
                                     hash[10] != ping_token[10] || 
                                     hash[11] != ping_token[11] || 
                                     hash[12] != ping_token[12] || 
                                     hash[13] != ping_token[13] || 
                                     hash[14] != ping_token[14] || 
                                     hash[15] != ping_token[15] || 
                                     hash[16] != ping_token[16] || 
                                     hash[17] != ping_token[17] || 
                                     hash[18] != ping_token[18] || 
                                     hash[19] != ping_token[19] || 
                                     hash[20] != ping_token[20] || 
                                     hash[21] != ping_token[21] || 
                                     hash[22] != ping_token[22] || 
                                     hash[23] != ping_token[23] || 
                                     hash[24] != ping_token[24] || 
                                     hash[25] != ping_token[25] || 
                                     hash[26] != ping_token[26] || 
                                     hash[27] != ping_token[27] || 
                                     hash[28] != ping_token[28] || 
                                     hash[29] != ping_token[29] || 
                                     hash[30] != ping_token[30] || 
                                     hash[31] != ping_token[31] )
                                {
                                    relay_printf( "client ping token did not verify" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_PING_PACKET_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }
                                
                                struct whitelist_key key;
                                key.address = ip->saddr;
                                key.port = udp->source;
                                
                                struct whitelist_value value;
                                value.expire_timestamp = state->current_timestamp + WHITELIST_TIMEOUT;
                                memcpy( value.source_address, eth->h_source, 6 );
                                memcpy( value.dest_address, eth->h_dest, 6 );

                                bpf_map_update_elem( &whitelist_map, &key, &value, BPF_ANY );

                                packet_data[0] = RELAY_CLIENT_PONG_PACKET;

                                const int payload_bytes = 18 + 8 + 8;

                                relay_reflect_packet( data, payload_bytes, state->current_magic, config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL );

                                bpf_xdp_adjust_tail( ctx, -( 8 + RELAY_PING_TOKEN_BYTES ) );

                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_PING_PACKET_RESPONDED_WITH_PONG );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, payload_bytes + sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr) );
        
                                return XDP_TX;
                            }
                            break;

                            case RELAY_SERVER_PING_PACKET:
                            {
                                relay_printf( "server ping packet from %x:%d to %x:%d", bpf_htonl( ip->saddr ), bpf_htons( udp->source ), bpf_htonl( ip->daddr ), bpf_htons( udp->dest ) );

                                INCREMENT_COUNTER( RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED );

                                // IMPORTANT: for the verifier, because it's fucking stupid
                                if ( (void*) packet_data + 18 + 8 + 8 + RELAY_PING_TOKEN_BYTES > data_end )
                                {
                                    relay_printf( "server ping packet has wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_PING_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                if ( (void*) packet_data + 18 + 8 + 8 + RELAY_PING_TOKEN_BYTES != data_end )
                                {
                                    relay_printf( "server ping packet has wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_PING_PACKET_WRONG_SIZE );
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
                                    relay_printf( "server ping token expired" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_PING_PACKET_EXPIRED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct ping_token_data verify_data;
                                verify_data.source_address = ip->saddr;
                                verify_data.source_port = udp->source;
                                verify_data.dest_address = config->relay_public_address;
                                verify_data.dest_port = udp->dest;
                                verify_data.expire_timestamp = expire_timestamp;
                                memcpy( verify_data.ping_key, state->ping_key, RELAY_PING_KEY_BYTES );

                                __u8 hash[RELAY_PING_TOKEN_BYTES];
                                bpf_relay_sha256( &verify_data, sizeof(struct ping_token_data), hash, 32 );

                                __u8 * ping_token = payload + 8 + 8;

                                if ( hash[0] != ping_token[0] || 
                                     hash[1] != ping_token[1] || 
                                     hash[2] != ping_token[2] || 
                                     hash[3] != ping_token[3] || 
                                     hash[4] != ping_token[4] || 
                                     hash[5] != ping_token[5] || 
                                     hash[6] != ping_token[6] || 
                                     hash[7] != ping_token[7] || 
                                     hash[8] != ping_token[8] || 
                                     hash[9] != ping_token[9] || 
                                     hash[10] != ping_token[10] || 
                                     hash[11] != ping_token[11] || 
                                     hash[12] != ping_token[12] || 
                                     hash[13] != ping_token[13] || 
                                     hash[14] != ping_token[14] || 
                                     hash[15] != ping_token[15] || 
                                     hash[16] != ping_token[16] || 
                                     hash[17] != ping_token[17] || 
                                     hash[18] != ping_token[18] || 
                                     hash[19] != ping_token[19] || 
                                     hash[20] != ping_token[20] || 
                                     hash[21] != ping_token[21] || 
                                     hash[22] != ping_token[22] || 
                                     hash[23] != ping_token[23] || 
                                     hash[24] != ping_token[24] || 
                                     hash[25] != ping_token[25] || 
                                     hash[26] != ping_token[26] || 
                                     hash[27] != ping_token[27] || 
                                     hash[28] != ping_token[28] || 
                                     hash[29] != ping_token[29] || 
                                     hash[30] != ping_token[30] || 
                                     hash[31] != ping_token[31] )
                                {
                                    relay_printf( "server ping token did not verify" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_PING_PACKET_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }
                                
                                struct whitelist_key key;
                                key.address = ip->saddr;
                                key.port = udp->source;
                                
                                struct whitelist_value value;
                                value.expire_timestamp = state->current_timestamp + WHITELIST_TIMEOUT;
                                memcpy( value.source_address, eth->h_source, 6 );
                                memcpy( value.dest_address, eth->h_dest, 6 );

                                bpf_map_update_elem( &whitelist_map, &key, &value, BPF_ANY );

                                packet_data[0] = RELAY_SERVER_PONG_PACKET;

                                const int payload_bytes = 18 + 8;

                                ip->daddr = config->relay_public_address;       // IMPORTANT: We must respond from the relay public address or it will get filtered out

                                relay_reflect_packet( data, payload_bytes, state->current_magic, config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL );

                                bpf_xdp_adjust_tail( ctx, -( 8 + RELAY_PING_TOKEN_BYTES ) );

                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                INCREMENT_COUNTER( RELAY_COUNTER_SERVER_PING_PACKET_RESPONDED_WITH_PONG );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, payload_bytes + sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr) );
        
                                return XDP_TX;
                            }
                            break;
                        }

                        // if the packet is not from a whitelisted address, drop it

                        struct whitelist_key key;
                        key.address = ip->saddr;
                        key.port = udp->source;

                        struct whitelist_value * whitelist = (struct whitelist_value*) bpf_map_lookup_elem( &whitelist_map, &key );
                        if ( whitelist == NULL )
                        {
                            relay_printf( "address %x:%d is not in whitelist", bpf_ntohl( ip->saddr ), bpf_ntohs( udp->source ) );
                            INCREMENT_COUNTER( RELAY_COUNTER_NOT_IN_WHITELIST );
                            INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                            ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                            return XDP_DROP;
                        }

                        // process packets types that should only be processed after whitelist check

                        switch ( packet_type )
                        {
                            case RELAY_PONG_PACKET:
                            {
                                relay_printf( "relay pong packet from %x:%d to %x:%d", bpf_htonl( ip->saddr ), bpf_htons( udp->source ), bpf_htonl( ip->daddr ), bpf_htons( udp->dest ) );

                                INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED );

                                // IMPORTANT: for the verifier, because it's fucking stupid
                                if ( (void*) packet_data + 18 + 8 > data_end )
                                {
                                    relay_printf( "relay pong packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                if ( (void*) packet_data + 18 + 8 != data_end )
                                {
                                    relay_printf( "relay pong packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 relay_map_key = ( ( (__u64)ip->saddr ) << 32 ) | udp->source;
                                void * relay_map_value = bpf_map_lookup_elem( &relay_map, &relay_map_key );
                                if ( relay_map_value == NULL )
                                {
                                    relay_printf( "unknown relay %x:%d", bpf_ntohl( ip->saddr ), bpf_ntohs( udp->source ) );
                                    INCREMENT_COUNTER( RELAY_COUNTER_RELAY_PONG_PACKET_UNKNOWN_RELAY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_PASS;
                            }
                            break;

                            case RELAY_ROUTE_REQUEST_PACKET:
                            {
                                relay_printf( "route request packet" );

                                INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED );

                                if ( (void*) packet_data + 18 + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES > data_end )
                                {
                                    relay_printf( "route request packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_REQUEST_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct decrypt_route_token_data decrypt_data;
                                memcpy( decrypt_data.relay_secret_key, config->relay_secret_key, RELAY_SECRET_KEY_BYTES );
                                if ( relay_decrypt_route_token( &decrypt_data, packet_data + 18, RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES ) == 0 )
                                {
                                    relay_printf( "route request could not decrypt route token" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_DECRYPT_ROUTE_TOKEN );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct route_token * token = (struct route_token*) ( packet_data + 18 + 24 );

                                if ( token->expire_timestamp < state->current_timestamp )
                                {
                                    relay_printf( "route request route token expired" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct session_data session;

                                memcpy( session.session_private_key, token->session_private_key, RELAY_SESSION_PRIVATE_KEY_BYTES );
                                session.expire_timestamp = token->expire_timestamp;
                                session.session_id = token->session_id;
                                session.envelope_kbps_up = token->envelope_kbps_up;
                                session.envelope_kbps_down = token->envelope_kbps_down;
                                session.next_address = token->next_address;
                                session.prev_address = token->prev_address;
                                session.next_port = token->next_port;
                                session.prev_port = token->prev_port;
                                session.session_version = token->session_version;
                                session.next_internal = token->next_internal;
                                session.prev_internal = token->prev_internal;

                                if ( token->prev_port == 0 )
                                {
                                    session.first_hop = 1;
                                    session.prev_port = udp->source;
                                }
                                else
                                {
                                    session.first_hop = 0;
                                }

                                session.payload_client_to_server_sequence = 0;
                                session.payload_server_to_client_sequence = 0;
                                session.special_client_to_server_sequence = 0;
                                session.special_server_to_client_sequence = 0;

                                struct session_key key;
                                key.session_id = token->session_id;
                                key.session_version = token->session_version;
                                if ( bpf_map_update_elem( &session_map, &key, &session, BPF_NOEXIST ) == 0 )
                                {
                                    relay_printf( "created session 0x%llx:%d", session.session_id, session.session_version );
                                }

                                memcpy( data + RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES, data, sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr) );

                                data += RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES;

                                packet_data = (__u8*) ( data + sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr) );

                                packet_data[0] = RELAY_ROUTE_REQUEST_PACKET;

                                struct redirect_args_t args;
                                args.data = data;
                                args.payload_bytes = data_end - (void*)packet_data;
                                args.source_address = config->relay_internal_address;
                                args.dest_address = session.next_address;
                                args.source_port = config->relay_port;
                                args.dest_port = session.next_port;
                                args.magic = state->current_magic;
                                args.gateway_ethernet_address = config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL;

                                relay_printf( "route request forward to next hop: %x:%d -> %x.%d", args.source_address, args.source_port, args.dest_address, args.dest_port );

                                int result = relay_redirect_packet( &args );
                                if ( result == XDP_DROP )
                                {
                                    relay_printf( "route request redirect address is not in whitelist" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                bpf_xdp_adjust_head( ctx, RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES );

                                INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP );
                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, data_end - data );

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_TX;
                            }
                            break;

                            case RELAY_ROUTE_RESPONSE_PACKET:
                            {
                                relay_printf( "route response packet" );

                                INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED );

                                __u8 * header = packet_data + 18;

                                // IMPORTANT: required for verifier because it's fucking stupid as shit
                                if ( (void*) header + RELAY_HEADER_BYTES > data_end )
                                {
                                    relay_printf( "route response packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                if ( (void*) header + RELAY_HEADER_BYTES != data_end )
                                {
                                    relay_printf( "route response packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 session_id;
                                session_id  = header[8];
                                session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );

                                __u8 session_version = header[8+8];

                                struct session_key key;
                                key.session_id = session_id;
                                key.session_version = session_version;
                                struct session_data * session = (struct session_data*) bpf_map_lookup_elem( &session_map, &key );
                                if ( session == NULL )
                                {
                                    relay_printf( "route response packet could not find session" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 packet_sequence = 0;
                                packet_sequence  = header[0];
                                packet_sequence |= ( ( (__u64)( header[1] ) ) << 8  );
                                packet_sequence |= ( ( (__u64)( header[2] ) ) << 16 );
                                packet_sequence |= ( ( (__u64)( header[3] ) ) << 24 );
                                packet_sequence |= ( ( (__u64)( header[4] ) ) << 32 );
                                packet_sequence |= ( ( (__u64)( header[5] ) ) << 40 );
                                packet_sequence |= ( ( (__u64)( header[6] ) ) << 48 );
                                packet_sequence |= ( ( (__u64)( header[7] ) ) << 56 );

                                __u64 server_to_client_sequence = session->special_server_to_client_sequence;

                                if ( packet_sequence <= server_to_client_sequence )
                                {
                                    relay_printf( "route response packet already received" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct header_data verify_data;
                                memset( &verify_data, 0, sizeof(struct header_data) );
                                memcpy( verify_data.session_private_key, session->session_private_key, RELAY_SESSION_PRIVATE_KEY_BYTES );
                                verify_data.packet_type = packet_type;
                                verify_data.packet_sequence = packet_sequence;
                                verify_data.session_id = session_id;
                                verify_data.session_version = session_version;
                                
                                __u8 hash[32];
                                bpf_relay_sha256( &verify_data, sizeof(struct header_data), hash, 32 );

                                __u8 * expected = header + 8 + 8 + 1;
                                
                                if ( hash[0] != expected[0] || 
                                     hash[1] != expected[1] || 
                                     hash[2] != expected[2] || 
                                     hash[3] != expected[3] || 
                                     hash[4] != expected[4] || 
                                     hash[5] != expected[5] || 
                                     hash[6] != expected[6] || 
                                     hash[7] != expected[7] )
                                {
                                    relay_printf( "route response packet header did not verify" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __sync_bool_compare_and_swap( &session->special_server_to_client_sequence, server_to_client_sequence, packet_sequence );

                                relay_printf( "route response packet forward to previous hop" );

                                struct redirect_args_t args;
                                args.data = data;
                                args.payload_bytes = 18 + RELAY_HEADER_BYTES;
                                args.source_address = config->relay_internal_address;
                                args.dest_address = session->prev_address;
                                args.source_port = config->relay_port;
                                args.dest_port = session->prev_port;
                                args.magic = state->current_magic;
                                args.gateway_ethernet_address = config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL;

                                int result = relay_redirect_packet( &args );
                                if ( result == XDP_DROP )
                                {
                                    relay_printf( "route response packet redirect address is not in whitelist" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                INCREMENT_COUNTER( RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP );
                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, data_end - data );

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_TX;
                            }
                            break;

                            case RELAY_CONTINUE_REQUEST_PACKET:
                            {
                                relay_printf( "continue request packet" );

                                INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED );

                                if ( (void*) packet_data + 18 + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES > data_end )
                                {
                                    relay_printf( "continue request packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                relay_printf( "decrypting continue token" );

                                struct decrypt_continue_token_data decrypt_data;
                                memcpy( decrypt_data.relay_secret_key, config->relay_secret_key, RELAY_SECRET_KEY_BYTES );
                                if ( relay_decrypt_continue_token( &decrypt_data, packet_data + 18, RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES ) == 0 )
                                {
                                    relay_printf( "could not decrypt continue token" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_DECRYPT_CONTINUE_TOKEN );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct continue_token * token = (struct continue_token*) ( packet_data + 18 + 24 );

                                if ( token->expire_timestamp < state->current_timestamp )
                                {
                                    relay_printf( "continue request packet continue token expired" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct session_key key;
                                key.session_id = token->session_id;
                                key.session_version = token->session_version;
                                struct session_data * session = (struct session_data*) bpf_map_lookup_elem( &session_map, &key );
                                if ( session == NULL )
                                {
                                    relay_printf( "continue request packet could not find session" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_FIND_SESSION );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 current_expire_timestamp = session->expire_timestamp;

                                __sync_bool_compare_and_swap( &session->expire_timestamp, current_expire_timestamp, token->expire_timestamp );

                                memmove( data + RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES, data, sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr) );

                                data += RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES;

                                packet_data = (__u8*) ( data + sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct udphdr) );

                                packet_data[0] = RELAY_CONTINUE_REQUEST_PACKET;

                                relay_printf( "continue request packet forward to next hop" );

                                struct redirect_args_t args;
                                args.data = data;
                                args.payload_bytes = data_end - (void*)packet_data;
                                args.source_address = config->relay_internal_address;
                                args.dest_address = session->next_address;
                                args.source_port = config->relay_port;
                                args.dest_port = session->next_port;
                                args.magic = state->current_magic;
                                args.gateway_ethernet_address = config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL;

                                int result = relay_redirect_packet( &args );
                                if ( result == XDP_DROP )
                                {
                                    relay_printf( "continue request packet redirect address is not in whitelist" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                bpf_xdp_adjust_head( ctx, RELAY_ENCRYPTED_CONTINUE_TOKEN_BYTES );

                                INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP );
                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, data_end - data );

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_TX;
                            }
                            break;

                            case RELAY_CONTINUE_RESPONSE_PACKET:
                            {
                                relay_printf( "continue response packet" );

                                INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED );

                                __u8 * header = packet_data + 18;

                                // IMPORTANT: required for verifier because it's dumber than shit
                                if ( (void*) header + RELAY_HEADER_BYTES > data_end )
                                {
                                    relay_printf( "continue response packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                if ( (void*) header + RELAY_HEADER_BYTES != data_end )
                                {
                                    relay_printf( "continue response packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 session_id;
                                session_id  = header[8];
                                session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );

                                __u8 session_version = header[8+8];

                                struct session_key key;
                                key.session_id = session_id;
                                key.session_version = session_version;
                                struct session_data * session = (struct session_data*) bpf_map_lookup_elem( &session_map, &key );
                                if ( session == NULL )
                                {
                                    relay_printf( "continue response packet could not find session" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 packet_sequence = 0;
                                packet_sequence  = header[0];
                                packet_sequence |= ( ( (__u64)( header[1] ) ) << 8  );
                                packet_sequence |= ( ( (__u64)( header[2] ) ) << 16 );
                                packet_sequence |= ( ( (__u64)( header[3] ) ) << 24 );
                                packet_sequence |= ( ( (__u64)( header[4] ) ) << 32 );
                                packet_sequence |= ( ( (__u64)( header[5] ) ) << 40 );
                                packet_sequence |= ( ( (__u64)( header[6] ) ) << 48 );
                                packet_sequence |= ( ( (__u64)( header[7] ) ) << 56 );

                                __u64 server_to_client_sequence = session->special_server_to_client_sequence;

                                __sync_bool_compare_and_swap( &session->special_server_to_client_sequence, server_to_client_sequence, packet_sequence );

                                if ( packet_sequence <= server_to_client_sequence )
                                {
                                    relay_printf( "continue response packet already received" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct header_data verify_data;
                                memset( &verify_data, 0, sizeof(struct header_data) );
                                memcpy( verify_data.session_private_key, session->session_private_key, RELAY_SESSION_PRIVATE_KEY_BYTES );
                                verify_data.packet_type = packet_type;
                                verify_data.packet_sequence = packet_sequence;
                                verify_data.session_id  = header[8];
                                verify_data.session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                verify_data.session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                verify_data.session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                verify_data.session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                verify_data.session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                verify_data.session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                verify_data.session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );
                                verify_data.session_version = header[8+8];

                                __u8 hash[32];
                                bpf_relay_sha256( &verify_data, sizeof(struct header_data), hash, 32 );

                                __u8 * expected = header + 8 + 8 + 1;
                                
                                if ( hash[0] != expected[0] || 
                                     hash[1] != expected[1] || 
                                     hash[2] != expected[2] || 
                                     hash[3] != expected[3] || 
                                     hash[4] != expected[4] || 
                                     hash[5] != expected[5] || 
                                     hash[6] != expected[6] || 
                                     hash[7] != expected[7] )
                                {
                                    relay_printf( "continue response packet header did not verify" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                session->special_server_to_client_sequence = packet_sequence;

                                relay_printf( "continue response packet forward to previous hop" );

                                struct redirect_args_t args;
                                args.data = data;
                                args.payload_bytes = 18 + RELAY_HEADER_BYTES;
                                args.source_address = config->relay_internal_address;
                                args.dest_address = session->prev_address;
                                args.source_port = config->relay_port;
                                args.dest_port = session->prev_port;
                                args.magic = state->current_magic;
                                args.gateway_ethernet_address = config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL;

                                int result = relay_redirect_packet( &args );
                                if ( result == XDP_DROP )
                                {
                                    relay_printf( "continue response packet redirect address is not in whitelist" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                INCREMENT_COUNTER( RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP );
                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, data_end - data );

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_TX;
                            }
                            break;

                            case RELAY_CLIENT_TO_SERVER_PACKET:
                            {
                                relay_printf( "client to server packet" );

                                INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED );

                                __u8 * header = packet_data + 18;

                                if ( (void*)header + RELAY_HEADER_BYTES > data_end )
                                {
                                    relay_printf( "client to server packet is too small" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                int packet_bytes = data_end - (void*)udp - sizeof(struct udphdr);

                                int payload_bytes = packet_bytes - ( sizeof(struct udphdr) + 18 + RELAY_HEADER_BYTES );

                                if ( payload_bytes > RELAY_MTU )
                                {
                                    relay_printf( "client to server packet is too big" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 session_id;
                                session_id  = header[8];
                                session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );

                                __u8 session_version = header[8+8];

                                struct session_key key;
                                key.session_id = session_id;
                                key.session_version = session_version;
                                struct session_data * session = (struct session_data*) bpf_map_lookup_elem( &session_map, &key );
                                if ( session == NULL )
                                {
                                    relay_printf( "client to server packet could not find session" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 packet_sequence = 0;
                                packet_sequence  = header[0];
                                packet_sequence |= ( ( (__u64)( header[1] ) ) << 8  );
                                packet_sequence |= ( ( (__u64)( header[2] ) ) << 16 );
                                packet_sequence |= ( ( (__u64)( header[3] ) ) << 24 );
                                packet_sequence |= ( ( (__u64)( header[4] ) ) << 32 );
                                packet_sequence |= ( ( (__u64)( header[5] ) ) << 40 );
                                packet_sequence |= ( ( (__u64)( header[6] ) ) << 48 );
                                packet_sequence |= ( ( (__u64)( header[7] ) ) << 56 );

                                __u64 client_to_server_sequence = session->payload_client_to_server_sequence;

                                if ( packet_sequence <= client_to_server_sequence )
                                {
                                    relay_printf( "client to server packet already received" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct header_data verify_data;
                                memset( &verify_data, 0, sizeof(struct header_data) );
                                memcpy( verify_data.session_private_key, session->session_private_key, RELAY_SESSION_PRIVATE_KEY_BYTES );
                                verify_data.packet_type = packet_type;
                                verify_data.packet_sequence = packet_sequence;
                                verify_data.session_id  = header[8];
                                verify_data.session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                verify_data.session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                verify_data.session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                verify_data.session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                verify_data.session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                verify_data.session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                verify_data.session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );
                                verify_data.session_version = header[8+8];

                                __u8 hash[32];
                                bpf_relay_sha256( &verify_data, sizeof(struct header_data), hash, 32 );

                                __u8 * expected = header + 8 + 8 + 1;
                                
                                if ( hash[0] != expected[0] || 
                                     hash[1] != expected[1] || 
                                     hash[2] != expected[2] || 
                                     hash[3] != expected[3] || 
                                     hash[4] != expected[4] || 
                                     hash[5] != expected[5] || 
                                     hash[6] != expected[6] || 
                                     hash[7] != expected[7] )
                                {
                                    relay_printf( "client to server packet header did not verify" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_HEADER_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                } 

                                __sync_bool_compare_and_swap( &session->payload_client_to_server_sequence, client_to_server_sequence, packet_sequence );

                                relay_printf( "client to server packet forward to next hop" );

                                struct redirect_args_t args;
                                args.data = data;
                                args.payload_bytes = (int) ( data_end - (void*)packet_data );
                                args.source_address = config->relay_internal_address;
                                args.dest_address = session->next_address;
                                args.source_port = config->relay_port;
                                args.dest_port = session->next_port;
                                args.magic = state->current_magic;
                                args.gateway_ethernet_address = config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL;

                                int result = relay_redirect_packet( &args );
                                if ( result == XDP_DROP )
                                {
                                    relay_printf( "client to server packet redirect address is not in whitelist" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                INCREMENT_COUNTER( RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP );
                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, data_end - data );

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_TX;
                            }
                            break;

                            case RELAY_SERVER_TO_CLIENT_PACKET:
                            {
                                relay_printf( "server to client packet" );

                                INCREMENT_COUNTER( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED );

                                __u8 * header = packet_data + 18;

                                if ( (void*)header + RELAY_HEADER_BYTES > data_end )
                                {
                                    relay_printf( "server to client packet is too small" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                int packet_bytes = data_end - (void*)udp - sizeof(struct udphdr);

                                int payload_bytes = packet_bytes - ( sizeof(struct udphdr) + 18 + RELAY_HEADER_BYTES );

                                if ( payload_bytes > RELAY_MTU )
                                {
                                    relay_printf( "server to client packet is too big" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 session_id;
                                session_id  = header[8];
                                session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );

                                __u8 session_version = header[8+8];

                                struct session_key key;
                                key.session_id = session_id;
                                key.session_version = session_version;
                                struct session_data * session = (struct session_data*) bpf_map_lookup_elem( &session_map, &key );
                                if ( session == NULL )
                                {
                                    relay_printf( "server to client packet could not find session" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 packet_sequence = 0;
                                packet_sequence  = header[0];
                                packet_sequence |= ( ( (__u64)( header[1] ) ) << 8  );
                                packet_sequence |= ( ( (__u64)( header[2] ) ) << 16 );
                                packet_sequence |= ( ( (__u64)( header[3] ) ) << 24 );
                                packet_sequence |= ( ( (__u64)( header[4] ) ) << 32 );
                                packet_sequence |= ( ( (__u64)( header[5] ) ) << 40 );
                                packet_sequence |= ( ( (__u64)( header[6] ) ) << 48 );
                                packet_sequence |= ( ( (__u64)( header[7] ) ) << 56 );

                                __u64 server_to_client_sequence = session->payload_server_to_client_sequence;

                                if ( packet_sequence <= server_to_client_sequence )
                                {
                                    relay_printf( "server to client packet already received: %lld < %lld" , packet_sequence, server_to_client_sequence );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct header_data verify_data;
                                memset( &verify_data, 0, sizeof(struct header_data) );
                                memcpy( verify_data.session_private_key, session->session_private_key, RELAY_SESSION_PRIVATE_KEY_BYTES );
                                verify_data.packet_type = packet_type;
                                verify_data.packet_sequence = packet_sequence;
                                verify_data.session_id  = header[8];
                                verify_data.session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                verify_data.session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                verify_data.session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                verify_data.session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                verify_data.session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                verify_data.session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                verify_data.session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );
                                verify_data.session_version = header[8+8];

                                __u8 hash[32];
                                bpf_relay_sha256( &verify_data, sizeof(struct header_data), hash, 32 );

                                __u8 * expected = header + 8 + 8 + 1;
                                
                                if ( hash[0] != expected[0] || 
                                     hash[1] != expected[1] || 
                                     hash[2] != expected[2] || 
                                     hash[3] != expected[3] || 
                                     hash[4] != expected[4] || 
                                     hash[5] != expected[5] || 
                                     hash[6] != expected[6] || 
                                     hash[7] != expected[7] )
                                {
                                    relay_printf( "server to client packet header did not verify" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_HEADER_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __sync_bool_compare_and_swap( &session->payload_server_to_client_sequence, server_to_client_sequence, packet_sequence );

                                relay_printf( "server to client packet forward to previous hop" );

                                struct redirect_args_t args;
                                args.data = data;
                                args.payload_bytes = (int) ( data_end - (void*)packet_data );
                                args.source_address = config->relay_internal_address;
                                args.dest_address = session->prev_address;
                                args.source_port = config->relay_port;
                                args.dest_port = session->prev_port;
                                args.magic = state->current_magic;
                                args.gateway_ethernet_address = config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL;

                                int result = relay_redirect_packet( &args );
                                if ( result == XDP_DROP )
                                {
                                    relay_printf( "server to client packet redirect address is not in whitelist" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                INCREMENT_COUNTER( RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP );
                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, data_end - data );

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_TX;
                            }
                            break;

                            case RELAY_SESSION_PING_PACKET:
                            {
                                relay_printf( "session ping packet" );

                                INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED );

                                __u8 * header = packet_data + 18;

                                // IMPORTANT: required for verifier because it's thick as a brick
                                if ( (void*) header + RELAY_HEADER_BYTES + 8 > data_end )
                                {
                                    relay_printf( "session ping packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PING_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                if ( (void*) header + RELAY_HEADER_BYTES + 8 != data_end )
                                {
                                    relay_printf( "session ping packet is the wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PING_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 session_id;
                                session_id  = header[8];
                                session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );

                                __u8 session_version = header[8+8];

                                struct session_key key;
                                key.session_id = session_id;
                                key.session_version = session_version;
                                struct session_data * session = (struct session_data*) bpf_map_lookup_elem( &session_map, &key );
                                if ( session == NULL )
                                {
                                    relay_printf( "session ping packet could not find session" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_FIND_SESSION );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 packet_sequence = 0;
                                packet_sequence  = header[0];
                                packet_sequence |= ( ( (__u64)( header[1] ) ) << 8  );
                                packet_sequence |= ( ( (__u64)( header[2] ) ) << 16 );
                                packet_sequence |= ( ( (__u64)( header[3] ) ) << 24 );
                                packet_sequence |= ( ( (__u64)( header[4] ) ) << 32 );
                                packet_sequence |= ( ( (__u64)( header[5] ) ) << 40 );
                                packet_sequence |= ( ( (__u64)( header[6] ) ) << 48 );
                                packet_sequence |= ( ( (__u64)( header[7] ) ) << 56 );

                                __u64 client_to_server_sequence = session->special_client_to_server_sequence;

                                if ( packet_sequence <= client_to_server_sequence )
                                {
                                    relay_printf( "session ping packet already received" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                struct header_data verify_data;
                                memset( &verify_data, 0, sizeof(struct header_data) );
                                memcpy( verify_data.session_private_key, session->session_private_key, RELAY_SESSION_PRIVATE_KEY_BYTES );
                                verify_data.packet_type = packet_type;
                                verify_data.packet_sequence = packet_sequence;
                                verify_data.session_id  = header[8];
                                verify_data.session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                verify_data.session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                verify_data.session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                verify_data.session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                verify_data.session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                verify_data.session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                verify_data.session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );
                                verify_data.session_version = header[8+8];

                                __u8 hash[32];
                                bpf_relay_sha256( &verify_data, sizeof(struct header_data), hash, 32 );

                                __u8 * expected = header + 8 + 8 + 1;
                                
                                if ( hash[0] != expected[0] || 
                                     hash[1] != expected[1] || 
                                     hash[2] != expected[2] || 
                                     hash[3] != expected[3] || 
                                     hash[4] != expected[4] || 
                                     hash[5] != expected[5] || 
                                     hash[6] != expected[6] || 
                                     hash[7] != expected[7] )
                                {
                                    relay_printf( "session ping packet header did not verify" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PING_PACKET_HEADER_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __sync_bool_compare_and_swap( &session->special_client_to_server_sequence, client_to_server_sequence, packet_sequence );

                                relay_printf( "session ping packet forward to next hop" );

                                struct redirect_args_t args;
                                args.data = data;
                                args.payload_bytes = 18 + RELAY_HEADER_BYTES + 8;
                                args.source_address = config->relay_internal_address;
                                args.dest_address = session->next_address;
                                args.source_port = config->relay_port;
                                args.dest_port = session->next_port;
                                args.magic = state->current_magic;
                                args.gateway_ethernet_address = config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL;

                                int result = relay_redirect_packet( &args );
                                if ( result == XDP_DROP )
                                {
                                    relay_printf( "session ping packet redirect address is not in whitelist" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP );
                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, data_end - data );

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_TX;
                            }
                            break;

                            case RELAY_SESSION_PONG_PACKET:
                            {
                                relay_printf( "session pong packet" );

                                INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED );

                                __u8 * header = packet_data + 18;

                                // IMPORTANT: required for verifier because it's not all there
                                if ( (void*)header + RELAY_HEADER_BYTES + 8 > data_end )
                                {
                                    relay_printf( "session pong packet is wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PONG_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                if ( (void*)header + RELAY_HEADER_BYTES + 8 != data_end )
                                {
                                    relay_printf( "session pong packet is wrong size" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PONG_PACKET_WRONG_SIZE );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 session_id;
                                session_id  = header[8];
                                session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );

                                __u8 session_version = header[8+8];

                                struct session_key key;
                                key.session_id = session_id;
                                key.session_version = session_version;
                                struct session_data * session = (struct session_data*) bpf_map_lookup_elem( &session_map, &key );
                                if ( session == NULL )
                                {
                                    relay_printf( "session pong packet could not find session" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_FIND_SESSION );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __u64 packet_sequence = 0;
                                packet_sequence  = header[0];
                                packet_sequence |= ( ( (__u64)( header[1] ) ) << 8  );
                                packet_sequence |= ( ( (__u64)( header[2] ) ) << 16 );
                                packet_sequence |= ( ( (__u64)( header[3] ) ) << 24 );
                                packet_sequence |= ( ( (__u64)( header[4] ) ) << 32 );
                                packet_sequence |= ( ( (__u64)( header[5] ) ) << 40 );
                                packet_sequence |= ( ( (__u64)( header[6] ) ) << 48 );
                                packet_sequence |= ( ( (__u64)( header[7] ) ) << 56 );

                                __u64 server_to_client_sequence = session->special_server_to_client_sequence;

                                if ( packet_sequence <= server_to_client_sequence )
                                {
                                    relay_printf( "session pong packet already received" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                relay_printf( "session pong packet verifying header" );

                                struct header_data verify_data;
                                memset( &verify_data, 0, sizeof(struct header_data) );
                                memcpy( verify_data.session_private_key, session->session_private_key, RELAY_SESSION_PRIVATE_KEY_BYTES );
                                verify_data.packet_type = packet_type;
                                verify_data.packet_sequence = packet_sequence;
                                verify_data.session_id  = header[8];
                                verify_data.session_id |= ( ( (__u64)( header[8+1] ) ) << 8  );
                                verify_data.session_id |= ( ( (__u64)( header[8+2] ) ) << 16 );
                                verify_data.session_id |= ( ( (__u64)( header[8+3] ) ) << 24 );
                                verify_data.session_id |= ( ( (__u64)( header[8+4] ) ) << 32 );
                                verify_data.session_id |= ( ( (__u64)( header[8+5] ) ) << 40 );
                                verify_data.session_id |= ( ( (__u64)( header[8+6] ) ) << 48 );
                                verify_data.session_id |= ( ( (__u64)( header[8+7] ) ) << 56 );
                                verify_data.session_version = header[8+8];

                                __u8 hash[32];
                                bpf_relay_sha256( &verify_data, sizeof(struct header_data), hash, 32 );

                                __u8 * expected = header + 8 + 8 + 1;
                                
                                if ( hash[0] != expected[0] || 
                                     hash[1] != expected[1] || 
                                     hash[2] != expected[2] || 
                                     hash[3] != expected[3] || 
                                     hash[4] != expected[4] || 
                                     hash[5] != expected[5] || 
                                     hash[6] != expected[6] || 
                                     hash[7] != expected[7] )
                                {
                                    relay_printf( "session pong packet header did not verify" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PONG_PACKET_HEADER_DID_NOT_VERIFY );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                __sync_bool_compare_and_swap( &session->special_server_to_client_sequence, server_to_client_sequence, packet_sequence );
   
                                relay_printf( "session pong packet forward to previous hop" );

                                struct redirect_args_t args;
                                args.data = data;
                                args.payload_bytes = 18 + RELAY_HEADER_BYTES + 8;
                                args.source_address = config->relay_internal_address;
                                args.dest_address = session->prev_address;
                                args.source_port = config->relay_port;
                                args.dest_port = session->prev_port;
                                args.magic = state->current_magic;
                                args.gateway_ethernet_address = config->use_gateway_ethernet_address ? config->gateway_ethernet_address : NULL;

                                int result = relay_redirect_packet( &args );
                                if ( result == XDP_DROP )
                                {
                                    relay_printf( "session pong packet redirect address is not in whitelist" );
                                    INCREMENT_COUNTER( RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST );
                                    INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                                    ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                                    return XDP_DROP;
                                }

                                INCREMENT_COUNTER( RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP );
                                INCREMENT_COUNTER( RELAY_COUNTER_PACKETS_SENT );
                                ADD_COUNTER( RELAY_COUNTER_BYTES_SENT, data_end - data );

                                __u64 whitelist_expire_timestamp = whitelist->expire_timestamp;

                                __sync_bool_compare_and_swap( &whitelist->expire_timestamp, whitelist_expire_timestamp, state->current_timestamp + WHITELIST_TIMEOUT );

                                return XDP_TX;
                            }
                            break;
                        }

                        // unknown packet type

                        INCREMENT_COUNTER( RELAY_COUNTER_DROPPED_PACKETS );
                        ADD_COUNTER( RELAY_COUNTER_DROPPED_BYTES, data_end - data );
                        return XDP_DROP;
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
