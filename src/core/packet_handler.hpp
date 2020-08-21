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
#include "relay/relay.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "token.hpp"
#include "util/macros.hpp"

using core::packets::Type;
using crypto::Keychain;
using os::Socket;
using util::ThroughputRecorder;
using core::RelayPingPacket;

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

    void handle_packet(GenericPacket<>& packet);
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
    core::GenericPacket<> packet;

    while (!this->socket.closed() && this->should_receive) {
      if (!this->socket.recv(packet)) {
        LOG(ERROR, "failed to receive packet");
        continue;
      }

      this->handle_packet(packet);
    }
  }

  /*
   * Switch based on packet type.
   *
   * If the relay is shutting down only reject ping packets
   *
   * This is so that other relays stop receiving proper stats and this one
   * is slowly removed from route decisions
   *
   * However to not disrupt player experience the remaining packets are still
   * handled until the global killswitch is flagged
   */
  INLINE void PacketHandler::handle_packet(GenericPacket<>& packet)
  {
    size_t headerBytes = 0;

    if (packet.Addr.Type == net::AddressType::IPv4) {
      headerBytes = net::IPv4UDPHeaderSize;
    } else if (packet.Addr.Type == net::AddressType::IPv6) {
      headerBytes = net::IPv6UDPHeaderSize;
    }

    size_t wholePacketSize = packet.Len + headerBytes;

    Type type;

    bool isSigned;
    if (crypto::IsNetworkNextPacket(packet.Buffer, packet.Len)) {
      type = static_cast<Type>(packet.Buffer[crypto::PacketHashLength]);
      isSigned = true;
    } else {
      // TODO uncomment below once all packets coming through have the hash
      // return;
      type = static_cast<Type>(packet.Buffer[0]);
      isSigned = false;
    }

    if (type != Type::RelayPing && type != Type::RelayPong) {
      if (isSigned) {
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

        if (packet.Len == RelayPingPacket::BYTE_SIZE) {
          this->recorder.InboundPingRx.add(wholePacketSize);

          handlers::NewRelayPingHandler handler(packet, mRecorder);

          handler.handle(outputBuff, mSocket);
        } else {
          mRecorder.UnknownRx.add(wholePacketSize);
        }
      } break;
      case Type::RelayPong: {
        if (packet.Len == packets::NewRelayPingPacket::ByteSize) {
          mRecorder.PongRx.add(wholePacketSize);

          handlers::NewRelayPongHandler handler(packet, mRelayManager);

          handler.handle();
        } else {
          mRecorder.UnknownRx.add(wholePacketSize);
        }
      } break;
      case packets::Type::RouteRequest: {
        mRecorder.RouteRequestRx.add(wholePacketSize);

        handlers::RouteRequestHandler handler(packet, packet.Addr, mKeychain, mSessionMap, mRecorder, mRouterInfo);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::RouteResponse: {
        mRecorder.RouteResponseRx.add(wholePacketSize);

        handlers::RouteResponseHandler handler(packet, mSessionMap, mRecorder);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::ContinueRequest: {
        mRecorder.ContinueRequestRx.add(wholePacketSize);

        handlers::ContinueRequestHandler handler(packet, mSessionMap, mKeychain, mRecorder, mRouterInfo);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::ContinueResponse: {
        mRecorder.ContinueResponseRx.add(wholePacketSize);

        handlers::ContinueResponseHandler handler(packet, mSessionMap, mRecorder);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::ClientToServer: {
        mRecorder.ClientToServerRx.add(wholePacketSize);

        handlers::ClientToServerHandler handler(packet, mSessionMap, mRecorder);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::ServerToClient: {
        mRecorder.ServerToClientRx.add(wholePacketSize);

        handlers::ServerToClientHandler handler(packet, mSessionMap, mRecorder);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::SessionPing: {
        mRecorder.SessionPingRx.add(wholePacketSize);

        handlers::SessionPingHandler handler(packet, mSessionMap, mRecorder);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::SessionPong: {
        mRecorder.SessionPongRx.add(wholePacketSize);

        handlers::SessionPongHandler handler(packet, mSessionMap, mRecorder);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::NearPing: {
        mRecorder.NearPingRx.add(wholePacketSize);

        handlers::NearPingHandler handler(packet, mRecorder);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      default: {
        mRecorder.UnknownRx.add(wholePacketSize);
      } break;
    }
  }

  INLINE bool PacketProcessor::getAddrFromMsgHdr(net::Address& addr, const msghdr& hdr) const
  {
    bool retval = false;
    auto sockad = reinterpret_cast<sockaddr*>(hdr.msg_name);

    switch (sockad->sa_family) {
      case AF_INET: {
        auto sin = reinterpret_cast<sockaddr_in*>(sockad);
        addr = *sin;
        retval = true;
      } break;
      case AF_INET6: {
        auto sin = reinterpret_cast<sockaddr_in6*>(sockad);
        addr = *sin;
        retval = true;
      } break;
    }

    return retval;
  }
}  // namespace core
