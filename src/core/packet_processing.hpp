#pragma once

#include "core/continue_token.hpp"
#include "core/relay_manager.hpp"
#include "core/route_token.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
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
          LOG(ERROR, "failed to send new ping to ", pkt.addr);
        }

        size_t header_size = 0;
        if (pkt.addr.type == AddressType::IPv4) {
          header_size = net::IPV4_UDP_PACKET_HEADER_SIZE;
        } else if (pkt.addr.type == AddressType::IPv6) {
          header_size = net::IPV6_UDP_PACKET_HEADER_SIZE;
        }

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
        LOG(ERROR, "failed to receive packet");
        continue;
      }

      PacketType type;
      type = static_cast<PacketType>(packet.buffer[0]);
      size_t header_bytes = 0;

      if (packet.addr.type == net::AddressType::IPv4) {
        header_bytes = net::IPV4_UDP_PACKET_HEADER_SIZE;
      } else if (packet.addr.type == net::AddressType::IPv6) {
        header_bytes = net::IPV6_UDP_PACKET_HEADER_SIZE;
      }

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
        default: {
          recorder.unknown_rx.add(whole_packet_size);
        } break;
      }
    }
  }
}  // namespace core
