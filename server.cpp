#include <enet.h>
#include <next.h>
#include <stdio.h>
#include <string.h>
#include <signal.h>

const int MaxClients = 32;
const int MaxChannels = 2;
const int MaxIncomingBandwidth = 0;
const int MaxOutgoingBandwidth = 0;

const char * bind_address = "0.0.0.0:50000";
const char * server_address = "127.0.0.1:50000";
const char * server_datacenter = "local";
const char * server_backend_hostname = "prod.spacecats.net";
const char * customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn";

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

int main( int argc, char ** argv ) 
{
    signal( SIGINT, interrupt_handler ); signal( SIGTERM, interrupt_handler );

    next_config_t config;
    next_default_config( &config );
    strncpy( config.server_backend_hostname, server_backend_hostname, sizeof(config.server_backend_hostname) - 1 );
    strncpy( config.customer_private_key, customer_private_key, sizeof(config.customer_private_key) - 1 );

    if ( next_init( NULL, &config ) != NEXT_OK )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "could not initialize network next" );
        return 1;
    }

    if ( enet_initialize() != 0 )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to initialize enet" );
        return 1;
    }

#if ENET_NETWORK_NEXT

    ENetHostConfig host_config;
    host_config.client = 0;
    host_config.bind_address = bind_address;
    host_config.server_address = server_address;
    host_config.server_datacenter = server_datacenter;

    ENetHost * server = enet_host_create( &host_config, MaxClients, MaxChannels, MaxIncomingBandwidth, MaxOutgoingBandwidth );

#else // #if ENET_NETWORK_NEXT

    ENetAddress address;
    address.host = ENET_HOST_ANY;
    address.port = 50000;

    ENetHost * server = enet_host_create( &address, MaxClients, MaxChannels, MaxIncomingBandwidth, MaxOutgoingBandwidth );

#endif // #if ENET_NETWORK_NEXT

    if ( server == NULL )
    {   
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to create enet server" );
        return 1;
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "started server on port %d", server->address.port );

    while ( !quit )
    {
        ENetEvent event;
    
        while ( enet_host_service( server, &event, 0 ) > 0 )
        {
            switch ( event.type )
            {
                case ENET_EVENT_TYPE_CONNECT:
                {
                    next_printf( NEXT_LOG_LEVEL_INFO, "client connected from %x:%u", event.peer->address.host, event.peer->address.port );
                    const char * user_id = "put your own unique user id in this string and we will hash it to create an anonymous user hash";
                    next_address_t client_address;
       	            enet_address_to_next( &event.peer->address, &client_address );
					next_server_upgrade_session( server->next_server, &client_address, user_id );
				}
                break;

                case ENET_EVENT_TYPE_DISCONNECT:
                    next_printf( NEXT_LOG_LEVEL_INFO, "client disconnected from %x:%u", event.peer->address.host, event.peer->address.port );
                    break;

                case ENET_EVENT_TYPE_RECEIVE:
                    if ( event.packet->dataLength == 6 && strcmp( (const char*) event.packet->data, "hello" ) == 0 )
                    {
                        next_printf( NEXT_LOG_LEVEL_INFO, "received packet from client %x:%u on channel %u", event.peer->address.host, event.peer->address.port, event.channelID );
                        ENetPacket * packet = enet_packet_create( "how are you?", strlen("how are you?") + 1, 0 );
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

    printf( "\n" );
    fflush( stdout );

    next_printf( NEXT_LOG_LEVEL_INFO, "shutting down..." );

    enet_host_destroy( server );

    enet_deinitialize();

    next_term();

    next_printf( NEXT_LOG_LEVEL_INFO, "done" );

    return 0;
}
