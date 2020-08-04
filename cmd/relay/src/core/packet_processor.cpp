#include "includes.h"

#include "core/continue_token.hpp"
#include "core/route_token.hpp"
#include "crypto/hash.hpp"
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
   const crypto::Keychain& keychain,
   SessionMap& sessions,
   RelayManager<Relay>& relayManager,
   RelayManager<V3Relay>& v3RelayManager,
   const volatile bool& handle,
   util::ThroughputRecorder& logger,
   util::Sender<GenericPacket<>>& sender,
   legacy::v3::TrafficStats& stats,
   const uint64_t oldRelayID,
   const std::atomic<legacy::v3::ResponseState>& state,
   const RouterInfo& routerInfo)
   : mShouldReceive(shouldReceive),
     mSocket(socket),
     mKeychain(keychain),
     mSessionMap(sessions),
     mRelayManager(relayManager),
     mV3RelayManager(v3RelayManager),
     mShouldProcess(handle),
     mRecorder(logger),
     mChannel(sender),
     mStats(stats),
     mOldRelayID(oldRelayID),
     mState(state),
     mRouterInfo(routerInfo)
  {}

  void PacketProcessor::process(std::atomic<bool>& readyToReceive)
  {
    static std::atomic<int> listenCounter;
    int listenIndx = listenCounter.fetch_add(1);
    (void)listenIndx;

    readyToReceive = true;

    GenericPacketBuffer<MaxPacketsToReceive> outputBuffer;

#ifdef RELAY_MULTISEND
    GenericPacketBuffer<MaxPacketsToReceive> inputBuffer;
#else
    core::GenericPacket<> pkt;
#endif

    while (!mSocket.closed() && mShouldReceive) {
#ifdef RELAY_MULTISEND
      if (!mSocket.multirecv(inputBuffer)) {
        Log("failed to recv packets");
      }

      for (int i = 0; i < inputBuffer.Count; i++) {
        auto& pkt = inputBuffer.Packets[i];
        auto& header = inputBuffer.Headers[i];
        if (header.msg_len > 0) {
          pkt.Len = header.msg_len;
          getAddrFromMsgHdr(pkt.Addr, header.msg_hdr);
          processPacket(pkt, outputBuffer);
        }
      }

      if (outputBuffer.Count > 0) {
        mSocket.multisend(outputBuffer);
        outputBuffer.Count = 0;
      }
#else
      if (!mSocket.recv(pkt)) {
        Log("failed to receive packet");
        continue;
      }

      processPacket(pkt, outputBuffer);
#endif
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
  inline void PacketProcessor::processPacket(GenericPacket<>& packet, GenericPacketBuffer<MaxPacketsToSend>& outputBuff)
  {
    size_t headerBytes = 0;

    if (packet.Addr.Type == net::AddressType::IPv4) {
      headerBytes = net::IPv4UDPHeaderSize;
    } else if (packet.Addr.Type == net::AddressType::IPv6) {
      headerBytes = net::IPv6UDPHeaderSize;
    }

    size_t wholePacketSize = packet.Len + headerBytes;

    packets::Type type;

    bool isSigned;
    if (crypto::IsNetworkNextPacket(packet.Buffer, packet.Len)) {
      type = static_cast<packets::Type>(packet.Buffer[crypto::PacketHashLength]);
      isSigned = true;
    } else {
      // TODO uncomment below once all packets coming through have the hash
      // return;
      type = static_cast<packets::Type>(packet.Buffer[0]);
      isSigned = false;
    }

    if (
     type != packets::Type::NewRelayPing && type != packets::Type::NewRelayPong && type != packets::Type::OldRelayPing &&
     type != packets::Type::OldRelayPong) {
      if (isSigned) {
        LogDebug("packet is from network next");
      } else {
        LogDebug("packet is not on network next");
      }
      LogDebug("incoming packet, type = ", type);
    }

    switch (type) {
      case packets::Type::NewRelayPing: {
        if (!mShouldProcess) {
          Log("relay in process of shutting down, rejecting relay ping packet");
          return;
        }

        if (packet.Len == packets::NewRelayPingPacket::ByteSize) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::NewRelayPingHandler handler(packet, mRecorder, mStats);

          handler.handle(outputBuff, mSocket);
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::NewRelayPong: {
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
        if (packet.Len == packets::OldRelayPingPacket::ByteSize) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::OldRelayPingHandler handler(packet, mRecorder, mStats, mOldRelayID);

          handler.handle(outputBuff, mSocket);
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::OldRelayPong: {
        if (packet.Len == packets::OldRelayPongPacket::ByteSize) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecMeasurementRx += wholePacketSize;

          handlers::OldRelayPongHandler handler(packet, mV3RelayManager);

          handler.handle();
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::RouteRequest: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::RouteRequestHandler handler(packet, packet.Addr, mKeychain, mSessionMap, mRecorder, mStats, mRouterInfo);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::RouteResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::RouteResponseHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::ContinueRequest: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::ContinueRequestHandler handler(packet, mSessionMap, mKeychain, mRecorder, mStats, mRouterInfo);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::ContinueResponse: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecManagementRx += wholePacketSize;

        handlers::ContinueResponseHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::ClientToServer: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecPaidRx += wholePacketSize;

        handlers::ClientToServerHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::ServerToClient: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecPaidRx += wholePacketSize;

        handlers::ServerToClientHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::SessionPing: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::SessionPingHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::SessionPong: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::SessionPongHandler handler(packet, mSessionMap, mRecorder, mStats);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      case packets::Type::NearPing: {
        mRecorder.addToReceived(wholePacketSize);
        mStats.BytesPerSecMeasurementRx += wholePacketSize;

        handlers::NearPingHandler handler(packet, mRecorder, mStats);

        handler.handle(outputBuff, mSocket, isSigned);
      } break;
      // Next three all do the same thing
      case packets::Type::V3BackendInitResponse: {
        if (mState == legacy::v3::ResponseState::Init) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecManagementRx += wholePacketSize;
          mChannel.send(packet);
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::V3BackendConfigResponse: {
        if (mState == legacy::v3::ResponseState::Config) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecManagementRx += wholePacketSize;
          mChannel.send(packet);
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      case packets::Type::V3BackendUpdateResponse: {
        if (mState == legacy::v3::ResponseState::Update) {
          mRecorder.addToReceived(wholePacketSize);
          mStats.BytesPerSecManagementRx += wholePacketSize;
          mChannel.send(packet);
        } else {
          mRecorder.addToUnknown(wholePacketSize);
          mStats.BytesPerSecInvalidRx += wholePacketSize;
        }
      } break;
      default: {
        mRecorder.addToUnknown(wholePacketSize);
        mStats.BytesPerSecInvalidRx += wholePacketSize;
      } break;
    }
  }
}  // namespace core
