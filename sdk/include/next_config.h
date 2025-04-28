/*
    Network Next. Copyright © 2017 - 2025 Network Next, Inc.
    
    Licensed under the Network Next Source Available License 1.0

    If you use this software with a game, you must add this to your credits:

    "This game uses Network Next (networknext.com)"
*/

#ifndef NEXT_CONFIG_H
#define NEXT_CONFIG_H

#include "next.h"
#include "next_constants.h"

#define NEXT_PROD_SERVER_BACKEND_HOSTNAME "server.virtualgo.net"
#define NEXT_PROD_SERVER_BACKEND_PUBLIC_KEY "f3QGkf+Hy5BATBPq+IGzoTSrWVosmQTcBDU8BHEL0z0="
#define NEXT_PROD_RELAY_BACKEND_PUBLIC_KEY "Xg+WVwYouISwX/2h3Slu8knq5W/d+6AID0aF/Vatfg0="

#define NEXT_DEV_SERVER_BACKEND_HOSTNAME "server-dev.virtualgo.net"
#define NEXT_DEV_SERVER_BACKEND_PUBLIC_KEY "FOl98B81AtRxlmisjkL5ROaVaV3bC6v3/hv+wubm6hs="
#define NEXT_DEV_RELAY_BACKEND_PUBLIC_KEY "2UrvlOyXfQk+F3QZhXrP36kecqlLaSo28+eIubVDS2Y="

#if !NEXT_DEVELOPMENT
#define NEXT_SERVER_BACKEND_HOSTNAME   NEXT_PROD_SERVER_BACKEND_HOSTNAME
#define NEXT_SERVER_BACKEND_PUBLIC_KEY NEXT_PROD_SERVER_BACKEND_PUBLIC_KEY
#define NEXT_RELAY_BACKEND_PUBLIC_KEY  NEXT_PROD_RELAY_BACKEND_PUBLIC_KEY
#else // #if !NEXT_DEVELOPMENT
#define NEXT_SERVER_BACKEND_HOSTNAME   NEXT_DEV_SERVER_BACKEND_HOSTNAME
#define NEXT_SERVER_BACKEND_PUBLIC_KEY NEXT_DEV_SERVER_BACKEND_PUBLIC_KEY
#define NEXT_RELAY_BACKEND_PUBLIC_KEY  NEXT_DEV_RELAY_BACKEND_PUBLIC_KEY
#endif // #if !NEXT_DEVELOPMENT

#define NEXT_CONFIG_BUCKET_NAME "theodore_network_next_sdk_config"

extern uint8_t next_server_backend_public_key[];

extern uint8_t next_relay_backend_public_key[];

#endif // #ifndef NEXT_CONFIG_H
