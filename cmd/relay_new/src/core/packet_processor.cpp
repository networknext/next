#include "includes.h"
#include "packet_processor.hpp"

#include "encoding/read.hpp"
#include "encoding/write.hpp"

#include "relay/relay_platform.hpp"
#include "relay/relay.hpp"

#include "core/route_token.hpp"
#include "core/continue_token.hpp"

#include "handlers/relay_ping_handler.hpp"
#include "handlers/relay_pong_handler.hpp"
#include "handlers/route_request_handler.hpp"
#include "handlers/route_response_handler.hpp"
#include "handlers/continue_request_handler.hpp"
#include "handlers/continue_response_handler.hpp"
#include "handlers/client_to_server_handler.hpp"
#include "handlers/server_to_client_handler.hpp"
#include "handlers/session_ping_handler.hpp"
#include "handlers/session_pong_handler.hpp"
#include "handlers/near_ping_handler.hpp"

namespace
{
  // TODO finish implementing this
  enum class PacketType
  {
    RelayPing = RELAY_PING_PACKET,
    RelayPong = RELAY_PONG_PACKET,
    RouteRequest = RELAY_ROUTE_REQUEST_PACKET,
    RouteResponse = RELAY_ROUTE_RESPONSE_PACKET,
    ContinueRequest = RELAY_CONTINUE_REQUEST_PACKET,
    ContinueResponse = RELAY_CONTINUE_RESPONSE_PACKET,
    ClientToServer = RELAY_CLIENT_TO_SERVER_PACKET,
    ServerToClient = RELAY_SERVER_TO_CLIENT_PACKET,
    SessionPing = RELAY_SESSION_PING_PACKET,
    SessionPong = RELAY_SESSION_PONG_PACKET,
    NearPing = RELAY_NEAR_PING_PACKET,
    NearPong = RELAY_NEAR_PONG_PACKET
  };
}  // namespace

namespace core
{
  PacketProcessor::PacketProcessor(os::Socket& socket,
   const util::Clock& relayClock,
   const crypto::Keychain& keychain,
   const core::RouterInfo& routerInfo,
   core::SessionMap& sessions,
   core::RelayManager& relayManager,
   volatile bool& handle,
   util::ThroughputLogger* logger)
   : mSocket(socket),
     mRelayClock(relayClock),
     mKeychain(keychain),
     mRouterInfo(routerInfo),
     mSessionMap(sessions),
     mRelayManager(relayManager),
     mShouldProcess(handle),
     mLogger(logger),
     mSender(socket)
  {}

  void PacketProcessor::process(std::condition_variable& var, std::atomic<bool>& readyToReceive)
  {
    static std::atomic<int> listenCounter;
    int listenIndx = listenCounter.fetch_add(1);
    (void)listenIndx;

    GenericPacket rawPacket;

    std::vector<net::Message> messages;
    messages.resize(100);

    LogDebug("listening for packets {", listenIndx, '}');

    readyToReceive = true;
    var.notify_one();

    while (this->mShouldProcess) {
      if (false) {
        net::Address from;
        const int packetSize = mSocket.recv(from, rawPacket.data(), sizeof(uint8_t) * rawPacket.size());

        LogDebug("got packet on {", listenIndx, "} / type: ", static_cast<unsigned int>(rawPacket[0]));

        processPacket(rawPacket, packetSize, from);

      } else {
        auto count = mSocket.multirecv(messages);

        LogDebug("got packets on {", listenIndx, "}, / count: ", count);

        for (size_t i = 0; i < count; i++) {
          auto& msg = messages[i];
          if (msg.Len > rawPacket.size()) {
            continue;
          }

          std::copy(msg.Data.begin(), msg.Data.begin() + msg.Len, rawPacket.begin());
          processPacket(rawPacket, msg.Len, msg.Addr);
        }
      }
    }
  }

  inline void PacketProcessor::processPacket(GenericPacket& rawPacket, int packetSize, const net::Address& from)
  {
    LogDebug("packet type: ", static_cast<unsigned int>(rawPacket[0]));
    switch (rawPacket[0]) {
      case RELAY_PING_PACKET: {
        if (packetSize == RELAY_PING_PACKET_BYTES) {
          LogDebug("got relay ping packet");
          if (mLogger != nullptr) {
            mLogger->addToRelayPingPacket(packetSize);
          }

          handlers::RelayPingHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, mSocket);

          handler.handle();
        }
      } break;
      case RELAY_PONG_PACKET: {
        if (packetSize == RELAY_PING_PACKET_BYTES) {
          LogDebug("got relay pong packet");
          if (mLogger != nullptr) {
            mLogger->addToRelayPongPacket(packetSize);
          }

          handlers::RelayPongHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, mRelayManager);

          handler.handle();
        }
      } break;
      case RELAY_ROUTE_REQUEST_PACKET: {
        if (mLogger != nullptr) {
          mLogger->addToRouteReq(packetSize);
        }

        handlers::RouteRequestHandler handler(
         mRelayClock, mRouterInfo, rawPacket, packetSize, from, mKeychain, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_ROUTE_RESPONSE_PACKET: {
        if (mLogger != nullptr) {
          mLogger->addToRouteResp(packetSize);
        }

        LogDebug("got route response from ", from);

        handlers::RouteResponseHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_CONTINUE_REQUEST_PACKET: {
        if (mLogger != nullptr) {
          mLogger->addToContReq(packetSize);
        }

        handlers::ContinueRequestHandler handler(
         mRelayClock, mRouterInfo, rawPacket, packetSize, mSessionMap, mSocket, mKeychain);

        handler.handle();
      } break;
      case RELAY_CONTINUE_RESPONSE_PACKET: {
        if (mLogger != nullptr) {
          mLogger->addToContResp(packetSize);
        }

        handlers::ContinueResponseHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_CLIENT_TO_SERVER_PACKET: {
        LogDebug("got client to server packet");
        if (mLogger != nullptr) {
          mLogger->addToCliToServ(packetSize);
        }

        handlers::ClientToServerHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_SERVER_TO_CLIENT_PACKET: {
        LogDebug("got server to client packet");
        if (mLogger != nullptr) {
          mLogger->addToServToCli(packetSize);
        }

        handlers::ServerToClientHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_SESSION_PING_PACKET: {
        if (mLogger != nullptr) {
          mLogger->addToSessionPing(packetSize);
        }

        handlers::SessionPingHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_SESSION_PONG_PACKET: {
        if (mLogger != nullptr) {
          mLogger->addToSessionPong(packetSize);
        }

        handlers::SessionPongHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_NEAR_PING_PACKET: {
        if (mLogger != nullptr) {
          mLogger->addToNearPing(packetSize);
        }

        handlers::NearPingHandler handler(mRelayClock, mRouterInfo, rawPacket, packetSize, from, mSocket);

        handler.handle();
      } break;
      default: {
        LogDebug("received unknown packet type: ", std::hex, (int)rawPacket[0]);
        if (mLogger != nullptr) {
          mLogger->addToUnknown(packetSize);
        }
      } break;
    }
  }
}  // namespace core
