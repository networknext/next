/*
    Network Next Relay.
    Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

#include "relay.h"

#ifndef RELAY_LINUX_H
#define RELAY_LINUX_H

#if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#include <pthread.h>
#include <unistd.h>
#include <sched.h>

#define RELAY_PLATFORM_HAS_IPV6                  1
#define RELAY_PLATFORM_SOCKET_NON_BLOCKING       0
#define RELAY_PLATFORM_SOCKET_BLOCKING           1

// -------------------------------------

typedef int relay_platform_socket_handle_t;

struct relay_platform_socket_t
{
    int type;
    relay_platform_socket_handle_t handle;
};

// -------------------------------------

struct relay_platform_thread_t
{
    pthread_t handle;
};

typedef void * relay_platform_thread_return_t;

#define RELAY_PLATFORM_THREAD_RETURN() do { return NULL; } while ( 0 )

#define RELAY_PLATFORM_THREAD_FUNC

typedef relay_platform_thread_return_t (RELAY_PLATFORM_THREAD_FUNC relay_platform_thread_func_t)(void*);

// -------------------------------------

struct relay_platform_mutex_t
{
    pthread_mutex_t handle;
};

// -------------------------------------

#endif // #if RELAY_PLATFORM == RELAY_PLATFORM_LINUX

#endif // #ifndef RELAY_LINUX_H
