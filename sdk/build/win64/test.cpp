/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"
#include "next_tests.h"
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>

int main()
{
    printf("\nRunning tests...\n\n");

    next_log_level(NEXT_LOG_LEVEL_NONE);

    if (next_init(NULL, NULL) != NEXT_OK)
    {
        printf("error: failed to initialize network next\n");
    }

    next_log_level(NEXT_LOG_LEVEL_NONE);

    next_run_tests();

    next_term();

    fflush(stdout);

    printf("\nAll tests completed successfully!\n\n");

    return 0;
}
