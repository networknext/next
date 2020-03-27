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
  const uint8_t IPv4UDPHeaderSize = 28;
  const uint8_t IPv6UDPHeaderSize = 48;
}  // namespace

namespace core
{
  PacketProcessor::PacketProcessor(
   const std::atomic<bool>& shouldReceive,
   os::Socket& socket,
   const util::Clock& relayClock,
   const crypto::Keychain& keychain,
   core::SessionMap& sessions,
   core::RelayManager& relayManager,
   const volatile bool& handle,
   util::ThroughputLogger& logger,
   const net::Address& receivingAddr)
   : mShouldReceive(shouldReceive),
     mSocket(socket),
     mRelayClock(relayClock),
     mKeychain(keychain),
     mSessionMap(sessions),
     mRelayManager(relayManager),
     mShouldProcess(handle),
     mLogger(logger),
     mRecvAddr(receivingAddr)
  {}

  void PacketProcessor::process(std::condition_variable& var, std::atomic<bool>& readyToReceive)
  {
    static std::atomic<int> listenCounter;
    int listenIndx = listenCounter.fetch_add(1);
    (void)listenIndx;

    GenericPacketBuffer<MaxPacketsToReceive> inputBuffer;
    GenericPacketBuffer<MaxPacketsToReceive> outputBuffer;

    LogDebug("listening for packets {", listenIndx, '}');

    readyToReceive = true;
    var.notify_one();

    while (mShouldReceive) {
      if (!mSocket.multirecv(inputBuffer)) {
        Log("failed to recv packets");
      }

      // LogDebug("got packets on {", listenIndx, "} / count: ", inputBuffer.Count);

      for (int i = 0; i < inputBuffer.Count; i++) {
        getAddrFromMsgHdr(inputBuffer.Packets[i].Addr, inputBuffer.Headers[i].msg_hdr);
        processPacket(inputBuffer.Packets[i], inputBuffer.Headers[i], outputBuffer);
      }

      if (outputBuffer.Count > 0) {
        mSocket.multisend(outputBuffer);
        outputBuffer.Count = 0;
      }
    }
  }

  /*
   * Some handlers in here take a object and a function pointer of that object as an argument
   * the only purpose to that was so that different objects that are responsable for sending packets
   * can be easilly swapped out for benchmarking purposes, once there is a definite solution to the
   * throughput problem the function params will be written strictly
   */

  inline void PacketProcessor::processPacket(
   GenericPacket<>& packet, mmsghdr& header, GenericPacketBuffer<MaxPacketsToSend>& outputBuff)
  {
    LogDebug("packet type: ", static_cast<unsigned int>(packet.Buffer[0]));

    packet.Len = header.msg_len;

    size_t headerBytes = 0;

    if (packet.Addr.Type == net::AddressType::IPv4) {
      headerBytes = IPv4UDPHeaderSize;
    } else if (packet.Addr.Type == net::AddressType::IPv6) {
      headerBytes = IPv6UDPHeaderSize;
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
    switch (packet.Buffer[0]) {
      case RELAY_PING_PACKET: {
        if (!mShouldProcess) {
          Log("relay in process of shutting down, rejecting relay ping packet");
          return;
        }

        if (packet.Len == RELAY_PING_PACKET_BYTES) {
          LogDebug("got relay ping packet");
          mLogger.addToRelayPingPacket(packet.Len + headerBytes);

          handlers::RelayPingHandler handler(packet, packet.Len, mSocket, mRecvAddr);

          handler.handle();
        }
      } break;
      case RELAY_PONG_PACKET: {
        if (packet.Len == RELAY_PING_PACKET_BYTES) {
          LogDebug("got relay pong packet");
          mLogger.addToRelayPongPacket(packet.Len + headerBytes);

          handlers::RelayPongHandler handler(packet, packet.Len, mRelayManager);

          handler.handle();
        }
      } break;
      case RELAY_ROUTE_REQUEST_PACKET: {
        mLogger.addToRouteReq(packet.Len + headerBytes);

        handlers::RouteRequestHandler handler(mRelayClock, packet, packet.Len, packet.Addr, mKeychain, mSessionMap);

        handler.handle(outputBuff, &core::GenericPacketBuffer<1024UL>::push);
      } break;
      case RELAY_ROUTE_RESPONSE_PACKET: {
        mLogger.addToRouteResp(packet.Len + headerBytes);

        LogDebug("got route response from ", packet.Addr);

        handlers::RouteResponseHandler handler(packet, packet.Len, mSessionMap);

        // handler.handle(mSender, &decltype(mSender)::queue);
        handler.handle(outputBuff, &core::GenericPacketBuffer<1024UL>::push);
      } break;
      case RELAY_CONTINUE_REQUEST_PACKET: {
        mLogger.addToContReq(packet.Len + headerBytes);

        handlers::ContinueRequestHandler handler(mRelayClock, packet, packet.Len, mSessionMap, mKeychain);

        handler.handle(outputBuff, &core::GenericPacketBuffer<1024UL>::push);
      } break;
      case RELAY_CONTINUE_RESPONSE_PACKET: {
        mLogger.addToContResp(packet.Len + headerBytes);

        handlers::ContinueResponseHandler handler(packet, packet.Len, mSessionMap);

        handler.handle(outputBuff, &core::GenericPacketBuffer<1024UL>::push);
      } break;
      case RELAY_CLIENT_TO_SERVER_PACKET: {
        LogDebug("got client to server packet");
        mLogger.addToCliToServ(packet.Len + headerBytes);

        handlers::ClientToServerHandler handler(packet, packet.Len, mSessionMap);

        handler.handle(outputBuff, &core::GenericPacketBuffer<1024UL>::push);
      } break;
      case RELAY_SERVER_TO_CLIENT_PACKET: {
        LogDebug("got server to client packet");
        mLogger.addToServToCli(packet.Len + headerBytes);

        handlers::ServerToClientHandler handler(packet, packet.Len, mSessionMap);

        handler.handle(outputBuff, &core::GenericPacketBuffer<1024UL>::push);
      } break;
      case RELAY_SESSION_PING_PACKET: {
        mLogger.addToSessionPing(packet.Len + headerBytes);

        handlers::SessionPingHandler handler(packet, packet.Len, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_SESSION_PONG_PACKET: {
        mLogger.addToSessionPong(packet.Len + headerBytes);

        handlers::SessionPongHandler handler(packet, packet.Len, mSessionMap, mSocket);

        handler.handle();
      } break;
      case RELAY_NEAR_PING_PACKET: {
        mLogger.addToNearPing(packet.Len + headerBytes);

        handlers::NearPingHandler handler(packet, packet.Len, packet.Addr, mSocket);

        handler.handle();
      } break;
      default: {
        LogDebug("received unknown packet type: ", std::hex, (int)packet.Buffer[0], std::dec);
        mLogger.addToUnknown(packet.Len + headerBytes);
      } break;
    }
  }
}  // namespace core
