/*
    Network Next. Copyright 2017 - 2025 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_AUTODETECT_H
#define NEXT_AUTODETECT_H

#include "next.h"

#if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

bool next_autodetect_google( char * output, size_t output_size );

bool next_autodetect_amazon( char * output, size_t output_size );

bool next_autodetect_datacenter( const char * input_datacenter, const char * public_address, char * output, size_t output_size );

#endif // #if NEXT_PLATFORM == NEXT_PLATFORM_LINUX || NEXT_PLATFORM == NEXT_PLATFORM_MAC || NEXT_PLATFORM == NEXT_PLATFORM_WINDOWS

#endif // #ifndef NEXT_AUTODETECT_H
