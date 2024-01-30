
// Network Next PS4 Testbed

#include "next.h"
#include "next_tests.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <kernel.h>

unsigned int sceLibcHeapExtendedAlloc = 1;

size_t sceLibcHeapSize = SCE_LIBC_HEAP_SIZE_EXTENDED_ALLOC_NO_LIMIT;

int32_t main( int argc, const char * const argv[] )
{
    next_log_level( NEXT_LOG_LEVEL_NONE );

    next_config_t config;
    next_default_config( &config );

    next_init( NULL, &config );

    printf( "\nRunning tests...\n\n" );

    next_run_tests();

    printf( "\nAll tests passed successfully!\n\n" );

    return 0;
}
