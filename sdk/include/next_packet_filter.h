/*
    Network Next. Copyright 2017 - 2026 Network Next, Inc.  
    Licensed under the Network Next Source Available License 1.0
*/

#ifndef NEXT_PACKET_FILTER_H
#define NEXT_PACKET_FILTER_H

#include "next.h"

void next_generate_pittle( uint8_t * output, const uint8_t * from_address, const uint8_t * to_address, uint16_t packet_length );

void next_generate_chonkle( uint8_t * output, const uint8_t * magic, const uint8_t * from_address, const uint8_t * to_address, uint16_t packet_length );

bool next_basic_packet_filter( const uint8_t * data, uint16_t packet_length );

void next_address_data( const next_address_t * address, uint8_t * address_data );

#if NEXT_ADVANCED_PACKET_FILTER

bool next_advanced_packet_filter( const uint8_t * data, const uint8_t * magic, const uint8_t * from_address, const uint8_t * to_address, uint16_t packet_length );

#endif // #if NEXT_ADVANCED_PACKET_FILTER

#endif // #ifndef NEXT_PACKET_FILTER_H
