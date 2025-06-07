/*
    Network Next. Copyright © 2017 - 2025 Network Next, Inc.
    
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_CONFIG_H
#define NEXT_CONFIG_H

#include "next.h"
#include "next_constants.h"

#define NEXT_PROD_SERVER_BACKEND_HOSTNAME "server.virtualgo.net"
#define NEXT_PROD_SERVER_BACKEND_PUBLIC_KEY "Uycn3KibCfXJo1uM+NNWgCySRzM2Ti3bhvom9XBkxfE="
#define NEXT_PROD_RELAY_BACKEND_PUBLIC_KEY "unH/Yxm0C6JCZ1dTGZH2BTBOFhGMcYsOEDURd9qY72w="

#define NEXT_DEV_SERVER_BACKEND_HOSTNAME "server-dev.virtualgo.net"
#define NEXT_DEV_SERVER_BACKEND_PUBLIC_KEY "9ShrWyUWRm+twHzd5xJuO3kpH1j9hGuxK2MlfI44x38="
#define NEXT_DEV_RELAY_BACKEND_PUBLIC_KEY "XMZGJsM6tHP2vFBXQZBmXguSNa9Czs8V/pcCfhF2KhU="

#if !NEXT_DEVELOPMENT
#define NEXT_SERVER_BACKEND_HOSTNAME   NEXT_PROD_SERVER_BACKEND_HOSTNAME
#define NEXT_SERVER_BACKEND_PUBLIC_KEY NEXT_PROD_SERVER_BACKEND_PUBLIC_KEY
#define NEXT_RELAY_BACKEND_PUBLIC_KEY  NEXT_PROD_RELAY_BACKEND_PUBLIC_KEY
#else // #if !NEXT_DEVELOPMENT
#define NEXT_SERVER_BACKEND_HOSTNAME   NEXT_DEV_SERVER_BACKEND_HOSTNAME
#define NEXT_SERVER_BACKEND_PUBLIC_KEY NEXT_DEV_SERVER_BACKEND_PUBLIC_KEY
#define NEXT_RELAY_BACKEND_PUBLIC_KEY  NEXT_DEV_RELAY_BACKEND_PUBLIC_KEY
#endif // #if !NEXT_DEVELOPMENT

#define NEXT_CONFIG_BUCKET_NAME "next_network_next_sdk_config"

extern uint8_t next_server_backend_public_key[];

extern uint8_t next_relay_backend_public_key[];

#endif // #ifndef NEXT_CONFIG_H
