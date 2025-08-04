/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_CONFIG_H
#define NEXT_CONFIG_H

#include "next.h"
#include "next_constants.h"

#define NEXT_PROD_AUTODETECT_URL "https://autodetect.virtualgo.net"
#define NEXT_PROD_SERVER_BACKEND_HOSTNAME "server.virtualgo.net"
#define NEXT_PROD_SERVER_BACKEND_PUBLIC_KEY "3UISqg8chLDHu4BKFc+3lr1elIha0IMI2c8vKpwzpbA="
#define NEXT_PROD_RELAY_BACKEND_PUBLIC_KEY "TINP/TnYY/0W7JvLFlYGrB0MUw+b4aIrN20Vq7g5bhU="

#define NEXT_DEV_AUTODETECT_URL "https://autodetect-dev.virtualgo.net"
#define NEXT_DEV_SERVER_BACKEND_HOSTNAME "server-dev.virtualgo.net"
#define NEXT_DEV_SERVER_BACKEND_PUBLIC_KEY "0dvRVqU+krtetlEosEdPN+IxVsNqi7/+Hi6gVjSwSl0="
#define NEXT_DEV_RELAY_BACKEND_PUBLIC_KEY "Z+9puZkCkV03nm4yO49ySF+H181jAlWVy7JPGMlk10I="

#if !NEXT_DEVELOPMENT

#define NEXT_AUTODETECT_URL            NEXT_PROD_AUTODETECT_URL
#define NEXT_SERVER_BACKEND_HOSTNAME   NEXT_PROD_SERVER_BACKEND_HOSTNAME
#define NEXT_SERVER_BACKEND_PUBLIC_KEY NEXT_PROD_SERVER_BACKEND_PUBLIC_KEY
#define NEXT_RELAY_BACKEND_PUBLIC_KEY  NEXT_PROD_RELAY_BACKEND_PUBLIC_KEY

#else // #if !NEXT_DEVELOPMENT

#define NEXT_AUTODETECT_URL            NEXT_DEV_AUTODETECT_URL
#define NEXT_SERVER_BACKEND_HOSTNAME   NEXT_DEV_SERVER_BACKEND_HOSTNAME
#define NEXT_SERVER_BACKEND_PUBLIC_KEY NEXT_DEV_SERVER_BACKEND_PUBLIC_KEY
#define NEXT_RELAY_BACKEND_PUBLIC_KEY  NEXT_DEV_RELAY_BACKEND_PUBLIC_KEY

#endif // #if !NEXT_DEVELOPMENT

#define NEXT_CONFIG_BUCKET_NAME "sloclap_network_next_sdk_config"

extern uint8_t next_server_backend_public_key[];

extern uint8_t next_relay_backend_public_key[];

#endif // #ifndef NEXT_CONFIG_H
