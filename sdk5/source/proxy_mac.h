/*
    Network Next Proxy. Copyright © 2017 - 2022 Network Next, Inc.

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

#include "proxy.h"

#ifndef PROXY_MAC_H
#define PROXY_MAC_H

#if PROXY_PLATFORM == PROXY_PLATFORM_MAC

#include <pthread.h>
#include <unistd.h>

#define PROXY_PLATFORM_SOCKET_NON_BLOCKING       (1<<0)
#define PROXY_PLATFORM_SOCKET_REUSE_PORT         (1<<1)

// -------------------------------------

typedef int proxy_platform_socket_handle_t;

struct proxy_platform_socket_t
{
    proxy_platform_socket_handle_t handle;
};

// -------------------------------------

struct proxy_platform_thread_t
{
    pthread_t handle;
};

typedef void * proxy_platform_thread_return_t;

#define PROXY_PLATFORM_THREAD_RETURN() do { return NULL; } while ( 0 )

#define PROXY_PLATFORM_THREAD_FUNC

typedef proxy_platform_thread_return_t (PROXY_PLATFORM_THREAD_FUNC proxy_platform_thread_func_t)(void*);

// -------------------------------------

struct proxy_platform_mutex_t
{
    bool ok;
    pthread_mutex_t handle;
};

// -------------------------------------

#endif // #if PROXY_PLATFORM == PROXY_PLATFORM_MAC

#endif // #ifndef PROXY_MAC_H
