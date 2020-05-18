#include "includes.h"

#include "core/continue_token.hpp"
#include "core/route_token.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "handlers/client_to_server_handler.hpp"
#include "handlers/continue_request_handler.hpp"
#include "handlers/continue_response_handler.hpp"
#include "handlers/near_ping_handler.hpp"
#include "handlers/new_relay_ping_handler.hpp"
#include "handlers/new_relay_pong_handler.hpp"
#include "handlers/old_relay_ping_handler.hpp"
#include "handlers/old_relay_pong_handler.hpp"
#include "handlers/route_request_handler.hpp"
#include "handlers/route_response_handler.hpp"
#include "handlers/server_to_client_handler.hpp"
#include "handlers/session_ping_handler.hpp"
#include "handlers/session_pong_handler.hpp"
#include "packet_processor.hpp"
#include "packets/types.hpp"
#include "relay/relay.hpp"
#include "relay/relay_platform.hpp"

namespace core
{
  PacketProcessor::PacketProcessor(
   const std::atomic<bool>& shouldReceive,
   os::Socket& socket,
   const util::Clock& relayClock,
   const crypto::Keychain& keychain,
   SessionMap& sessions,
   RelayManager<Relay>& relayManager,
   RelayManager<V3Relay>& v3RelayManager,
   const volatile bool& handle,
   util::ThroughputRecorder& logger,
   const net::Address& receivingAddr,
   util::Sender<GenericPacket<>>& sender,
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
     mChannel(sender),
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
    switch (static_cast<packets::Type>(packet.Buffer[0])) {
      case packets::Type::NewRelayPing: {
        if (!mShouldProcess) {
          Log("relay in process of shutting down, rejecting relay ping packet");
          return;
        }

        LogDebug("new ping");
        if (packet.Len == packets::NewRelayPingPacket::ByteSize) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::NewRelayPingHandler handler(packet, mSocket, mRecvAddr, mRecorder, mStats);

          handler.handle();
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::NewRelayPong: {
        LogDebug("new pong");
        if (packet.Len == packets::NewRelayPingPacket::ByteSize) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::NewRelayPongHandler handler(packet, mRelayManager);

          handler.handle();
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::OldRelayPing: {
        LogDebug("old ping");
        if (packet.Len == packets::OldRelayPingPacket::ByteSize) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::OldRelayPingHandler handler(packet, mSocket, mRecorder, mStats);

          handler.handle();
        } else {
          LogDebug("got invalid old ping packet from ", packet.Addr);
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::OldRelayPong: {
        LogDebug("old pong");
        if (packet.Len == packets::OldRelayPongPacket::ByteSize) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::OldRelayPongHandler handler(packet, mV3RelayManager);

          handler.handle();
        } else {
          LogDebug("got invalid old pong packet from ", packet.Addr);
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::RouteRequest: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::RouteRequestHandler handler(mRelayClock, packet, packet.Addr, mKeychain, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff);
      } break;
      case packets::Type::RouteResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::RouteResponseHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff);
      } break;
      case packets::Type::ContinueRequest: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::ContinueRequestHandler handler(mRelayClock, packet, mSessionMap, mKeychain, mRecorder, mStats);

        handler.handle(outputBuff);
      } break;
      case packets::Type::ContinueResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::ContinueResponseHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff);
      } break;
      case packets::Type::ClientToServer: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecPaidRx += wholePacketSize;

        handlers::ClientToServerHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff);
      } break;
      case packets::Type::ServerToClient: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecPaidRx += wholePacketSize;

        handlers::ServerToClientHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff);
      } break;
      case packets::Type::SessionPing: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::SessionPingHandler handler(packet, mSessionMap, mSocket, mRecorder, mStats);

        handler.handle();
      } break;
      case packets::Type::SessionPong: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::SessionPongHandler handler(packet, mSessionMap, mSocket, mRecorder, mStats);

        handler.handle();
      } break;
      case packets::Type::NearPing: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::NearPingHandler handler(packet, packet.Addr, mSocket, mRecorder, mStats);

        handler.handle();
      } break;
      // Next three all do the same thing
      case packets::Type::V3BackendInitResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;
        mChannel.send(packet);
        LogDebug("got init response, current number of items in channel ", mChannel.size());
      } break;
      case packets::Type::V3BackendConfigResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;
        mChannel.send(packet);
        LogDebug("got config response, current number of items in channel ", mChannel.size());
      } break;
      case packets::Type::V3BackendUpdateResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;
        mChannel.send(packet);
        LogDebug("got relay response, current number of items in channel ", mChannel.size());
      } break;
      default: {
        LogDebug("received unknown packet type: ", std::hex, (int)packet.Buffer[0], std::dec);
        mRecorder.addToUnknown(packet.Len + headerBytes);
      } break;
    }
  }
}  // namespace core
