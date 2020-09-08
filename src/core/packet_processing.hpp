#pragma once

#include "core/continue_token.hpp"
#include "core/relay_manager.hpp"
#include "core/route_token.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/hash.hpp"
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
using core::RELAY_PING_PACKET_SIZE;
using core::PacketType;
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
      // Sleep for 10ms, but the actual ping rate is controlled by RELAY_PING_TIME
      std::this_thread::sleep_for(10ms);

      std::array<core::PingData, MAX_RELAYS> pings;

      auto numberOfRelaysToPing = relay_manager.get_ping_targets(pings);

      if (numberOfRelaysToPing == 0) {
        continue;
      }

      for (unsigned int i = 0; i < numberOfRelaysToPing; i++) {
        const auto& ping = pings[i];

        pkt.addr = ping.address;

        size_t index = crypto::PACKET_HASH_LENGTH;

        // write data to the buffer
        {
          if (!encoding::write_uint8(pkt.buffer, index, static_cast<uint8_t>(PacketType::RelayPing))) {
            LOG(ERROR, "could not write packet type");
            assert(false);
          }

          if (!encoding::write_uint64(pkt.buffer, index, ping.sequence)) {
            LOG(ERROR, "could not write sequence");
            assert(false);
          }

          size_t sign_index = 0;
          if (!crypto::sign_network_next_packet(pkt.buffer, sign_index, index)) {
            LOG(ERROR, "unable to sign ping packet");
            continue;
          }
        }

        pkt.length = index;

        if (socket.closed() || !should_process) {
          break;
        }

        if (!socket.send(pkt)) {
          LOG(ERROR, "failed to send new ping to ", pkt.addr);
        }

        size_t headerSize = 0;
        if (pkt.addr.Type == AddressType::IPv4) {
          headerSize = net::IPv4UDPHeaderSize;
        } else if (pkt.addr.Type == AddressType::IPv6) {
          headerSize = net::IPv6UDPHeaderSize;
        }

        size_t wholePacketSize = headerSize + pkt.length;

        // could also just do: (1 + 8) * number of relays to ping to make this faster
        recorder.outbound_ping_tx.add(wholePacketSize);
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

      size_t headerBytes = 0;

      if (packet.addr.Type == net::AddressType::IPv4) {
        headerBytes = net::IPv4UDPHeaderSize;
      } else if (packet.addr.Type == net::AddressType::IPv6) {
        headerBytes = net::IPv6UDPHeaderSize;
      }

      size_t wholePacketSize = packet.length + headerBytes;

      PacketType type;

      bool is_signed;
      size_t sign_index = 0;
      if (crypto::is_network_next_packet(packet.buffer, sign_index, packet.length)) {
        type = static_cast<PacketType>(packet.buffer[crypto::PACKET_HASH_LENGTH]);
        is_signed = true;
      } else {
        type = static_cast<PacketType>(packet.buffer[0]);
        is_signed = false;
      }

      if (type != PacketType::RelayPing && type != PacketType::RelayPong) {
        if (is_signed) {
          LOG(DEBUG, "packet is from network next");
        } else {
          LOG(DEBUG, "packet is not on network next");
        }
        LOG(DEBUG, "incoming packet, type = ", type);
      }

      switch (type) {
        case PacketType::RelayPing: {
          recorder.inbound_ping_rx.add(wholePacketSize);
          handlers::relay_ping_handler(packet, recorder, socket, should_handle);
        } break;
        case PacketType::RelayPong: {
          recorder.pong_rx.add(wholePacketSize);
          handlers::relay_pong_handler(packet, relay_manager, should_handle);
        } break;
        case PacketType::RouteRequest: {
          recorder.route_request_rx.add(wholePacketSize);
          handlers::route_request_handler(packet, keychain, session_map, recorder, router_info, socket, is_signed);
        } break;
        case PacketType::RouteResponse: {
          recorder.route_response_rx.add(wholePacketSize);
          handlers::route_response_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case PacketType::ContinueRequest: {
          recorder.continue_request_rx.add(wholePacketSize);
          handlers::continue_request_handler(packet, session_map, keychain, recorder, router_info, socket, is_signed);
        } break;
        case PacketType::ContinueResponse: {
          recorder.continue_response_rx.add(wholePacketSize);
          handlers::continue_response_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case PacketType::ClientToServer: {
          recorder.client_to_server_rx.add(wholePacketSize);
          handlers::client_to_server_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case PacketType::ServerToClient: {
          recorder.server_to_client_rx.add(wholePacketSize);
          handlers::server_to_client_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case PacketType::SessionPing: {
          recorder.session_ping_rx.add(wholePacketSize);
          handlers::session_ping_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case PacketType::SessionPong: {
          recorder.session_pong_rx.add(wholePacketSize);
          handlers::session_pong_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case PacketType::NearPing: {
          recorder.near_ping_rx.add(wholePacketSize);
          handlers::near_ping_handler(packet, recorder, socket, is_signed);
        } break;
        default: {
          recorder.unknown_rx.add(wholePacketSize);
        } break;
      }
    }
  }
}  // namespace core
