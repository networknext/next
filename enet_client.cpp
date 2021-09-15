#include <enet/enet.h>
#include <stdio.h>
#include <signal.h>
#include <next.h>

const int MaxChannels = 2;
const int MaxIncomingBandwidth = 0;
const int MaxOutgoingBandwidth = 0;

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

int main( int argc, char ** argv ) 
{
    if ( next_init( NULL, NULL ) != NEXT_OK )
    {
        printf( "error: could not initialize network next\n" );
        return 1;
    }

    if ( enet_initialize() != 0 )
    {
        printf( "error: failed to initialize enet\n" );
        return 1;
    }

    ENetHost * client = enet_host_create( NULL, 1, MaxChannels, MaxIncomingBandwidth, MaxOutgoingBandwidth ); 

    if ( client == NULL )
    {
        printf( "error: failed to create enet client\n" );
        return 1;
    }

    ENetAddress address;
    enet_address_set_host( &address, "localhost" );
    address.port = 1234;

    ENetPeer * peer = enet_host_connect( client, &address, MaxChannels, 0 );    
  
    if ( peer == NULL )
    {
        printf( "error: could not create client peer\n" );
        return 1;
    }

    ENetEvent event;
       
    if ( enet_host_service( client, &event, 5000) > 0 && event.type == ENET_EVENT_TYPE_CONNECT )
    {
        printf( "client connected to server\n" );
    }
    else
    {
        printf( "error: client could not connect to server\n" );
        return 1;
    }

    while ( true )
    {
        while ( enet_host_service( client, &event, 100 ) > 0 )
        {
            switch ( event.type )
            {
                case ENET_EVENT_TYPE_RECEIVE:
                    printf( "packet of length %u was received from server on channel %u\n", int(event.packet->dataLength), event.channelID );
                    enet_packet_destroy( event.packet );
                    break;
           
                default:
                    break;
            }
        }

        // todo: sleep
    }

    enet_peer_reset( peer );

    enet_host_destroy( client );

    enet_deinitialize();

    next_term();

    return 0;
}
