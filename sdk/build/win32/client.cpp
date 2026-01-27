/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"
#include "next_tests.h"
#include "next_platform.h"
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>

const char* server_address = "35.232.190.226:30000";

const char * buyer_public_key = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ==";

static volatile int quit = 0;

void interrupt_handler( int signal )
{
    (void) signal; quit = 1;
}

void client_packet_received( next_client_t * client, void * context, const next_address_t * from, const uint8_t * packet_data, int packet_bytes )
{
    (void)client; (void)context; (void)from; (void)packet_data; (void)packet_bytes;
}

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

    fflush(stdout);

    printf("\nAll tests completed successfully!\n\n");

    next_term();

    printf("Starting client...\n\n");

    next_log_level(NEXT_LOG_LEVEL_INFO);

    signal(SIGINT, interrupt_handler); signal(SIGTERM, interrupt_handler);

    next_config_t config;
    next_default_config(&config);
    strncpy_s(config.buyer_public_key, buyer_public_key, sizeof(config.buyer_public_key) - 1);

    next_init(NULL, &config);

    next_client_t* client = next_client_create(NULL, "0.0.0.0:0", client_packet_received);
    if (client == NULL)
    {
        printf("error: failed to create client\n");
        return 1;
    }

    next_client_open_session(client, server_address);

    uint8_t packet_data[32];
    memset(packet_data, 0, sizeof(packet_data));

    while (!quit)
    {
        next_client_update(client);

        next_client_send_packet(client, packet_data, sizeof(packet_data));

        next_platform_sleep(1.0 / 60.0);
    }

    next_client_destroy(client);

    next_term();

    return 0;
}
