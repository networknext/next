#include <enet.h>
#include <next.h>
#include <stdio.h>
#include <string.h>
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

    printf( "started server on port %d\n", address.port );

    while ( true )
    {
        ENetEvent event;
    
        while ( enet_host_service( server, &event, 1000 ) > 0 )
        {
            switch ( event.type )
            {
                case ENET_EVENT_TYPE_CONNECT:
                    printf( "client connected from %x:%u\n", event.peer->address.host, event.peer->address.port );
                    break;

                case ENET_EVENT_TYPE_DISCONNECT:
                    printf( "client disconnected from %x:%u\n", event.peer->address.host, event.peer->address.port );
                    break;

                case ENET_EVENT_TYPE_RECEIVE:
                    if ( event.packet->dataLength == 6 && strcmp( (const char*) event.packet->data, "hello" ) == 0 )
                    {
                        printf( "received packet from client %x:%u on channel %u\n", event.peer->address.host, event.peer->address.port, event.channelID );
                        ENetPacket * packet = enet_packet_create( "how are you?", strlen ("how are you?") + 1, 0 );
                        enet_peer_send( event.peer, 0, packet );
                        enet_host_flush( server );
                    }
                    enet_packet_destroy( event.packet );
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
