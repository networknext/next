#pragma once

#include "core/continue_token.hpp"
#include "core/route_token.hpp"
#include "core/throughput_recorder.hpp"
#include "crypto/hash.hpp"
#include "crypto/keychain.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "handlers/client_to_server_handler.hpp"
#include "handlers/continue_request_handler.hpp"
#include "handlers/continue_response_handler.hpp"
#include "handlers/near_ping_handler.hpp"
#include "handlers/new_relay_ping_handler.hpp"
#include "handlers/new_relay_pong_handler.hpp"
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

using core::packets::Type;
using core::packets::RELAY_PING_PACKET_SIZE;
using crypto::Keychain;
using os::Socket;
using util::ThroughputRecorder;

namespace core
{
  class PacketHandler
  {
   public:
    PacketHandler(
     const std::atomic<bool>& shouldReceive,
     Socket& socket,
     const crypto::Keychain& keychain,
     SessionMap& sessions,
     RelayManager& relayManager,
     const volatile bool& handle,
     ThroughputRecorder& recorder,
     const RouterInfo& routerInfo);
    ~PacketHandler() = default;

    void handle_packets();

   private:
    const std::atomic<bool>& should_receive;
    const Socket& socket;
    const Keychain& keychain;
    SessionMap& session_map;
    RelayManager& relay_manager;
    const volatile bool& should_process;
    ThroughputRecorder& recorder;
    const RouterInfo& router_info;

    void handle_packet(Packet& packet);
  };

  INLINE PacketHandler::PacketHandler(
   const std::atomic<bool>& should_receive,
   Socket& socket,
   const crypto::Keychain& keychain,
   SessionMap& sessions,
   RelayManager& relay_manager,
   const volatile bool& loop_handle,
   ThroughputRecorder& recorder,
   const RouterInfo& router_info)
   : should_receive(should_receive),
     socket(socket),
     keychain(keychain),
     session_map(sessions),
     relay_manager(relay_manager),
     should_process(loop_handle),
     recorder(recorder),
     router_info(router_info)
  {}

  INLINE void PacketHandler::handle_packets()
  {
    core::Packet packet;

    while (!this->socket.closed() && this->should_receive) {
      if (!this->socket.recv(packet)) {
        LOG(ERROR, "failed to receive packet");
        continue;
      }

      this->handle_packet(packet);
    }
  }

  INLINE void PacketHandler::handle_packet(Packet& packet)
  {
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
        if (!this->should_process) {
          LOG(INFO, "relay in process of shutting down, rejecting relay ping packet");
          return;
        }
        if (packet.Len == RELAY_PING_PACKET_SIZE) {
          this->recorder.InboundPingRx.add(wholePacketSize);
          handlers::relay_ping_handler(packet, this->recorder, this->socket);
        } else {
          this->recorder.UnknownRx.add(wholePacketSize);
        }
      } break;
      case Type::RelayPong: {
        if (!this->should_process) {
          LOG(INFO, "relay in process of shutting down, rejecting relay pong packet");
          return;
        }
        if (packet.Len == RELAY_PING_PACKET_SIZE) {
          this->recorder.PongRx.add(wholePacketSize);
          handlers::relay_pong_handler(packet, this->relay_manager);
        } else {
          this->recorder.UnknownRx.add(wholePacketSize);
        }
      } break;
      case packets::Type::RouteRequest: {
        this->recorder.RouteRequestRx.add(wholePacketSize);
        handlers::route_request_handler(
         packet, this->keychain, this->session_map, this->recorder, this->router_info, this->socket, is_signed);
      } break;
      case packets::Type::RouteResponse: {
        this->recorder.RouteResponseRx.add(wholePacketSize);
        handlers::route_response_handler(packet, this->session_map, this->recorder, this->socket, is_signed);
      } break;
      case packets::Type::ContinueRequest: {
        this->recorder.ContinueRequestRx.add(wholePacketSize);
        handlers::continue_request_handler(
         packet, this->session_map, this->keychain, this->recorder, this->router_info, this->socket, is_signed);
      } break;
      case packets::Type::ContinueResponse: {
        this->recorder.ContinueResponseRx.add(wholePacketSize);
        handlers::continue_response_handler(packet, this->session_map, this->recorder, this->socket, is_signed);
      } break;
      case packets::Type::ClientToServer: {
        this->recorder.ClientToServerRx.add(wholePacketSize);
        handlers::client_to_server_handler(packet, this->session_map, this->recorder, this->socket, is_signed);
      } break;
      case packets::Type::ServerToClient: {
        this->recorder.ServerToClientRx.add(wholePacketSize);
        handlers::server_to_client_handler(packet, this->session_map, this->recorder, this->socket, is_signed);
      } break;
      case packets::Type::SessionPing: {
        this->recorder.SessionPingRx.add(wholePacketSize);
        handlers::session_ping_handler(packet, this->session_map, this->recorder, this->socket, is_signed);
      } break;
      case packets::Type::SessionPong: {
        this->recorder.SessionPongRx.add(wholePacketSize);
        handlers::session_pong_handler(packet, this->session_map, this->recorder, this->socket, is_signed);
      } break;
      case packets::Type::NearPing: {
        this->recorder.NearPingRx.add(wholePacketSize);
        handlers::near_ping_handler(packet, this->recorder, this->socket, is_signed);
      } break;
      default: {
        this->recorder.UnknownRx.add(wholePacketSize);
      } break;
    }
  }
}  // namespace core
