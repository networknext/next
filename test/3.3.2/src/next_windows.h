/*
    Network Next SDK 3.3.2

    Copyright © 2017 - 2019 Network Next, Inc.

    Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following 
    conditions are met:

    1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

    2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions 
       and the following disclaimer in the documentation and/or other materials provided with the distribution.

    3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote 
       products derived from this software without specific prior written permission.

    THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, 
    INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. 
    IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR 
    CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; 
    OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING 
    NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

#include "next.h"

#ifndef NEXT_WINDOWS_H
#define NEXT_WINDOWS_H

#if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

#define _WINSOCKAPI_

#pragma pack(push, 8)
#include <windows.h>
#include <winsock2.h>
#pragma pack(pop)

#define NEXT_PLATFORM_SOCKET_NON_BLOCKING       0
#define NEXT_PLATFORM_SOCKET_BLOCKING           1

#if WINVER <= 0x0502
#define NEXT_PLATFORM_HAS_IPV6 0
#else
#define NEXT_PLATFORM_HAS_IPV6 1
#endif

NEXT_PACK_PUSH()

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
    next_platform_socket_handle_t handle;
};

// -------------------------------------

struct next_platform_thread_t
{
    void * context;
    HANDLE handle;
};

typedef DWORD next_platform_thread_return_t;

#define NEXT_PLATFORM_THREAD_RETURN() do { return 0; } while ( 0 )

#define NEXT_PLATFORM_THREAD_FUNC WINAPI

typedef next_platform_thread_return_t (NEXT_PLATFORM_THREAD_FUNC next_platform_thread_func_t)(void*);

// -------------------------------------

struct next_platform_mutex_t
{
    void * context;
    CRITICAL_SECTION handle;
};

// -------------------------------------

NEXT_PACK_POP()

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

#endif // #ifndef NEXT_WINDOWS_H