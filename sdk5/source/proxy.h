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

#ifndef PROXY_H
#define PROXY_H

#ifndef __STDC_FORMAT_MACROS
#define __STDC_FORMAT_MACROS
#endif

#include <stdint.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>

#define PROXY_LOG_LEVEL_NONE                                       0
#define PROXY_LOG_LEVEL_ERROR                                      1
#define PROXY_LOG_LEVEL_INFO                                       2
#define PROXY_LOG_LEVEL_WARN                                       3
#define PROXY_LOG_LEVEL_DEBUG                                      4

#define PROXY_ADDRESS_NONE                                         0
#define PROXY_ADDRESS_IPV4                                         1
#define PROXY_ADDRESS_IPV6                                         2

#define PROXY_MAX_ADDRESS_STRING_LENGTH                          256

#define PROXY_ADDRESS_BUFFER_SAFETY								  16

#define PROXY_PLATFORM_UNKNOWN                                     0
#define PROXY_PLATFORM_MAC                                         1
#define PROXY_PLATFORM_LINUX                                       2
#define PROXY_PLATFORM_MAX                                         2

#if defined(__APPLE__)
    #define PROXY_PLATFORM PROXY_PLATFORM_MAC
#else
    #define PROXY_PLATFORM PROXY_PLATFORM_LINUX
#endif

// ----------------------------------------------

bool proxy_init();

void proxy_term();

double proxy_time();

void proxy_sleep( double seconds );

void proxy_printf( int level, const char * format, ... );

// ----------------------------------------------

struct proxy_address_t
{
    union { uint8_t ipv4[4]; uint16_t ipv6[8]; } data;
    uint16_t port;
    uint8_t type;
};

bool proxy_address_parse( struct proxy_address_t * address, const char * address_string );

const char * proxy_address_to_string( const struct proxy_address_t * address, char * buffer );

bool proxy_address_equal( const struct proxy_address_t * a, const struct proxy_address_t * b );

void proxy_address_anonymize( struct proxy_address_t * address );

// ----------------------------------------------

#define PROXY_MUTEX_BYTES 256

struct proxy_mutex_t { uint8_t dummy[PROXY_MUTEX_BYTES]; };

int proxy_mutex_create( struct proxy_mutex_t * mutex );

void proxy_mutex_destroy( struct proxy_mutex_t * mutex );

void proxy_mutex_acquire( struct proxy_mutex_t * mutex );

void proxy_mutex_release( struct proxy_mutex_t * mutex );

struct proxy_mutex_helper_t
{
    struct proxy_mutex_t * _mutex;
    proxy_mutex_helper_t( struct proxy_mutex_t * mutex ) : _mutex( mutex ) { assert( mutex ); proxy_mutex_acquire( _mutex ); }
    ~proxy_mutex_helper_t() { assert( _mutex ); proxy_mutex_release( _mutex ); _mutex = NULL; }
};

#define proxy_mutex_guard( _mutex ) proxy_mutex_helper_t __mutex_helper( _mutex )

// ----------------------------------------------

#endif // #ifndef PROXY_H
