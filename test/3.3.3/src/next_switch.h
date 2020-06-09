/*
    Network Next SDK 3.3.3

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

#ifndef NEXT_SWITCH_H
#define NEXT_SWITCH_H

#if NEXT_PLATFORM == NEXT_PLATFORM_SWITCH

#include <nn/os.h>

#define NEXT_PLATFORM_HAS_IPV6                  0
#define NEXT_PLATFORM_SOCKET_NON_BLOCKING       0
#define NEXT_PLATFORM_SOCKET_BLOCKING           1

NEXT_PACK_PUSH()

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

typedef void next_platform_thread_return_t;

#define NEXT_PLATFORM_THREAD_RETURN() do { return; } while ( 0 )

#define NEXT_PLATFORM_THREAD_FUNC

typedef next_platform_thread_return_t (NEXT_PLATFORM_THREAD_FUNC next_platform_thread_func_t)(void*);

// -------------------------------------

struct next_platform_mutex_t
{
    nn::os::MutexType handle;
	void * context;
};

// -------------------------------------

NEXT_PACK_POP()

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_SWITCH

#endif // #ifndef NEXT_SWITCH_H
