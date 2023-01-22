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

// ------------------------------------

#define NEXT_RELAY_PING_PACKET 		20
#define NEXT_RELAY_PONG_PACKET 		21

void relay_write_uint8( uint8_t ** p, uint8_t value )
{
    **p = value;
    ++(*p);
}

void relay_write_uint16( uint8_t ** p, uint16_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = value >> 8;
    *p += 2;
}

void relay_write_uint32( uint8_t ** p, uint32_t value )
{
    (*p)[0] = value & 0xFF;
    (*p)[1] = ( value >> 8  ) & 0xFF;
    (*p)[2] = ( value >> 16 ) & 0xFF;
    (*p)[3] = value >> 24;
    *p += 4;
}

void relay_write_uint64( uint8_t ** p, uint64_t value )
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

void relay_write_float32( uint8_t ** p, float value )
{
    uint32_t value_int = 0;
    char * p_value = (char*)(&value);
    char * p_value_int = (char*)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint32_t));
    relay_write_uint32( p, value_int);
}

void relay_write_float64( uint8_t ** p, double value )
{
    uint64_t value_int = 0;
    char * p_value = (char *)(&value);
    char * p_value_int = (char *)(&value_int);
    memcpy(p_value_int, p_value, sizeof(uint64_t));
    relay_write_uint64( p, value_int);
}

void relay_write_bytes( uint8_t ** p, const uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        relay_write_uint8( p, byte_array[i] );
    }
}

void relay_write_string( uint8_t ** p, const char * string_data, uint32_t max_length )
{
    uint32_t length = strlen( string_data );
    assert( length <= max_length );
    if ( length > max_length - 1 )
        length = max_length - 1;
    relay_write_uint32( p, length );
    for ( uint32_t i = 0; i < length; ++i )
    {
        relay_write_uint8( p, string_data[i] );
    }
}

uint8_t relay_read_uint8( const uint8_t ** p )
{
    uint8_t value = **p;
    ++(*p);
    return value;
}

uint16_t relay_read_uint16( const uint8_t ** p )
{
    uint16_t value;
    value = (*p)[0];
    value |= ( ( (uint16_t)( (*p)[1] ) ) << 8 );
    *p += 2;
    return value;
}

uint32_t relay_read_uint32( const uint8_t ** p )
{
    uint32_t value;
    value  = (*p)[0];
    value |= ( ( (uint32_t)( (*p)[1] ) ) << 8 );
    value |= ( ( (uint32_t)( (*p)[2] ) ) << 16 );
    value |= ( ( (uint32_t)( (*p)[3] ) ) << 24 );
    *p += 4;
    return value;
}

uint64_t relay_read_uint64( const uint8_t ** p )
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

float relay_read_float32( const uint8_t ** p )
{
    uint32_t value_int = relay_read_uint32( p );
    float value_float = 0.0f;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy( pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

double relay_read_float64( const uint8_t ** p )
{
    uint64_t value_int = relay_read_uint64( p );
    double value_float = 0.0;
    uint8_t * pointer_int = (uint8_t *)( &value_int );
    uint8_t * pointer_float = (uint8_t *)( &value_float );
    memcpy( pointer_float, pointer_int, sizeof( value_int ) );
    return value_float;
}

void relay_read_bytes( const uint8_t ** p, uint8_t * byte_array, int num_bytes )
{
    for ( int i = 0; i < num_bytes; ++i )
    {
        byte_array[i] = relay_read_uint8( p );
    }
}

void relay_read_string( const uint8_t ** p, char * string_data, uint32_t max_length )
{
    uint32_t length = relay_read_uint32( p );
    if ( length > max_length )
    {
        length = 0;
        return;
    }
    uint32_t i = 0;
    for ( ; i < length; ++i )
    {
        string_data[i] = relay_read_uint8( p );
    }
    string_data[i] = 0;
}

int relay_write_relay_pong_packet( uint8_t * packet_data, uint64_t ping_sequence, uint64_t session_id, const uint8_t * magic, const uint8_t * from_address, int from_address_bytes, uint16_t from_port, const uint8_t * to_address, int to_address_bytes, uint16_t to_port )
{
    uint8_t * p = packet_data;
    relay_write_uint8( &p, NEXT_RELAY_PONG_PACKET );
    uint8_t * a = p; p += 15;
    relay_write_uint64( &p, ping_sequence );
    relay_write_uint64( &p, session_id );
    uint8_t * b = p; p += 2;
    int packet_length = p - packet_data;
    // todo: bring these across
    /*
    relay_generate_chonkle( a, magic, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
    relay_generate_pittle( b, from_address, from_address_bytes, from_port, to_address, to_address_bytes, to_port, packet_length );
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

        // could also just do: (1 + 8) * number of relays to ping to make this faster
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
		    uint64_t read_ping_sequence = relay_read_uint64( &p );
		    uint64_t read_ping_session_id = relay_read_uint64( &p );

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
