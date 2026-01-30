/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#ifndef NEXT_PLATFORM_PS4_H
#define NEXT_PLATFORM_PS4_H

#if NEXT_PLATFORM == NEXT_PLATFORM_PS4

#include <kernel.h>
#include <net.h>

// -------------------------------------

typedef SceNetId next_platform_socket_handle_t;

struct next_platform_socket_t
{
    next_platform_socket_handle_t handle;
    int type;
    float timeout_seconds;
    void * context;
};

// -------------------------------------

struct next_platform_thread_t
{
    ScePthread handle;
    void * context;
};

// -------------------------------------

struct next_platform_mutex_t
{
    bool ok;
    ScePthreadMutex handle;
};

// -------------------------------------

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_PS4

#endif // #ifndef NEXT_PLATFORM_PS4_H
