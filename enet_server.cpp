#include <enet/enet.h>
#include <stdio.h>

const int MaxClients = 32;
const int MaxChannels = 2;
const int MaxIncomingBandwidth = 0;
const int MaxOutgoingBandwidth = 0;

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

    // ...

    printf( "yey server\n" );

    enet_host_destroy( server );

    enet_deinitialize();

    return 0;
}