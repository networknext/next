#include "includes.h"

#include "core/continue_token.hpp"
#include "core/route_token.hpp"
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
#include "packet_processor.hpp"
#include "relay/relay.hpp"
#include "relay/relay_platform.hpp"

namespace
{
  const uint8_t V3BackendRelayResponse = 49;
  const uint8_t V3BackendConfigResponse = 51;
  const uint8_t V3BackendInitResponse = 52;
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
   core::RelayManager& v3RelayManager,
   const volatile bool& handle,
   util::ThroughputRecorder& logger,
   const net::Address& receivingAddr,
   util::Sender<core::GenericPacket<>>& sender,
   legacy::v3::TrafficStats& stats)
   : mShouldReceive(shouldReceive),
     mSocket(socket),
     mRelayClock(relayClock),
     mKeychain(keychain),
     mSessionMap(sessions),
     mRelayManager(relayManager),
     mV3RelayManager(v3RelayManager),
     mShouldProcess(handle),
     mRecorder(logger),
     mRecvAddr(receivingAddr),
     mSender(sender),
     mStats(stats)
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

  inline void PacketProcessor::processPacket(
   GenericPacket<>& packet, mmsghdr& header, GenericPacketBuffer<MaxPacketsToSend>& outputBuff)
  {
    packet.Len = header.msg_len;

    size_t headerBytes = 0;

    if (packet.Addr.Type == net::AddressType::IPv4) {
      headerBytes = net::IPv4UDPHeaderSize;
    } else if (packet.Addr.Type == net::AddressType::IPv6) {
      headerBytes = net::IPv6UDPHeaderSize;
    }

    size_t wholePacketSize = packet.Len + headerBytes;

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
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::RelayPingHandler handler(packet, packet.Len, mSocket, mRecvAddr, mRecorder);

          handler.handle();
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case RELAY_PONG_PACKET: {
        if (packet.Len == RELAY_PING_PACKET_BYTES) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::RelayPongHandler handler(packet, packet.Len, mRelayManager, mV3RelayManager);

          handler.handle();
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case RELAY_ROUTE_REQUEST_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::RouteRequestHandler handler(mRelayClock, packet, packet.Len, packet.Addr, mKeychain, mSessionMap, mRecorder);

        handler.handle(outputBuff);
      } break;
      case RELAY_ROUTE_RESPONSE_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::RouteResponseHandler handler(packet, packet.Len, mSessionMap, mRecorder);

        handler.handle(outputBuff);
      } break;
      case RELAY_CONTINUE_REQUEST_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::ContinueRequestHandler handler(mRelayClock, packet, packet.Len, mSessionMap, mKeychain, mRecorder);

        handler.handle(outputBuff);
      } break;
      case RELAY_CONTINUE_RESPONSE_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::ContinueResponseHandler handler(packet, packet.Len, mSessionMap, mRecorder);

        handler.handle(outputBuff);
      } break;
      case RELAY_CLIENT_TO_SERVER_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecPaidRx += wholePacketSize;

        handlers::ClientToServerHandler handler(packet, packet.Len, mSessionMap, mRecorder);

        handler.handle(outputBuff);
      } break;
      case RELAY_SERVER_TO_CLIENT_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecPaidRx += wholePacketSize;

        handlers::ServerToClientHandler handler(packet, packet.Len, mSessionMap, mRecorder);

        handler.handle(outputBuff);
      } break;
      case RELAY_SESSION_PING_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::SessionPingHandler handler(packet, packet.Len, mSessionMap, mSocket, mRecorder);

        handler.handle();
      } break;
      case RELAY_SESSION_PONG_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::SessionPongHandler handler(packet, packet.Len, mSessionMap, mSocket, mRecorder);

        handler.handle();
      } break;
      case RELAY_NEAR_PING_PACKET: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::NearPingHandler handler(packet, packet.Len, packet.Addr, mSocket, mRecorder);

        handler.handle();
      } break;
      // Next three all do the same thing
      case V3BackendInitResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;
        mSender.send(packet);
        LogDebug("got init response, current number of items in channel ", mSender.size());
      } break;
      case V3BackendConfigResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;
        mSender.send(packet);
        LogDebug("got config response, current number of items in channel ", mSender.size());
      } break;
      case V3BackendRelayResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;
        mSender.send(packet);
        LogDebug("got relay response, current number of items in channel ", mSender.size());
      } break;
      default: {
        LogDebug("received unknown packet type: ", std::hex, (int)packet.Buffer[0], std::dec);
        mRecorder.addToUnknown(packet.Len + headerBytes);
      } break;
    }
  }
}  // namespace core
