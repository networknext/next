#include <enet.h>
#include <next.h>
#include <stdio.h>
#include <string.h>
#include <signal.h>
#include <next.h>

const int MaxChannels = 2;
const int MaxIncomingBandwidth = 0;
const int MaxOutgoingBandwidth = 0;

const char * bind_address = "0.0.0.0:30000";
const char * server_address = "127.0.0.1:50000";
const char * customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==";

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
    strncpy( config.customer_public_key, customer_public_key, sizeof(config.customer_public_key) - 1 );

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

    ENetHost * client = enet_host_create( NULL, 1, MaxChannels, MaxIncomingBandwidth, MaxOutgoingBandwidth ); 

    if ( client == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "failed to create enet client" );
        return 1;
    }

    ENetAddress address;
    enet_address_set_host( &address, "localhost" );
    address.port = 1234;

    ENetPeer * peer = enet_host_connect( client, &address, MaxChannels, 0 );    
  
    if ( peer == NULL )
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "could not create client peer" );
        return 1;
    }

    next_printf( NEXT_LOG_LEVEL_INFO, "client connecting to server %x:%d", address.host, address.port );

    ENetEvent event;
       
    if ( enet_host_service( client, &event, 5000 ) > 0 && event.type == ENET_EVENT_TYPE_CONNECT )
    {
        next_printf( NEXT_LOG_LEVEL_INFO, "client connected to server" );
    }
    else
    {
        next_printf( NEXT_LOG_LEVEL_ERROR, "client could not connect to server" );
        return 1;
    }

    while ( true )
    {
        while ( enet_host_service( client, &event, 0 ) > 0 )
        {
            switch ( event.type )
            {
                case ENET_EVENT_TYPE_RECEIVE:
                    if ( event.packet->dataLength == 13 && strcmp( (const char*) event.packet->data, "how are you?" ) == 0 )
                    {
                        next_printf( NEXT_LOG_LEVEL_INFO, "client received packet from server on channel %u", event.channelID );
                    }
                    enet_packet_destroy( event.packet );
                    break;
           
                default:
                    break;
            }
        }

        ENetPacket * packet = enet_packet_create( "hello", strlen("hello") + 1, 0 );
        enet_peer_send( peer, 0, packet );
        enet_host_flush( client );

        next_sleep( 1.0f );
    }

    printf( "\n" );
    fflush( stdout );

    next_printf( NEXT_LOG_LEVEL_INFO, "shutting down..." );

    enet_peer_reset( peer );

    enet_host_destroy( client );

    enet_deinitialize();

    next_term();

    next_printf( NEXT_LOG_LEVEL_INFO, "done" );

    return 0;
}
