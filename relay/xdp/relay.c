/*
    Network Next XDP Relay (userspace)
*/

#include "relay.h"
#include "relay_version.h"
#include "relay_platform.h"
#include "relay_main.h"
#include "relay_ping.h"
#include "relay_bpf.h"
#include "relay_config.h"
#include "relay_debug.h"

#include <memory.h>
#include <stdio.h>
#include <sodium.h>
#include <signal.h>

static struct config_t config;
static struct bpf_t bpf;
#if RELAY_DEBUG
static struct debug_t debug;
#else // #if RELAY_DEBUG
static struct main_t main_data;
static struct ping_t ping;
#endif // #if RELAY_DEBUG

volatile bool quit;
volatile bool relay_clean_shutdown = false;

void interrupt_handler( int signal )
{
    (void) signal; quit = true;
}

void clean_shutdown_handler( int signal )
{
    (void) signal;
    relay_clean_shutdown = true;
    quit = true;
}

static void cleanup()
{
#if RELAY_DEBUG
    debug_shutdown( &debug );
#else // #if RELAY_DEBUG
    ping_shutdown( &ping );
    main_shutdown( &main_data );
    bpf_shutdown( &bpf );
#endif // #if RELAY_DEBUG
    fflush( stdout );
}

int main( int argc, char *argv[] )
{
    char relay_version[RELAY_VERSION_LENGTH];
    snprintf( relay_version, RELAY_VERSION_LENGTH, "xdp-relay-%s", RELAY_VERSION );
    if ( argc == 2 && strcmp(argv[1], "version" ) == 0 ) 
    {
        printf( "%s\n", relay_version );
        fflush( stdout );
        exit(0);
    }

    relay_platform_init();

    printf( "[network next relay]\n" );

    signal( SIGINT,  interrupt_handler );
    signal( SIGTERM, clean_shutdown_handler );
    signal( SIGHUP,  clean_shutdown_handler );

    if ( read_config( &config ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

    if ( bpf_init( &bpf, config.relay_public_address ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

#if RELAY_DEBUG

    // debug relay

    if ( debug_init( &debug, &config, &bpf ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

    int result = debug_run( &debug );

#else // #if RELAY_DEBUG

    // regular relay

    if ( main_init( &main_data, &config, &bpf, relay_version ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

    if ( ping_init( &ping, &config, &main_data, &bpf ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

    if ( ping_start_thread( &ping ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

    int result = main_run( &main_data );

    ping_join_thread( &ping );

#endif // #if RELAY_DEBUG

    cleanup();

    return result;
}
