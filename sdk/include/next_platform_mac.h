/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#ifndef NEXT_PLATFORM_MAC_H
#define NEXT_PLATFORM_MAC_H

#if NEXT_PLATFORM == NEXT_PLATFORM_MAC

#include <pthread.h>
#include <unistd.h>

// -------------------------------------

typedef int next_platform_socket_handle_t;

struct next_platform_socket_t
{
    void * context;
    bool ipv6;
    next_platform_socket_handle_t handle;
};

// -------------------------------------

struct next_platform_thread_t
{
    void * context;
    pthread_t handle;
};

// -------------------------------------

struct next_platform_mutex_t
{
    bool ok;
    pthread_mutex_t handle;
};

// -------------------------------------

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_MAC

#endif // #ifndef NEXT_PLATFORM_MAC_H
