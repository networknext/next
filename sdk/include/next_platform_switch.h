/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#ifndef NEXT_SWITCH_H
#define NEXT_SWITCH_H

#if NEXT_PLATFORM == NEXT_PLATFORM_SWITCH

#include "next_address.h"

#include <nn/os.h>

// -------------------------------------

typedef int next_platform_socket_handle_t;

struct next_platform_socket_t
{
    next_platform_socket_handle_t handle;
    next_address_t address;
    int type;
    float timeout_seconds;
    int send_buffer_size;
    int receive_buffer_size;
    void * context;
};

// -------------------------------------

struct next_platform_thread_t
{
    nn::os::ThreadType handle;
    char * stack;
    void * context;
};

// -------------------------------------

struct next_platform_mutex_t
{
    nn::os::MutexType handle;
};

// -------------------------------------

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_SWITCH

#endif // #ifndef NEXT_SWITCH_H
