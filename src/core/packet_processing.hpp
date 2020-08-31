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
#include "packets/types.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "token.hpp"
#include "util/macros.hpp"

using namespace std::chrono_literals;

using core::Packet;
using core::packets::RELAY_PING_PACKET_SIZE;
using core::packets::Type;
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

      auto numberOfRelaysToPing = relay_manager.getPingData(pings);

      if (numberOfRelaysToPing == 0) {
        continue;
      }

      for (unsigned int i = 0; i < numberOfRelaysToPing; i++) {
        const auto& ping = pings[i];

        pkt.Addr = ping.Addr;

        size_t index = crypto::PacketHashLength;

        // write data to the buffer
        {
          if (!encoding::WriteUint8(pkt.Buffer, index, static_cast<uint8_t>(Type::RelayPing))) {
            LOG(ERROR, "could not write packet type");
            assert(false);
          }

          if (!encoding::WriteUint64(pkt.Buffer, index, ping.Seq)) {
            LOG(ERROR, "could not write sequence");
            assert(false);
          }

          crypto::SignNetworkNextPacket(pkt.Buffer, index);
        }

        pkt.Len = index;

        if (socket.closed() || !should_process) {
          break;
        }

        if (!socket.send(pkt)) {
          LOG(ERROR, "failed to send new ping to ", pkt.Addr);
        }

        size_t headerSize = 0;
        if (pkt.Addr.Type == AddressType::IPv4) {
          headerSize = net::IPv4UDPHeaderSize;
        } else if (pkt.Addr.Type == AddressType::IPv6) {
          headerSize = net::IPv6UDPHeaderSize;
        }

        size_t wholePacketSize = headerSize + pkt.Len;

        // could also just do: (1 + 8) * number of relays to ping to make this faster
        recorder.OutboundPingTx.add(wholePacketSize);
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

      if (packet.Addr.Type == net::AddressType::IPv4) {
        headerBytes = net::IPv4UDPHeaderSize;
      } else if (packet.Addr.Type == net::AddressType::IPv6) {
        headerBytes = net::IPv6UDPHeaderSize;
      }

      size_t wholePacketSize = packet.Len + headerBytes;

      Type type;

      bool is_signed;
      if (crypto::IsNetworkNextPacket(packet.Buffer, packet.Len)) {
        type = static_cast<Type>(packet.Buffer[crypto::PacketHashLength]);
        is_signed = true;
      } else {
        type = static_cast<Type>(packet.Buffer[0]);
        is_signed = false;
      }

      if (type != Type::RelayPing && type != Type::RelayPong) {
        if (is_signed) {
          LOG(DEBUG, "packet is from network next");
        } else {
          LOG(DEBUG, "packet is not on network next");
        }
        LOG(DEBUG, "incoming packet, type = ", type);
      }

      switch (type) {
        case Type::RelayPing: {
          recorder.InboundPingRx.add(wholePacketSize);
          handlers::relay_ping_handler(packet, recorder, socket, should_handle);
        } break;
        case Type::RelayPong: {
          recorder.PongRx.add(wholePacketSize);
          handlers::relay_pong_handler(packet, relay_manager, should_handle);
        } break;
        case Type::RouteRequest: {
          recorder.RouteRequestRx.add(wholePacketSize);
          handlers::route_request_handler(packet, keychain, session_map, recorder, router_info, socket, is_signed);
        } break;
        case Type::RouteResponse: {
          recorder.RouteResponseRx.add(wholePacketSize);
          handlers::route_response_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case Type::ContinueRequest: {
          recorder.ContinueRequestRx.add(wholePacketSize);
          handlers::continue_request_handler(packet, session_map, keychain, recorder, router_info, socket, is_signed);
        } break;
        case Type::ContinueResponse: {
          recorder.ContinueResponseRx.add(wholePacketSize);
          handlers::continue_response_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case Type::ClientToServer: {
          recorder.ClientToServerRx.add(wholePacketSize);
          handlers::client_to_server_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case Type::ServerToClient: {
          recorder.ServerToClientRx.add(wholePacketSize);
          handlers::server_to_client_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case Type::SessionPing: {
          recorder.SessionPingRx.add(wholePacketSize);
          handlers::session_ping_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case Type::SessionPong: {
          recorder.SessionPongRx.add(wholePacketSize);
          handlers::session_pong_handler(packet, session_map, recorder, router_info, socket, is_signed);
        } break;
        case Type::NearPing: {
          recorder.NearPingRx.add(wholePacketSize);
          handlers::near_ping_handler(packet, recorder, socket, is_signed);
        } break;
        default: {
          recorder.UnknownRx.add(wholePacketSize);
        } break;
      }
    }
  }
}  // namespace core
