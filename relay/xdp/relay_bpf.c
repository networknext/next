/*
    Network Next XDP Relay
*/

#include "relay_bpf.h"

#ifdef COMPILE_WITH_BPF

#include <stdio.h>
#include <unistd.h>
#include <ifaddrs.h>
#include <arpa/inet.h>
#include <net/if.h>
#include <errno.h>
#include <stdlib.h>

#include "relay_xdp_source.h"

int bpf_init( struct bpf_t * bpf, uint32_t relay_public_address, uint32_t relay_internal_address )
{
    // we can only run xdp programs as root

    if ( geteuid() != 0 ) 
    {
        printf( "\nerror: this program must be run as root\n\n" );
        return RELAY_ERROR;
    }

    // find the network interface that matches the relay public address *or* relay private address

    char network_interface_name[1024];
    memset( network_interface_name, 0, sizeof(network_interface_name) );
    {
        bool found = false;

        struct ifaddrs * addrs;
        if ( getifaddrs( &addrs ) != 0 )
        {
            printf( "\nerror: getifaddrs failed\n\n" );
            return RELAY_ERROR;
        }

        for ( struct ifaddrs * iap = addrs; iap != NULL; iap = iap->ifa_next ) 
        {
            if ( iap->ifa_addr && ( iap->ifa_flags & IFF_UP ) && iap->ifa_addr->sa_family == AF_INET )
            {
                struct sockaddr_in * sa = (struct sockaddr_in*) iap->ifa_addr;
                if ( ntohl( sa->sin_addr.s_addr ) == relay_public_address || ntohl( sa->sin_addr.s_addr ) == relay_internal_address )
                {
                    strncpy( network_interface_name, iap->ifa_name, sizeof(network_interface_name) );
                    printf( "found network interface: '%s'\n", network_interface_name );
                    bpf->interface_index = if_nametoindex( iap->ifa_name );
                    if ( !bpf->interface_index ) 
                    {
                        printf( "\nerror: if_nametoindex failed\n\n" );
                        return RELAY_ERROR;
                    }
                    found = true;
                    break;
                }
            }
        }

        freeifaddrs( addrs );

        if ( !found )
        {
            printf( "\nerror: could not find any network interface matching relay public address" );
            return RELAY_ERROR;
        }
    }

    // are we in AWS?

    bool running_in_aws = false;
    {
        printf( "Checking if we are running in AWS...\n" );
        char command_line[2048];
        strncpy( command_line, "curl -s \"http://169.254.169.254/latest/meta-data\" --max-time 2 -s 2>/dev/null", sizeof(command_line) );
        printf( "command line: '%s'\n", command_line );
        FILE * file = popen( command_line, "r" );
        if ( file )
        {
            char buffer[1024];
            while ( fgets( buffer, sizeof(buffer), file ) != NULL )
            {
                if ( strstr( buffer, "ami-id" ) != NULL )
                {
                    printf( "Detected that we are running in AWS\n" );
                    running_in_aws = true;
                    break;
                }
            }
            pclose( file );
        }
        if ( !running_in_aws )
        {
            printf( "We are not running in AWS\n" );
        }
        fflush( stdout );
    }

    // we need to set an MTU of 1500 in AWS, otherwise we can't attach the XDP program

    if ( running_in_aws )
    {
        printf( "Setting MTU 1500\n" );
        char command[2048];
        snprintf( command, sizeof(command), "ifconfig %s mtu 1500 up", (const char*) &network_interface_name[0] );
        FILE * file = popen( command, "r" );
        char buffer[1024];
        while ( fgets( buffer, sizeof(buffer), file ) != NULL )
        {
            if ( strlen( buffer ) > 0 )
            {
                printf( "%s", buffer );
            }
        }
        pclose( file );
        fflush( stdout );
    }

    // we need to only use half the network queues available on the NIC

    if ( running_in_aws )
    {
        printf( "AWS workaround for NIC queues\n" );
        fflush( stdout );

        // first we need to find how many combined queues we have

        int max_queues = 2;
        {
            char command[2048];
            snprintf( command, sizeof(command), "ethtool -l %s | grep -m 1 Combined |  perl -ne '/Combined:\\s*(\\d+)/ and print \"$1\\n\";'", (const char*) &network_interface_name[0] );
            FILE * file = popen( command, "r" );
            char buffer[1024];
            while ( fgets( buffer, sizeof(buffer), file ) != NULL )
            {
                if ( strlen( buffer ) > 0 )
                {
                    int result = atoi( buffer );
                    if ( result > 0 )
                    {
                        max_queues = result;
                        printf( "Maximum NIC combined queues is %d\n", max_queues );
                    }
                    break;
                }
            }
            pclose( file );
            fflush( stdout );
        }

        // now reduce to use only half max queues

        int num_queues = max_queues / 2;

        printf( "Setting NIC combined queues to %d\n", num_queues );

        char command[2048];
        snprintf( command, sizeof(command), "ethtool -L %s combined %d", (const char*) &network_interface_name[0], num_queues );
        FILE * file = popen( command, "r" );
        char buffer[1024];
        while ( fgets( buffer, sizeof(buffer), file ) != NULL ) {}
        pclose( file );

        fflush( stdout );
    }

    // be extra safe and let's make sure no xdp programs are running on this interface before we start
    {
        char command[2048];
        snprintf( command, sizeof(command), "xdp-loader unload %s --all", network_interface_name );
        FILE * file = popen( command, "r" );
        char buffer[1024];
        while ( fgets( buffer, sizeof(buffer), file ) != NULL ) {}
        pclose( file );
    }

    // delete all bpf maps we use so stale data doesn't stick around
    {
        const char * command = "rm -f /sys/fs/bpf/config_map && rm -f /sys/fs/bpf/state_map && rm -f /sys/fs/bpf/stats_map && rm -f /sys/fs/bpf/relay_map && rm -f /sys/fs/bpf/session_map && rm -rf /sys/fs/bpf/whitelist_map";
        FILE * file = popen( command, "r" );
        char buffer[1024];
        while ( fgets( buffer, sizeof(buffer), file ) != NULL ) {}
        pclose( file );
    }

    // write out source tar.gz for relay_xdp.o
    {
        FILE * file = fopen( "relay_xdp_source.tar.gz", "wb" );
        if ( !file )
        {
            printf( "\nerror: could not open relay_xdp_source.tar.gz for writing" );
            return RELAY_ERROR;
        }

        fwrite( relay_xdp_source_tar_gz, sizeof(relay_xdp_source_tar_gz), 1, file );

        fclose( file );
    }

    // unzip source build relay_xdp.o from source with make
    {
        const char * command = "rm -f Makefile && rm -f *.c && rm -f *.h && rm -f *.o && rm -f Makefile && tar -zxf relay_xdp_source.tar.gz && make relay_xdp.o";
        FILE * file = popen( command, "r" );
        char buffer[1024];
        while ( fgets( buffer, sizeof(buffer), file ) != NULL ) {}
        pclose( file );
    }

    // clean up after ourselves
    {
        const char * command = "rm -f Makefile && rm -f *.c && rm -f *.h && rm -f *.tar.gz";
        FILE * file = popen( command, "r" );
        char buffer[1024];
        while ( fgets( buffer, sizeof(buffer), file ) != NULL ) {}
        pclose( file );
    }

    // load the relay_xdp program and attach it to the network interface

    printf( "loading relay_xdp...\n" );

    fflush( stdout );

    bpf->program = xdp_program__open_file( "relay_xdp.o", "relay_xdp", NULL );
    if ( libxdp_get_error( bpf->program ) ) 
    {
        printf( "\nerror: could not load relay_xdp program\n\n");
        return RELAY_ERROR;
    }

    printf( "relay_xdp loaded successfully.\n" );

    fflush( stdout );

    printf( "attaching relay_xdp to network interface\n" );

    fflush( stdout );

    int ret = xdp_program__attach( bpf->program, bpf->interface_index, XDP_MODE_NATIVE, 0 );
    if ( ret == 0 )
    {
        bpf->attached_native = true;
    } 
    else
    {
        printf( "falling back to skb mode...\n" );
        ret = xdp_program__attach( bpf->program, bpf->interface_index, XDP_MODE_SKB, 0 );
        if ( ret == 0 )
        {
            bpf->attached_skb = true;
        }
        else
        {
            printf( "\nerror: failed to attach relay_xdp program to interface\n\n" );
            return RELAY_ERROR;
        }
    }

    // get file descriptors for maps so we can communicate with the relay_xdp program running in kernel space

    bpf->config_fd = bpf_obj_get( "/sys/fs/bpf/config_map" );
    if ( bpf->config_fd <= 0 )
    {
        printf( "\nerror: could not get relay config: %s\n\n", strerror(errno) );
        return RELAY_ERROR;
    }

    bpf->state_fd = bpf_obj_get( "/sys/fs/bpf/state_map" );
    if ( bpf->state_fd <= 0 )
    {
        printf( "\nerror: could not get relay state: %s\n\n", strerror(errno) );
        return RELAY_ERROR;
    }

    bpf->stats_fd = bpf_obj_get( "/sys/fs/bpf/stats_map" );
    if ( bpf->stats_fd <= 0 )
    {
        printf( "\nerror: could not get relay stats: %s\n\n", strerror(errno) );
        return RELAY_ERROR;
    }

    bpf->relay_map_fd = bpf_obj_get( "/sys/fs/bpf/relay_map" );
    if ( bpf->relay_map_fd <= 0 )
    {
        printf( "\nerror: could not get relay map: %s\n\n", strerror(errno) );
        return RELAY_ERROR;
    }

    bpf->session_map_fd = bpf_obj_get( "/sys/fs/bpf/session_map" );
    if ( bpf->session_map_fd <= 0 )
    {
        printf( "\nerror: could not get session map: %s\n\n", strerror(errno) );
        return RELAY_ERROR;
    }

    bpf->whitelist_map_fd = bpf_obj_get( "/sys/fs/bpf/whitelist_map" );
    if ( bpf->whitelist_map_fd <= 0 )
    {
        printf( "\nerror: could not get whitelist map: %s\n\n", strerror(errno) );
        return RELAY_ERROR;
    }

    return RELAY_OK;
}

void bpf_shutdown( struct bpf_t * bpf )
{
    assert( bpf );

    if ( bpf->program != NULL )
    {
        if ( bpf->attached_native )
        {
            xdp_program__detach( bpf->program, bpf->interface_index, XDP_MODE_NATIVE, 0 );
        }
        if ( bpf->attached_skb )
        {
            xdp_program__detach( bpf->program, bpf->interface_index, XDP_MODE_SKB, 0 );
        }
        xdp_program__close( bpf->program );
    }
}

#else // #ifdef COMPILE_WITH_BPF

int bpf_init( struct bpf_t * bpf, uint32_t relay_public_address )
{
    return RELAY_OK;
}

void bpf_shutdown( struct bpf_t * bpf )
{
    // ...
}

#endif // #ifdef COMPILE_WITH_BPF
