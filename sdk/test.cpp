/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
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
