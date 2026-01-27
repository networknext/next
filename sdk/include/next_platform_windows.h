/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#include "next.h"

#ifndef NEXT_PLATFORM_WINDOWS_H
#define NEXT_PLATFORM_WINDOWS_H

#if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

#if NEXT_UNREAL_ENGINE
#include "Windows/AllowWindowsPlatformTypes.h"
#include "Windows/PreWindowsApi.h"
#endif // #if NEXT_UNREAL_ENGINE

#if !NEXT_UNREAL_ENGINE
#define _WINSOCKAPI_
#include <windows.h>
#include <winsock2.h>
#else // #if !NEXT_UNREAL_ENGINE
#include "Windows/MinWindows.h"
#endif // #if !NEXT_UNREAL_ENGINE

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
    void * context;
    HANDLE handle;
};

// -------------------------------------

struct next_platform_mutex_t
{
    bool ok;
    CRITICAL_SECTION handle;
};

// -------------------------------------

#if NEXT_UNREAL_ENGINE
#include "Windows/PostWindowsApi.h"
#include "Windows/HideWindowsPlatformTypes.h"
#endif // #if NEXT_UNREAL_ENGINE

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

#endif // #ifndef NEXT_PLATFORM_WINDOWS_H
