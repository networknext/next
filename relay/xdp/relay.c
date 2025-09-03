/*
    Network Next XDP Relay.

    Runs on Ubuntu 22.04 LTS 64bit with Linux Kernel 6.5 *ONLY*

    Setup:

        sudo apt install -y build-essential libsodium-dev libcurl4-openssl-dev clang linux-headers-generic linux-headers-`uname -r` unzip libc6-dev-i386 gcc-12 dwarves libelf-dev pkg-config m4 libpcap-dev net-tools

        sudo cp /sys/kernel/btf/vmlinux /usr/lib/modules/`uname -r`/build/

        wget https://github.com/xdp-project/xdp-tools/releases/download/v1.4.2/xdp-tools-1.4.2.tar.gz
        tar -zxf xdp-tools-1.4.2.tar.gz
        cd xdp-tools-1.4.2
        ./configure
        make -j && sudo make install
        cd lib/libbpf/src
        make -j && sudo make install
*/

#include "relay.h"
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
    relay_platform_init();

    printf( "Network Next Relay (release)\n" );

    printf( "(http3)\n" );

    fflush( stdout );

    signal( SIGINT,  interrupt_handler );
    signal( SIGTERM, clean_shutdown_handler );
    signal( SIGHUP,  clean_shutdown_handler );

    printf( "Reading config\n" );

    fflush( stdout );

    if ( read_config( &config ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

    fflush( stdout );

    printf( "Initializing BPF\n" );

    fflush( stdout );

    if ( bpf_init( &bpf, config.relay_public_address, config.relay_internal_address ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

    fflush( stdout );

#if RELAY_DEBUG

    // debug relay

    printf( "Starting debug relay\n" );

    fflush( stdout );

    if ( debug_init( &debug, &config, &bpf ) != RELAY_OK )
    {
        cleanup();
        return 1;
    }

    fflush( stdout );

    int result = debug_run( &debug );

#else // #if RELAY_DEBUG

    printf( "Starting relay\n" );

    fflush( stdout );

    if ( main_init( &main_data, &config, &bpf ) != RELAY_OK )
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
