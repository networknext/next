/*
    Network Next XDP Relay
*/

#include "relay_main.h"
#include "relay_queue.h"
#include "relay_encoding.h"
#include "relay_platform.h"
#include "relay_config.h"
#include "relay_shared.h"
#include "relay_bpf.h"

#include <curl/curl.h>
#include <sodium.h>
#include <time.h>
#include <errno.h>
#include <bpf/bpf.h>
#include <bpf/libbpf.h>
#include <inttypes.h>
#include <math.h>

int main_init( struct main_t * main, struct config_t * config, struct bpf_t * bpf )
{
    // initialize curl so we can talk with the relay backend

    main->curl = curl_easy_init();
    if ( !main->curl )
    {
        printf( "\nerror: could not initialize curl\n\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    main->control_queue = relay_queue_create( 64 );
    if ( !main->control_queue )
    {
        printf( "\nerror: could not create control queue\n\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    main->control_mutex = relay_platform_mutex_create();
    if ( !main->control_mutex )
    {
        printf( "\nerror: could not create control mutex\n\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    main->stats_queue = relay_queue_create( 64 );
    if ( !main->stats_queue )
    {
        printf( "\nerror: could not create stats queue\n\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    main->stats_mutex = relay_platform_mutex_create();
    if ( !main->stats_mutex )
    {
        printf( "\nerror: could not create stats mutex\n\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    main->update_response_memory = (uint8_t*) malloc( RELAY_RESPONSE_MAX_BYTES );
    if ( !main->update_response_memory )
    {
        printf( "\nerror: could not allocate update response memory\n\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    main->start_time = time( NULL );
    main->relay_backend_url = config->relay_backend_url;
    main->relay_port = config->relay_port;
    main->relay_public_address = config->relay_public_address;
    main->relay_internal_address = config->relay_internal_address;
    memcpy( main->relay_public_key, config->relay_public_key, sizeof(config->relay_public_key) );
    memcpy( main->relay_private_key, config->relay_private_key, sizeof(config->relay_private_key) );
    memcpy( main->relay_backend_public_key, config->relay_backend_public_key, sizeof(config->relay_backend_public_key) );
#ifdef COMPILE_WITH_BPF
    main->stats_fd = bpf->stats_fd;
    main->state_fd = bpf->state_fd;
    main->session_map_fd = bpf->session_map_fd;
    main->whitelist_map_fd = bpf->whitelist_map_fd;
#endif // #idef COMPILE_WITH_BPF

#ifdef COMPILE_WITH_BPF
    
    // set relay config

    struct relay_config relay_config;

    memset( &relay_config, 0, sizeof(struct relay_config) );

    relay_config.relay_port = htons( config->relay_port );
    relay_config.relay_public_address = htonl( config->relay_public_address );
    relay_config.relay_internal_address = htonl( config->relay_internal_address );
    memcpy( relay_config.relay_secret_key, config->relay_secret_key, RELAY_SECRET_KEY_BYTES );
    memcpy( relay_config.relay_backend_public_key, config->relay_backend_public_key, RELAY_BACKEND_PUBLIC_KEY_BYTES );

    __u32 key = 0;
    int err = bpf_map_update_elem( bpf->config_fd, &key, &relay_config, BPF_ANY );
    if ( err != 0 )
    {
        printf( "\nerror: failed to set relay config: %s\n\n", strerror(errno) );
        return RELAY_ERROR;
    }

#endif // #ifdef COMPILE_WITH_BPF

    return RELAY_OK;
}

int main_update( struct main_t * main );

extern bool quit;
extern bool relay_clean_shutdown;

int main_run( struct main_t * main )
{
    printf( "Starting main thread\n" );

    fflush( stdout );

    bool aborted = false;

    int update_attempts = 0;

    while ( !quit )
    {
        if ( main_update( main ) == RELAY_OK )
        {
            update_attempts = 0;
        }
        else
        {
            if ( update_attempts++ >= RELAY_MAX_UPDATE_ATTEMPTS )
            {
                printf( "error: could not update relay %d times in a row. shutting down :(", RELAY_MAX_UPDATE_ATTEMPTS );
                fflush( stdout );
                aborted = true;
                quit = 1;
                break;
            }
        }

        relay_platform_sleep( 1.0 );
    }

    if ( relay_clean_shutdown )
    {
        printf( "\nClean shutdown...\n" );

        fflush( stdout );

        main->shutting_down = true;

        uint seconds = 0;
        while ( seconds <= 60 && main_update( main ) == RELAY_OK )
        {
            printf( "Shutting down in %d seconds\n", 60 - seconds );
            fflush( stdout );
            relay_platform_sleep( 1.0 );
            seconds++;
        }

        if ( seconds < 60 )
        {
            printf( "Sleeping for extra 30 seconds for safety...\n" );
            fflush( stdout );
            relay_platform_sleep( 30.0 );
        }

        printf( "Clean shutdown completed\n" );

        fflush( stdout );        
    }
    else
    {
        printf( "\nHard shutdown!\n" );

        fflush( stdout );        
    }

    return 0;
}

void main_shutdown( struct main_t * main )
{
    if ( main->curl )
    {
        curl_easy_cleanup( main->curl );
    }

    if ( main->update_response_memory )
    {
        free( main->update_response_memory );
    }

    if ( main->stats_queue )
    {
        relay_queue_destroy( main->stats_queue );
    }

    if ( main->stats_mutex )
    {
        relay_platform_mutex_destroy( main->stats_mutex );
    }

    if ( main->control_queue )
    {
        relay_queue_destroy( main->control_queue );
    }

    if ( main->control_mutex )
    {
        relay_platform_mutex_destroy( main->control_mutex );
    }

    memset( main, 0, sizeof(struct main_t) );
}

// -----------------------------------------------------------------------------------------------------------------------------

struct curl_buffer_t
{
    int size;
    int max_size;
    uint8_t * data;
};

size_t curl_buffer_write_function( char * ptr, size_t size, size_t nmemb, void * userdata )
{
    struct curl_buffer_t * buffer = (struct curl_buffer_t*) userdata;
    assert( buffer );
    assert( size == 1 );
    if ( (int) ( buffer->size + size*nmemb ) > buffer->max_size )
        return 0;
    memcpy( buffer->data + buffer->size, ptr, size*nmemb );
    buffer->size += size * nmemb;
    return size * nmemb;
}

void clamp( int * value, int min, int max )
{
    if ( *value > max )
    {
        *value = max;
    } 
    else if ( *value < min )
    {
        *value = min;
    }
}

struct session_stats
{
    uint64_t session_count;
    uint64_t envelope_kbps_up;
    uint64_t envelope_kbps_down;
};

struct session_stats main_update_timeouts( struct main_t * main )
{
    struct session_stats stats;
    memset( &stats, 0, sizeof(struct session_stats) );

    // timeout old sessions in session map
    {
        struct session_key current_key;
        struct session_key next_key;

        int next_key_result = bpf_map_get_next_key( main->session_map_fd, NULL, &next_key );

        uint64_t current_timestamp = main->current_timestamp;

        while ( next_key_result == 0 )
        {
            memcpy( &current_key, &next_key, sizeof(struct session_key) );

            bool timed_out = false;
            struct session_data current_value;
            int result = bpf_map_lookup_elem( main->session_map_fd, &current_key, &current_value );
            if ( result == 0 )
            {
                stats.session_count++;
                stats.envelope_kbps_up += current_value.envelope_kbps_up;
                stats.envelope_kbps_down += current_value.envelope_kbps_down;
                timed_out = current_value.expire_timestamp < current_timestamp;
            }

            next_key_result = bpf_map_get_next_key( main->session_map_fd, &current_key, &next_key );

            if ( timed_out )
            {
                bpf_map_delete_elem( main->session_map_fd, &current_key );
            }
        }
    }

    // timeout old entries in whitelist map
    {
        struct whitelist_key current_key;
        struct whitelist_key next_key;

        int next_key_result = bpf_map_get_next_key( main->whitelist_map_fd, NULL, &next_key );

        uint64_t current_timestamp = main->current_timestamp;

        while ( next_key_result == 0 )
        {
            memcpy( &current_key, &next_key, sizeof(struct whitelist_key) );

            bool timed_out = false;
            struct whitelist_value current_value;
            int result = bpf_map_lookup_elem( main->whitelist_map_fd, &current_key, &current_value );
            if ( result == 0 )
            {
                timed_out = current_value.expire_timestamp < current_timestamp;
            }

            next_key_result = bpf_map_get_next_key( main->whitelist_map_fd, &current_key, &next_key );

            if ( timed_out )
            {
                bpf_map_delete_elem( main->whitelist_map_fd, &current_key );
            }
        }
    }

    return stats;
}

int main_update( struct main_t * main )
{
    // update timeouts

    struct session_stats stats = main_update_timeouts( main );

    // get counters from xdp

    uint64_t counters[RELAY_NUM_COUNTERS];
    memset( &counters, 0, sizeof(counters) );

#ifdef COMPILE_WITH_BPF

    unsigned int num_cpus = libbpf_num_possible_cpus();

    struct relay_stats values[num_cpus];

    int key = 0;
    if ( bpf_map_lookup_elem( main->stats_fd, &key, values ) != 0 ) 
    {
        printf( "error: could not look up relay stats: %s\n", strerror( errno ) );
        fflush( stdout );
        return RELAY_ERROR;
    }

    for ( int i = 0; i < num_cpus; i++ )
    {
        for ( int j = 0; j < RELAY_NUM_COUNTERS; j++ )
        {
            counters[j] += values[i].counters[j];
        }
    }

    counters[RELAY_COUNTER_SESSIONS] = stats.session_count;
    counters[RELAY_COUNTER_ENVELOPE_KBPS_UP] = stats.envelope_kbps_up;
    counters[RELAY_COUNTER_ENVELOPE_KBPS_DOWN] = stats.envelope_kbps_down;

#endif // #ifdef COMPILE_WIH_BPF

    // pump stats messages from ping thread

    while ( true )
    {
        relay_platform_mutex_acquire( main->stats_mutex );
        struct relay_stats_message * message = (struct relay_stats_message*) relay_queue_pop( main->stats_queue );
        relay_platform_mutex_release( main->stats_mutex );

        if ( !message )
            break;

        main->pings_sent = message->pings_sent;
        main->bytes_sent = message->bytes_sent;

        memcpy( &main->ping_stats, &message->ping_stats, sizeof(struct relay_ping_stats_t) );

        free( message );
    }

    counters[RELAY_COUNTER_RELAY_PING_PACKET_SENT] += main->pings_sent;
    counters[RELAY_COUNTER_PACKETS_SENT] += main->pings_sent;
    counters[RELAY_COUNTER_BYTES_SENT] += main->bytes_sent;

    // derived relay statistics from counters

    double current_time = relay_platform_time();

    double time_since_last_update = current_time - main->last_stats_time;

    main->last_stats_time = current_time;

    uint64_t packets_sent_since_last_update = ( counters[RELAY_COUNTER_PACKETS_SENT] > main->last_stats_packets_sent ) ? counters[RELAY_COUNTER_PACKETS_SENT] - main->last_stats_packets_sent : 0;
    uint64_t packets_received_since_last_update = ( counters[RELAY_COUNTER_PACKETS_RECEIVED] > main->last_stats_packets_received ) ? counters[RELAY_COUNTER_PACKETS_RECEIVED] - main->last_stats_packets_received : 0;

    uint64_t bytes_sent_since_last_update = ( counters[RELAY_COUNTER_BYTES_SENT] > main->last_stats_bytes_sent ) ? counters[RELAY_COUNTER_BYTES_SENT] - main->last_stats_bytes_sent : 0;
    uint64_t bytes_received_since_last_update = ( counters[RELAY_COUNTER_BYTES_RECEIVED] > main->last_stats_bytes_received ) ? counters[RELAY_COUNTER_BYTES_RECEIVED] - main->last_stats_bytes_received : 0;

    uint64_t client_pings_since_last_update = ( counters[RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED] > main->last_stats_client_pings_received ) ? counters[RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED] - main->last_stats_client_pings_received : 0;
    uint64_t server_pings_since_last_update = ( counters[RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED] > main->last_stats_server_pings_received ) ? counters[RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED] - main->last_stats_server_pings_received : 0;
    uint64_t relay_pings_since_last_update = ( counters[RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED] > main->last_stats_relay_pings_received ) ? counters[RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED] - main->last_stats_relay_pings_received : 0;

    double packets_sent_per_second = 0.0;
    double packets_received_per_second = 0.0;
    double bytes_sent_per_second = 0.0;
    double bytes_received_per_second = 0.0;
    double client_pings_per_second = 0.0;
    double server_pings_per_second = 0.0;
    double relay_pings_per_second = 0.0;

    if ( time_since_last_update > 0.0 )
    {
        packets_sent_per_second = packets_sent_since_last_update / time_since_last_update;
        packets_received_per_second = packets_received_since_last_update / time_since_last_update;
        bytes_sent_per_second = bytes_sent_since_last_update / time_since_last_update;
        bytes_received_per_second = bytes_received_since_last_update / time_since_last_update;
        client_pings_per_second = client_pings_since_last_update / time_since_last_update;
        server_pings_per_second = server_pings_since_last_update / time_since_last_update;
        relay_pings_per_second = relay_pings_since_last_update / time_since_last_update;
    }

    double bandwidth_sent_kbps = bytes_sent_per_second * 8.0 / 1000.0;
    double bandwidth_received_kbps = bytes_received_per_second * 8.0 / 1000.0;

    main->last_stats_packets_sent = counters[RELAY_COUNTER_PACKETS_SENT];
    main->last_stats_packets_received = counters[RELAY_COUNTER_PACKETS_RECEIVED];
    main->last_stats_bytes_sent = counters[RELAY_COUNTER_BYTES_SENT];
    main->last_stats_bytes_received = counters[RELAY_COUNTER_BYTES_RECEIVED];
    main->last_stats_client_pings_received = counters[RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED];
    main->last_stats_server_pings_received = counters[RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED];
    main->last_stats_relay_pings_received = counters[RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED];

    uint64_t session_count = counters[RELAY_COUNTER_SESSIONS];
    uint64_t envelope_bandwidth_kbps_up = counters[RELAY_COUNTER_ENVELOPE_KBPS_UP];
    uint64_t envelope_bandwidth_kbps_down = counters[RELAY_COUNTER_ENVELOPE_KBPS_DOWN];

    // build relay update data

    uint8_t update_version = 1;

    static uint8_t update_data[10*1024*1024];

    uint8_t * p = update_data;

    relay_write_uint8( &p, update_version );

    relay_write_uint8( &p, RELAY_ADDRESS_IPV4 );
    relay_write_uint32( &p, htonl( main->relay_public_address ) );
    relay_write_uint16( &p, main->relay_port );

    uint8_t * encrypt_buffer = p;

    uint64_t local_timestamp = time( NULL );

    relay_write_uint64( &p, local_timestamp );             // IMPORTANT: local timestamp must not move
    relay_write_uint64( &p, main->start_time );

    relay_write_uint32( &p, main->ping_stats.num_relays );
    for ( int i = 0; i < main->ping_stats.num_relays; ++i )
    {
        relay_write_uint64( &p, main->ping_stats.relay_ids[i] );

        const float rtt = main->ping_stats.relay_rtt[i];
        const float jitter = main->ping_stats.relay_jitter[i];
        const float packet_loss = main->ping_stats.relay_packet_loss[i] / 100.0f * 65535.0f;

        int integer_rtt = (int) ceil( rtt );
        int integer_jitter = (int) ceil( jitter );
        int integer_packet_loss = (int) ceil( packet_loss );

        clamp( &integer_rtt, 0, 255 );
        clamp( &integer_jitter, 0, 255 );
        clamp( &integer_packet_loss, 0, 65535 );
     
        relay_write_uint8( &p, (uint8_t) integer_rtt );
        relay_write_uint8( &p, (uint8_t) integer_jitter );
        relay_write_uint16( &p, (uint16_t) integer_packet_loss );
    }

    relay_write_uint32( &p, (uint32_t) session_count );
    relay_write_uint32( &p, (uint32_t) envelope_bandwidth_kbps_up );
    relay_write_uint32( &p, (uint32_t) envelope_bandwidth_kbps_down );
    relay_write_float32( &p, (float) packets_sent_per_second );
    relay_write_float32( &p, (float) packets_received_per_second );
    relay_write_float32( &p, (float) bandwidth_sent_kbps );
    relay_write_float32( &p, (float) bandwidth_received_kbps );
    relay_write_float32( &p, (float) client_pings_per_second );
    relay_write_float32( &p, (float) server_pings_per_second );
    relay_write_float32( &p, (float) relay_pings_per_second );

    const uint64_t SHUTTING_DOWN = 1;
    uint64_t relay_flags = main->shutting_down ? SHUTTING_DOWN : 0;
    relay_write_uint64( &p, relay_flags );

    relay_write_string( &p, RELAY_VERSION, RELAY_VERSION_LENGTH );

    relay_write_uint32( &p, RELAY_NUM_COUNTERS );
    for ( int i = 0; i < RELAY_NUM_COUNTERS; ++i )
    {
        relay_write_uint64( &p, counters[i] );
    }

    // encrypt data after relay address

    const int encrypt_buffer_length = (int) ( p - encrypt_buffer );

    uint8_t nonce[crypto_box_NONCEBYTES];
    relay_platform_random_bytes( nonce, crypto_box_NONCEBYTES );

    if ( crypto_box_easy( encrypt_buffer, encrypt_buffer, encrypt_buffer_length, nonce, main->relay_backend_public_key, main->relay_private_key ) != 0 )
    {
        printf( "error: failed to encrypt relay update\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }
    
    p += crypto_box_MACBYTES;

    memcpy( p, nonce, crypto_box_NONCEBYTES );

    p += crypto_box_NONCEBYTES;

    const int update_data_length = p - update_data;

    // post relay update to the backend

    struct curl_slist * slist = curl_slist_append( NULL, "Content-Type:application/octet-stream" );

    struct curl_buffer_t update_response_buffer;
    update_response_buffer.size = 0;
    update_response_buffer.max_size = RELAY_RESPONSE_MAX_BYTES;
    update_response_buffer.data = (uint8_t*) main->update_response_memory;

    char update_url[1024];
    snprintf( update_url, sizeof(update_url), "%s/relay_update", main->relay_backend_url );

    curl_easy_setopt( main->curl, CURLOPT_BUFFERSIZE, 10 * 1024 * 1024L );
    curl_easy_setopt( main->curl, CURLOPT_URL, update_url );
    curl_easy_setopt( main->curl, CURLOPT_NOPROGRESS, 1L );
    curl_easy_setopt( main->curl, CURLOPT_POSTFIELDS, update_data );
    curl_easy_setopt( main->curl, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t) update_data_length );
    curl_easy_setopt( main->curl, CURLOPT_HTTPHEADER, slist );
    curl_easy_setopt( main->curl, CURLOPT_USERAGENT, "network next relay" );
    curl_easy_setopt( main->curl, CURLOPT_MAXREDIRS, 50L );
    curl_easy_setopt( main->curl, CURLOPT_HTTP_VERSION, (long)CURL_HTTP_VERSION_2TLS );
    curl_easy_setopt( main->curl, CURLOPT_TIMEOUT_MS, 10000L );
    curl_easy_setopt( main->curl, CURLOPT_WRITEDATA, &update_response_buffer );
    curl_easy_setopt( main->curl, CURLOPT_WRITEFUNCTION, &curl_buffer_write_function );
    curl_easy_setopt( main->curl, CURLOPT_DNS_CACHE_TIMEOUT, (long) -1 ); // IMPORTANT: Perform DNS lookup once and hold that IP address forever. Fixes transient DNS issues making relays go offline!

    CURLcode ret = curl_easy_perform( main->curl );

    curl_slist_free_all( slist );
    slist = NULL;

    if ( ret != CURLE_OK )
    {
        printf( "error: could not post relay update (%s)\n", curl_easy_strerror(ret) );
        fflush( stdout );
        return RELAY_ERROR;
    }

    long code;
    curl_easy_getinfo( main->curl, CURLINFO_RESPONSE_CODE, &code );
    if ( code != 200 )
    {
        printf( "error: relay update response is %d\n", (int)code );
        fflush( stdout );
        return RELAY_ERROR;
    }

    // parse response from relay backend

    curl_off_t response_size;
    curl_easy_getinfo( main->curl, CURLINFO_SIZE_DOWNLOAD_T, &response_size );

    const uint8_t * q = update_response_buffer.data;

    uint8_t version = relay_read_uint8( &q );

    const uint32_t update_response_version = 1;

    if ( version != update_response_version )
    {
        printf( "error: bad relay update response version. expected %d, got %d\n", update_response_version, version );
        fflush( stdout );
        return RELAY_ERROR;
    }

    uint64_t backend_timestamp = relay_read_uint64( &q );

    if ( !main->initialized )
    {
        printf( "Relay initialized\n" );
        fflush( stdout );
        main->initialized = true;
    }

    main->current_timestamp = backend_timestamp;

    int num_relays = relay_read_uint32( &q );

    if ( num_relays > MAX_RELAYS )
    {
        printf( "error: too many relays to ping. max is %d, got %d\n", MAX_RELAYS, num_relays );
        fflush( stdout );
        return RELAY_ERROR;
    }

    struct relay_set relay_ping_set;
    memset( &relay_ping_set, 0, sizeof(relay_ping_set) );
    relay_ping_set.num_relays = num_relays;
    for ( int i = 0; i < num_relays; i++ )
    {
        relay_ping_set.id[i] = relay_read_uint64( &q );
        uint8_t address_type = relay_read_uint8( &q );
        if ( address_type != RELAY_ADDRESS_IPV4 )
        {
            printf( "error: only ipv4 relay addresses are supported\n" );
            fflush( stdout );
            return RELAY_ERROR;
        }
        relay_ping_set.address[i] = ntohl( relay_read_uint32( &q ) );
        relay_ping_set.port[i] = relay_read_uint16( &q );
        relay_ping_set.internal[i] = relay_read_uint8( &q );
    }

    char target_version[RELAY_VERSION_LENGTH];
    relay_read_string( &q, target_version, RELAY_VERSION_LENGTH);

    uint8_t next_magic[8];
    uint8_t current_magic[8];
    uint8_t previous_magic[8];

    relay_read_bytes( &q, next_magic, 8 );
    relay_read_bytes( &q, current_magic, 8 );
    relay_read_bytes( &q, previous_magic, 8 );

    uint32_t expected_public_address;
    uint16_t expected_port;

    relay_read_address( &q, &expected_public_address, &expected_port );
    if ( main->relay_public_address != expected_public_address )
    {
        printf( "error: relay public address mismatch\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    if ( main->relay_port != expected_port )
    {
        printf( "error: relay port mismatch\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    uint8_t expected_has_internal_address = relay_read_uint8( &q );
    if ( expected_has_internal_address )
    {
        uint32_t expected_internal_address;
        relay_read_address( &q, &expected_internal_address, &expected_port );
        if ( main->relay_internal_address != expected_internal_address )
        {
            printf( "error: relay internal address mismatch\n" );
            fflush( stdout );
            return RELAY_ERROR;
        }
    }

    uint8_t expected_relay_public_key[crypto_box_PUBLICKEYBYTES];
    uint8_t expected_relay_backend_public_key[crypto_box_PUBLICKEYBYTES];
    relay_read_bytes( &q, expected_relay_public_key, crypto_box_PUBLICKEYBYTES );
    relay_read_bytes( &q, expected_relay_backend_public_key, crypto_box_PUBLICKEYBYTES );

    if ( memcmp( main->relay_public_key, expected_relay_public_key, crypto_box_PUBLICKEYBYTES ) != 0 )
    {
        printf( "error: relay is misconfigured. relay public key does not match expected value\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    uint8_t dummy[RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES];
    relay_read_bytes( &q, dummy, RELAY_ENCRYPTED_ROUTE_TOKEN_BYTES );

    uint8_t ping_key[RELAY_PING_KEY_BYTES];
    relay_read_bytes( &q, ping_key, RELAY_PING_KEY_BYTES );

#ifdef COMPILE_WITH_BPF

    // update bpf relay state

    struct relay_state state;

    state.current_timestamp = main->current_timestamp;
    memcpy( state.current_magic, current_magic, 8 );
    memcpy( state.previous_magic, previous_magic, 8 );
    memcpy( state.next_magic, next_magic, 8 );
    memcpy( state.ping_key, ping_key, RELAY_PING_KEY_BYTES );
    {
        int key = 0;
        if ( bpf_map_update_elem( main->state_fd, &key, &state, BPF_ANY ) != 0 )
        {
            printf( "error: failed to update relay state\n" );
            fflush( stdout );
            return RELAY_ERROR;
        }
    }

#endif // #ifdef COMPILE_WITH_BPF
    
    // stop if the relay queue is full, we can delta relays later and miss nothing

    relay_platform_mutex_acquire( main->control_mutex );
    bool full = relay_queue_full( main->control_queue );
    relay_platform_mutex_release( main->control_mutex );
    if ( full )
    {
        return RELAY_OK;
    }

    // create a control message for the ping thread

    struct relay_control_message * message = (struct relay_control_message*) malloc( sizeof(struct relay_control_message) );
    if ( !message )
    {
        printf( "error: could not allocate control message\n" );
        fflush( stdout );
        return RELAY_ERROR;
    }

    // find new relays

    message->new_relays.num_relays = 0;
    for ( int i = 0; i < relay_ping_set.num_relays; i++ )
    {
        /*
        if ( !relay_hash_exists( &main->relay_ping_hash, relay_ping_set.id[i] ) )
        */

        bool found = false;
        for ( int j = 0; j < main->relay_ping_set.num_relays; j++ )
        {
            if ( main->relay_ping_set.id[j] == relay_ping_set.id[i] )
            {
                found = true;
                break;
            }
        }

        if ( !found )
        {
            const int index = message->new_relays.num_relays;
            message->new_relays.id      [index] = relay_ping_set.id[i];
            message->new_relays.address [index] = relay_ping_set.address[i];
            message->new_relays.port    [index] = relay_ping_set.port[i];
            message->new_relays.internal[index] = relay_ping_set.internal[i];
            message->new_relays.num_relays++;
        }
    }

    // find relays to delete

    // todo
    /*
    struct relay_hash relay_ping_hash;
    relay_hash_initialize( &relay_ping_hash, (uint64_t*)relay_ping_set.id, relay_ping_set.num_relays );
    */

    message->delete_relays.num_relays = 0;
    for ( int i = 0; i < main->relay_ping_set.num_relays; i++ )
    {
        /*
        relay_hash_exists( &relay_ping_hash, main->relay_ping_set.id[i] );
        */

        bool found = false;
        for ( int j = 0; j < relay_ping_set.num_relays; j++ )
        {
            if ( relay_ping_set.id[j] == main->relay_ping_set.id[i] )
            {
                found = true;
                break;
            }
        }

        if ( !found )
        {
            const int index = message->delete_relays.num_relays;
            message->delete_relays.id      [index] = main->relay_ping_set.id[i];
            message->delete_relays.address [index] = main->relay_ping_set.address[i];
            message->delete_relays.port    [index] = main->relay_ping_set.port[i];
            message->delete_relays.internal[index] = main->relay_ping_set.internal[i];
            message->delete_relays.num_relays++;
        }
    }

    // send the control message to the ping thread

    message->current_timestamp = backend_timestamp;

    memcpy( message->current_magic, current_magic, 8 );

    memcpy( message->ping_key, ping_key, RELAY_PING_KEY_BYTES );

    relay_platform_mutex_acquire( main->control_mutex );
    bool result = relay_queue_push( main->control_queue, message );
    relay_platform_mutex_release( main->control_mutex );
    if ( !result )
    {
        return RELAY_OK;
    }

    // stash relay ping set and hash for next update

    memcpy( &main->relay_ping_set, &relay_ping_set, sizeof(struct relay_set) );
    // todo
    // memcpy( &main->relay_ping_hash, &relay_ping_hash, sizeof(struct relay_hash) );

    return RELAY_OK;
}
