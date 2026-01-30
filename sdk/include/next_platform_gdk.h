/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#ifndef NEXT_PLATFORM_GDK_H
#define NEXT_PLATFORM_GDK_H

#ifdef _GAMING_XBOX

#define _WINSOCKAPI_
#include <windows.h>
#include <winsock2.h>

// -------------------------------------

#pragma warning(disable:4996)

#if _WIN64
    typedef uint64_t next_platform_socket_handle_t;
#else
    typedef _W64 unsigned int next_platform_socket_handle_t;
#endif

struct next_platform_socket_t
{
    void * context;
    bool ipv6;
    next_platform_socket_handle_t handle;
};

// -------------------------------------

struct next_platform_thread_t
{
    void* context;
    HANDLE handle;
};

// -------------------------------------

struct next_platform_mutex_t
{
    bool ok;
    CRITICAL_SECTION handle;
};

// -------------------------------------

#endif // #ifdef _GAMING_XBOX

#endif // #ifndef NEXT_PLATFORM_GDK_H
