/*
    Network Next SDK. Copyright © 2017 - 2023 Network Next, Inc.

    Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following
    conditions are met:

    1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

    2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions
       and the following disclaimer in the documentation and/or other materials provided with the distribution.

    3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote
       products derived from this software without specific prior written permission.

    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
    INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
    IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
    CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
    OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
    NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

#include "next.h"
#include <time.h>
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <inttypes.h>
#ifndef _WIN32
#include <unistd.h>
#include <net/if.h>
#include <netinet/in.h>
#include <sys/ioctl.h>
#include <sys/syscall.h>
#endif

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void client_packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void) client; (void) context; (void) packet_data; (void) packet_bytes; (void) from;
}

const char * server_addresses[] = {
   "45.76.24.216:50000",       // vultr.chicago (dev test server)
    0
};

uint64_t raspberry_user_id()
{
    uint64_t user_id = 0;

#if defined(__linux__)

    struct ifreq ifr;
    struct ifconf ifc;
    char buf[1024];
    int success = 0;

    int sock = socket(AF_INET, SOCK_DGRAM, IPPROTO_IP);
    if (sock == -1) { /* handle error*/ };

    ifc.ifc_len = sizeof(buf);
    ifc.ifc_buf = buf;
    if (ioctl(sock, SIOCGIFCONF, &ifc) == -1) { /* handle error */ }

    struct ifreq* it = ifc.ifc_req;
    const struct ifreq* const end = it + (ifc.ifc_len / sizeof(struct ifreq));

    for (; it != end; ++it) {
        strcpy(ifr.ifr_name, it->ifr_name);
        if (ioctl(sock, SIOCGIFFLAGS, &ifr) == 0) {
            if (! (ifr.ifr_flags & IFF_LOOPBACK)) { // don't count loopback
                if (ioctl(sock, SIOCGIFHWADDR, &ifr) == 0) {
                    success = 1;
                    break;
                }
            }
        }
        else { /* handle error */ }
    }

    char * mac_address = (char*) &user_id;
    if ( success )
    {
        memcpy( mac_address, ifr.ifr_hwaddr.sa_data, 6 );
        return user_id;
    }

#endif // #ifdef __linux__

    char data[8];
    srand(time(NULL));
    for ( int i = 0; i < 8; i++ )
    {
    	data[i] = rand() % 256;
    }
    memcpy((char*)&user_id, data, 8);

    return user_id;
}

int main()
{
    printf( "\nRaspberry Pi Client\n\n" );

    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_config_t config;
    next_default_config(&config);
#ifdef _WIN32
    strncpy_s(config.customer_public_key, "UoFYERKJnCt18mU53IsWzlEXD2pYD9yd+TiZiq9+cMF9cHG4kMwRtw==", 256);
#else
    strncpy(config.customer_public_key, "UoFYERKJnCt18mU53IsWzlEXD2pYD9yd+TiZiq9+cMF9cHG4kMwRtw==", 256);
#endif
    config.customer_public_key[255] = 0;

    next_init( NULL, &config );

    uint64_t user_id = raspberry_user_id();

    if ( user_id != 0 )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "user id is: %" PRIu64, user_id );
        srand( (unsigned int) user_id );
    }
    else
    {
        srand( (unsigned int) time(NULL) );
    }

    while ( !quit )
    {
	    next_client_t * client = next_client_create( NULL, "0.0.0.0:0", client_packet_received );
	    if ( client == NULL )
	    {
	        printf( "error: failed to create client\n" );
	        return 1;
	    }

	    int server_count = 0;
	    while ( server_addresses[server_count] )
	    {
	        server_count++;
	    }

	    next_client_open_session( client, server_addresses[rand() % server_count] );

	    uint8_t packet_data[8];
	    memcpy( packet_data, &user_id, 8 );

	    double connect_time = next_time();
	    double game_length = 300;

	    while ( !quit )
	    {
	        next_client_send_packet( client, packet_data, sizeof( packet_data ) );

	        next_client_update( client );

	        if ( next_time() - connect_time > game_length )
	        	break;

	        next_sleep( 1.0f );
	    }

	    next_client_destroy( client );
	}

	next_term();

    return 0;
}
