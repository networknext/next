#include <enet/enet.h>
#include <stdio.h>
#include <signal.h>

const int MaxClients = 32;
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
    if ( enet_initialize() != 0 )
    {
        printf( "error: failed to initialize enet\n" );
        return 1;
    }


    ENetAddress address;
    address.host = ENET_HOST_ANY;
    address.port = 1234;

    ENetHost * server = enet_host_create( &address, MaxClients, MaxChannels, MaxIncomingBandwidth, MaxOutgoingBandwidth );

    if ( server == NULL )
    {   
        printf( "failed to create enet server\n" );
        return 1;
    }

    printf( "started server on %x:%d\n", address.host, address.port );

    while ( true )
    {
        ENetEvent event;
    
        while ( enet_host_service( server, &event, 1000 ) > 0 )
        {
            switch (event.type)
            {
                case ENET_EVENT_TYPE_CONNECT:
                    printf( "client connected from %x:%u\n", event.peer->address.host, event.peer->address.port );
                    break;

                case ENET_EVENT_TYPE_RECEIVE:
                    printf( "packet of length %u was received from %x:%u on channel %u\n", int(event.packet->dataLength), event.peer->address.host, event.peer->address.port, event.channelID );
                    enet_packet_destroy( event.packet );
                    break;
           
                case ENET_EVENT_TYPE_DISCONNECT:
                    printf( "client disconnected from %x:%u\n", event.peer->address.host, event.peer->address.port );
                    break;

                default:
                    break;
            }
        }
    }

    enet_host_destroy( server );

    enet_deinitialize();

    return 0;
}