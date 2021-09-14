#include <enet/enet.h>
#include <stdio.h>

int main( int argc, char ** argv ) 
{
    if ( enet_initialize() != 0 )
    {
        printf( "failed to initialize enet\n" );
        return 1;
    }

    printf( "yay enet\n" );

    enet_deinitialize();

    return 0;
}