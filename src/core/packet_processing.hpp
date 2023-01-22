#pragma once

#include "core/continue_token.hpp"
#include "core/relay_manager.hpp"
#include "core/route_token.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "handlers/client_to_server_handler.hpp"
#include "handlers/continue_request_handler.hpp"
#include "handlers/continue_response_handler.hpp"
#include "handlers/near_ping_handler.hpp"
#include "handlers/relay_ping_handler.hpp"
#include "handlers/relay_pong_handler.hpp"
#include "handlers/route_request_handler.hpp"
#include "handlers/route_response_handler.hpp"
#include "handlers/server_to_client_handler.hpp"
#include "handlers/session_ping_handler.hpp"
#include "handlers/session_pong_handler.hpp"
#include "os/socket.hpp"
#include "packet.hpp"
#include "packet_types.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "token.hpp"
#include "util/macros.hpp"

using namespace std::chrono_literals;

using core::Packet;
using core::PacketType;
using core::RELAY_PING_PACKET_SIZE;
using crypto::Keychain;
using net::Address;
using net::AddressType;
using os::Socket;
using util::ThroughputRecorder;

// -----------------------------------------------------

#define next_assert assert

#define NEXT_RELAY_PING_PACKET    20
#define NEXT_RELAY_PONG_PACKET    21

#define NEXT_ADDRESS_NONE          0
#define NEXT_ADDRESS_IPV4          1
#define NEXT_ADDRESS_IPV6          2

#define NEXT_EXPORT_FUNC

#define NEXT_BOOL                int

#define NEXT_ADDRESS_BYTES        19

// -----------------------------------------------------

struct next_address_t
{
    union { uint8_t ipv4[4]; uint16_t ipv6[8]; } data;
    uint16_t port;
    uint8_t type;
};

NEXT_EXPORT_FUNC int next_address_parse( struct next_address_t * address, const char * address_string );

NEXT_EXPORT_FUNC const char * next_address_to_string( const struct next_address_t * address, char * buffer );

NEXT_EXPORT_FUNC NEXT_BOOL next_address_equal( const struct next_address_t * a, const struct next_address_t * b );

NEXT_EXPORT_FUNC void next_address_anonymize( struct next_address_t * address );

// -----------------------------------------------------

void next_write_uint8( uint8_t ** p, uint8_t value )
{
    **p = value;
    ++(*p);
}

void next_write_uint16( uint8_t ** p, uint16_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = value >> 8;
    *p += 2;
}

void next_write_uint32( uint8_t ** p, uint32_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = ( value >> 8  ) & 0xFF;
    (*p)[2] = ( value >> 16 ) & 0xFF;
    (*p)[3] = value >> 24;
    *p += 4;
}

void next_write_uint64( uint8_t ** p, uint64_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = ( value >> 8  ) & 0xFF;
    (*p)[2] = ( value >> 16 ) & 0xFF;
    (*p)[3] = ( value >> 24 ) & 0xFF;
    (*p)[4] = ( value >> 32 ) & 0xFF;
    (*p)[5] = ( value >> 40 ) & 0xFF;
    (*p)[6] = ( value >> 48 ) & 0xFF;
    (*p)[7] = value >> 56;
    *p += 8;
}

void next_write_float32( uint8_t ** p, float value )
{
    uint32_t value_int = 0;
    char * p_value = (char*)(&value);
    char * p_value_int = (char*)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint32_t));
    next_write_uint32( p, value_int);
}

void next_write_float64( uint8_t ** p, double value )
{
    uint64_t value_int = 0;
    char * p_value = (char *)(&value);
    char * p_value_int = (char *)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint64_t));
    next_write_uint64( p, value_int);
}

void next_write_bytes( uint8_t ** p, const uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        next_write_uint8( p, byte_array[i] );
    }
}

uint8_t next_read_uint8( const uint8_t ** p )
{
    uint8_t value = **p;
    ++(*p);
    return value;
}

uint16_t next_read_uint16( const uint8_t ** p )
{
    uint16_t value;
    value = (*p)[0];
    value |= ( ( (uint16_t)( (*p)[1] ) ) << 8 );
    *p += 2;
    return value;
}

uint32_t next_read_uint32( const uint8_t ** p )
{
    uint32_t value;
    value  = (*p)[0];
    value |= ( ( (uint32_t)( (*p)[1] ) ) << 8 );
    value |= ( ( (uint32_t)( (*p)[2] ) ) << 16 );
    value |= ( ( (uint32_t)( (*p)[3] ) ) << 24 );
    *p += 4;
    return value;
}

uint64_t next_read_uint64( const uint8_t ** p )
{
    uint64_t value;
    value  = (*p)[0];
    value |= ( ( (uint64_t)( (*p)[1] ) ) << 8  );
    value |= ( ( (uint64_t)( (*p)[2] ) ) << 16 );
    value |= ( ( (uint64_t)( (*p)[3] ) ) << 24 );
    value |= ( ( (uint64_t)( (*p)[4] ) ) << 32 );
    value |= ( ( (uint64_t)( (*p)[5] ) ) << 40 );
    value |= ( ( (uint64_t)( (*p)[6] ) ) << 48 );
    value |= ( ( (uint64_t)( (*p)[7] ) ) << 56 );
    *p += 8;
    return value;
}

float next_read_float32( const uint8_t ** p )
{
    uint32_t value_int = next_read_uint32( p );
    float value_float = 0.0f;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy( pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

double next_read_float64( const uint8_t ** p )
{
    uint64_t value_int = next_read_uint64( p );
    double value_float = 0.0;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy( pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

void next_read_bytes( const uint8_t ** p, uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        byte_array[i] = next_read_uint8( p );
    }
}

// -----------------------------------------------------

void next_write_address( uint8_t ** buffer, const next_address_t * address )
{
    next_assert( buffer );
    next_assert( *buffer );
    next_assert( address );

    uint8_t * start = *buffer;

    (void) buffer;

    if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        next_write_uint8( buffer, NEXT_ADDRESS_IPV4 );
        for ( int i = 0; i < 4; ++i )
        {
            next_write_uint8( buffer, address->data.ipv4[i] );
        }
        next_write_uint16( buffer, address->port );
        for ( int i = 0; i < 12; ++i )
        {
            next_write_uint8( buffer, 0 );
        }
    }
    else if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        next_write_uint8( buffer, NEXT_ADDRESS_IPV6 );
        for ( int i = 0; i < 8; ++i )
        {
            next_write_uint16( buffer, address->data.ipv6[i] );
        }
        next_write_uint16( buffer, address->port );
    }
    else
    {
        for ( int i = 0; i < NEXT_ADDRESS_BYTES; ++i )
        {
            next_write_uint8( buffer, 0 );
        }
    }

    (void) start;

    next_assert( *buffer - start == NEXT_ADDRESS_BYTES );
}

void next_read_address( const uint8_t ** buffer, next_address_t * address )
{
    const uint8_t * start = *buffer;

    memset( address, 0, sizeof(next_address_t) );

    address->type = next_read_uint8( buffer );

    if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        for ( int j = 0; j < 4; ++j )
        {
            address->data.ipv4[j] = next_read_uint8( buffer );
        }
        address->port = next_read_uint16( buffer );
        for ( int i = 0; i < 12; ++i )
        {
            uint8_t dummy = next_read_uint8( buffer ); (void) dummy;
        }
    }
    else if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        for ( int j = 0; j < 8; ++j )
        {
            address->data.ipv6[j] = next_read_uint16( buffer );
        }
        address->port = next_read_uint16( buffer );
    }
    else
    {
        for ( int i = 0; i < NEXT_ADDRESS_BYTES - 1; ++i )
        {
            uint8_t dummy = next_read_uint8( buffer ); (void) dummy;
        }
    }

    (void) start;

    next_assert( *buffer - start == NEXT_ADDRESS_BYTES );
}

void next_read_address_variable( const uint8_t ** buffer, next_address_t * address )
{
    const uint8_t * start = *buffer;

    memset( address, 0, sizeof(next_address_t) );

    address->type = next_read_uint8( buffer );

    if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        for ( int j = 0; j < 4; ++j )
        {
            address->data.ipv4[j] = next_read_uint8( buffer );
        }
        address->port = next_read_uint16( buffer );
    }
    else if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        for ( int j = 0; j < 8; ++j )
        {
            address->data.ipv6[j] = next_read_uint16( buffer );
        }
        address->port = next_read_uint16( buffer );
    }

    (void) start;
}

// -----------------------------------------------------

typedef uint64_t next_fnv_t;

void next_fnv_init( next_fnv_t * fnv )
{
    *fnv = 0xCBF29CE484222325;
}

void next_fnv_write( next_fnv_t * fnv, const uint8_t * data, size_t size )
{
    for ( size_t i = 0; i < size; i++ )
    {
        (*fnv) ^= data[i];
        (*fnv) *= 0x00000100000001B3;
    }
}

uint64_t next_fnv_finalize( next_fnv_t * fnv )
{
    return *fnv;
}

uint64_t next_hash_string( const char * string )
{
    next_fnv_t fnv;
    next_fnv_init( &fnv );
    next_fnv_write( &fnv, (uint8_t *)( string ), strlen( string ) );
    return next_fnv_finalize( &fnv );
}

// -----------------------------------------------------

static void next_generate_pittle( uint8_t * output, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    next_assert( output );
    next_assert( from_address );
    next_assert( from_address_bytes > 0 );
    next_assert( to_address );
    next_assert( to_address_bytes >= 0 );
    next_assert( packet_length > 0 );
#if NEXT_BIG_ENDIAN
    next_bswap( from_port );
    next_bswap( to_port );
    next_bswap( packet_length );
#endif // #if NEXT_BIG_ENDIAN
    uint16_t sum = 0;
    for ( int i = 0; i < from_address_bytes; ++i ) { sum += uint8_t(from_address[i]); }
    const char * from_port_data = (const char*) &from_port;
    sum += uint8_t(from_port_data[0]);
    sum += uint8_t(from_port_data[1]);
    for ( int i = 0; i < to_address_bytes; ++i ) { sum += uint8_t(to_address[i]); }
    const char * to_port_data = (const char*) &to_port;
    sum += uint8_t(to_port_data[0]);
    sum += uint8_t(to_port_data[1]);
    const char * packet_length_data = (const char*) &packet_length;
    sum += uint8_t(packet_length_data[0]);
    sum += uint8_t(packet_length_data[1]);
    sum += uint8_t(packet_length_data[2]);
    sum += uint8_t(packet_length_data[3]);
#if NEXT_BIG_ENDIAN
    next_bswap( sum );
#endif // #if NEXT_BIG_ENDIAN
    const char * sum_data = (const char*) &sum;
    output[0] = 1 | ( uint8_t(sum_data[0]) ^ uint8_t(sum_data[1]) ^ 193 );
    output[1] = 1 | ( ( 255 - output[0] ) ^ 113 );
}

static void next_generate_chonkle( uint8_t * output, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    next_assert( output );
    next_assert( magic );
    next_assert( from_address );
    next_assert( from_address_bytes >= 0 );
    next_assert( to_address );
    next_assert( to_address_bytes >= 0 );
    next_assert( packet_length > 0 );
#if NEXT_BIG_ENDIAN
    next_bswap( from_port );
    next_bswap( to_port );
    next_bswap( packet_length );
#endif // #if NEXT_BIG_ENDIAN
    next_fnv_t fnv;
    next_fnv_init( &fnv );
    next_fnv_write( &fnv, magic, 8 );
    next_fnv_write( &fnv, from_address, from_address_bytes );
    next_fnv_write( &fnv, (const uint8_t*) &from_port, 2 );
    next_fnv_write( &fnv, to_address, to_address_bytes );
    next_fnv_write( &fnv, (const uint8_t*) &to_port, 2 );
    next_fnv_write( &fnv, (const uint8_t*) &packet_length, 4 );
    uint64_t hash = next_fnv_finalize( &fnv );
#if NEXT_BIG_ENDIAN
    next_bswap( hash );
#endif // #if NEXT_BIG_ENDIAN
    const char * data = (const char*) &hash;
    output[0] = ( ( data[6] & 0xC0 ) >> 6 ) + 42;
    output[1] = ( data[3] & 0x1F ) + 200;
    output[2] = ( ( data[2] & 0xFC ) >> 2 ) + 5;
    output[3] = data[0];
    output[4] = ( data[2] & 0x03 ) + 78;
    output[5] = ( data[4] & 0x7F ) + 96;
    output[6] = ( ( data[1] & 0xFC ) >> 2 ) + 100;
    if ( ( data[7] & 1 ) == 0 ) { output[7] = 79; } else { output[7] = 7; }
    if ( ( data[4] & 0x80 ) == 0 ) { output[8] = 37; } else { output[8] = 83; }
    output[9] = ( data[5] & 0x07 ) + 124;
    output[10] = ( ( data[1] & 0xE0 ) >> 5 ) + 175;
    output[11] = ( data[6] & 0x3F ) + 33;
    const int value = ( data[1] & 0x03 );
    if ( value == 0 ) { output[12] = 97; } else if ( value == 1 ) { output[12] = 5; } else if ( value == 2 ) { output[12] = 43; } else { output[12] = 13; }
    output[13] = ( ( data[5] & 0xF8 ) >> 3 ) + 210;
    output[14] = ( ( data[7] & 0xFE ) >> 1 ) + 17;
}

bool next_basic_packet_filter( const uint8_t * data, int packet_length )
{
    if ( packet_length == 0 )
        return false;

    if ( data[0] == 0 )
        return true;

    if ( packet_length < 18 )
        return false;

    if ( data[0] < 0x01 || data[0] > 0x63 )
        return false;

    if ( data[1] < 0x2A || data[1] > 0x2D )
        return false;

    if ( data[2] < 0xC8 || data[2] > 0xE7 )
        return false;

    if ( data[3] < 0x05 || data[3] > 0x44 )
        return false;

    if ( data[5] < 0x4E || data[5] > 0x51 )
        return false;

    if ( data[6] < 0x60 || data[6] > 0xDF )
        return false;

    if ( data[7] < 0x64 || data[7] > 0xE3 )
        return false;

    if ( data[8] != 0x07 && data[8] != 0x4F )
        return false;

    if ( data[9] != 0x25 && data[9] != 0x53 )
        return false;

    if ( data[10] < 0x7C || data[10] > 0x83 )
        return false;

    if ( data[11] < 0xAF || data[11] > 0xB6 )
        return false;

    if ( data[12] < 0x21 || data[12] > 0x60 )
        return false;

    if ( data[13] != 0x61 && data[13] != 0x05 && data[13] != 0x2B && data[13] != 0x0D )
        return false;

    if ( data[14] < 0xD2 || data[14] > 0xF1 )
        return false;

    if ( data[15] < 0x11 || data[15] > 0x90 )
        return false;

    return true;
}

void next_address_data( const next_address_t * address, uint8_t * address_data, int * address_bytes, uint16_t * address_port )
{
    next_assert( address );
    if ( address->type == NEXT_ADDRESS_IPV4 )
    {
        address_data[0] = address->data.ipv4[0];
        address_data[1] = address->data.ipv4[1];
        address_data[2] = address->data.ipv4[2];
        address_data[3] = address->data.ipv4[3];
        *address_bytes = 4;
    }
    else if ( address->type == NEXT_ADDRESS_IPV6 )
    {
        for ( int i = 0; i < 8; ++i )
        {
            address_data[i*2]   = address->data.ipv6[i] >> 8;
            address_data[i*2+1] = address->data.ipv6[i] & 0xFF;
        }
        *address_bytes = 16;
    }
    else
    {
        *address_bytes = 0;
    }
    *address_port = address->port;
}

bool next_advanced_packet_filter( const uint8_t * data, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port, int packet_length )
{
    if ( data[0] == 0 )
        return true;

    if ( packet_length < 18 )
        return false;
    
    uint8_t a[15];
    uint8_t b[2];

    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    if ( memcmp( a, data + 1, 15 ) != 0 )
        return false;
    if ( memcmp( b, data + packet_length - 2, 2 ) != 0 )
        return false;
    return true;
}

// -----------------------------------------------------

int next_write_relay_pong_packet( uint8_t * packet_data, uint64_t ping_sequence, uint64_t session_id, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    next_write_uint8( &p, NEXT_RELAY_PONG_PACKET );
    uint8_t * a = p; p += 15;
    next_write_uint64( &p, ping_sequence );
    next_write_uint64( &p, session_id );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    // todo: bring these across
    /*
    next_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    next_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    */
    return packet_length;
}

// ------------------------------------

namespace core
{
  INLINE void ping_loop(
   const Socket& socket, RelayManager& relay_manager, const volatile bool& should_process, ThroughputRecorder& recorder)
  {
    Packet pkt;

    while (!socket.closed() && should_process) {
      // Sleep for 10ms, but the actual ping rate is controlled by PING_RATE
      std::this_thread::sleep_for(10ms);

      std::array<core::PingData, MAX_RELAYS> pings;

      auto number_of_relays_to_ping = relay_manager.get_ping_targets(pings);

      if (number_of_relays_to_ping == 0) {
        continue;
      }

      for (unsigned int i = 0; i < number_of_relays_to_ping; i++) {
        const auto& ping = pings[i];

        pkt.addr = ping.address;

        // write data to the buffer
        size_t index = 0;

        if (!encoding::write_uint8(pkt.buffer, index, static_cast<uint8_t>(PacketType::RelayPing))) {
          LOG(ERROR, "could not write packet type");
          assert(false);
        }

        if (!encoding::write_uint64(pkt.buffer, index, ping.sequence)) {
          LOG(ERROR, "could not write sequence");
          assert(false);
        }

        pkt.length = index;

        if (socket.closed() || !should_process) {
          break;
        }

        if (!socket.send(pkt)) {
          LOG(DEBUG, "failed to send new ping to ", pkt.addr);
        }

        size_t header_size = net::IPV4_UDP_PACKET_HEADER_SIZE;

        size_t whole_packet_size = header_size + pkt.length;

        recorder.outbound_ping_tx.add(whole_packet_size);
      }
    }
  }

  INLINE void recv_loop(
   const std::atomic<bool>& should_receive,
   const Socket& socket,
   const Keychain& keychain,
   SessionMap& session_map,
   RelayManager& relay_manager,
   const volatile bool& should_handle,
   ThroughputRecorder& recorder,
   const RouterInfo& router_info)
  {
    core::Packet packet;

    while (!socket.closed() && should_receive) {
      if (!socket.recv(packet)) {
        LOG(DEBUG, "failed to receive packet");
        continue;
      }

      PacketType type;
      type = static_cast<PacketType>(packet.buffer[0]);
      size_t header_bytes = net::IPV4_UDP_PACKET_HEADER_SIZE;

      size_t whole_packet_size = packet.length + header_bytes;

      switch (type) {
        // Relay to relay
        case PacketType::RelayPing: {
          recorder.inbound_ping_rx.add(whole_packet_size);
          handlers::relay_ping_handler(packet, recorder, socket, should_handle);
        } break;
        case PacketType::RelayPong: {
          recorder.pong_rx.add(whole_packet_size);
          handlers::relay_pong_handler(packet, relay_manager, should_handle);
        } break;
        case PacketType::RouteRequest: {
          recorder.route_request_rx.add(whole_packet_size);
          handlers::route_request_handler(packet, keychain, session_map, recorder, router_info, socket);
        } break;
        case PacketType::RouteResponse: {
          recorder.route_response_rx.add(whole_packet_size);
          handlers::route_response_handler(packet, session_map, recorder, router_info, socket);
        } break;
        case PacketType::ClientToServer: {
          recorder.client_to_server_rx.add(whole_packet_size);
          handlers::client_to_server_handler(packet, session_map, recorder, router_info, socket);
        } break;
        case PacketType::ServerToClient: {
          recorder.server_to_client_rx.add(whole_packet_size);
          handlers::server_to_client_handler(packet, session_map, recorder, router_info, socket);
        } break;
        case PacketType::SessionPing: {
          recorder.session_ping_rx.add(whole_packet_size);
          handlers::session_ping_handler(packet, session_map, recorder, router_info, socket);
        } break;
        case PacketType::SessionPong: {
          recorder.session_pong_rx.add(whole_packet_size);
          handlers::session_pong_handler(packet, session_map, recorder, router_info, socket);
        } break;
        case PacketType::ContinueRequest: {
          recorder.continue_request_rx.add(whole_packet_size);
          handlers::continue_request_handler(packet, session_map, keychain, recorder, router_info, socket);
        } break;
        case PacketType::ContinueResponse: {
          recorder.continue_response_rx.add(whole_packet_size);
          handlers::continue_response_handler(packet, session_map, recorder, router_info, socket);
        } break;
        case PacketType::NearPing: {
          recorder.near_ping_rx.add(whole_packet_size);
          handlers::near_ping_handler(packet, recorder, socket);
        } break;


        // sdk5
	    case PacketType::NewNearPing: 
	    {
	    	printf( "received new near relay ping packet\n" );

	    	int packet_bytes = int(packet.length);
	    	uint8_t * packet_data = &packet.buffer[0];

	    	if ( packet_bytes != 16 + 8 + 8 + 86 + 2 )
	    	{
	    		printf( "wrong packet size %d bytes\n", packet_bytes );
	    		return;
	    	}

		    const uint8_t * p = packet_data + 16;
		    uint64_t read_ping_sequence = next_read_uint64( &p );
		    uint64_t read_ping_session_id = next_read_uint64( &p );

		    printf( "ping sequence is %" PRIx64 "\n", read_ping_sequence );
		    printf( "ping session id is %" PRIx64 "\n", read_ping_session_id );

/*
		      size_t length = packet.length;

		      if (length != 1 + 8 + 8 + 8 + 8) {
		        LOG(ERROR, "ignoring near ping packet, length invalid: ", length);
		        return;
		      }

		      length = packet.length - 16;

		      packet.buffer[0] = static_cast<uint8_t>(PacketType::NearPong);

		      recorder.near_ping_tx.add(length);

		      if (!socket.send(packet.addr, packet.buffer.data(), length)) {
		        LOG(ERROR, "failed to send near pong to ", packet.addr);
		      }
*/
	    }
	    break;


        default: {
          recorder.unknown_rx.add(whole_packet_size);
        } break;
      }
    }
  }
}  // namespace core
