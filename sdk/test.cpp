/*
    Network Next. Copyright © 2017 - 2025 Network Next, Inc.
    
    Licensed under the Network Next Source Available License 1.0

    If you use this software with a game, you must add this to your credits:

    "This game uses Network Next (networknext.com)"
*/

#include "next.h"
#include "next_tests.h"

#include <stdio.h>
#include <string.h>

int main()
{
    next_quiet( true );

    if ( next_init( NULL, NULL ) != NEXT_OK )
    {
        printf( "error: failed to initialize network next\n" );
    }

    printf( "\nRunning SDK tests:\n\n" );

    next_run_tests();

    next_term();

    printf( "\n" );

    fflush( stdout );

    return 0;
}
