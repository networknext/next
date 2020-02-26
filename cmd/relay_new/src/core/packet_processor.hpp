#ifndef CORE_PACKET_PROCESSOR_HPP
#define CORE_PACKET_PROCESSOR_HPP

#include "session.hpp"

#include "core/packet.hpp"
#include "core/relay_manager.hpp"
#include "core/router_info.hpp"
#include "core/token.hpp"

#include "crypto/keychain.hpp"

#include "os/platform.hpp"

#include "util/throughput_logger.hpp"

#include "net/buffered_sender.hpp"

namespace core
{
  class PacketProcessor
  {
   public:
    PacketProcessor(os::Socket& socket,
     const util::Clock& relayClock,
     const crypto::Keychain& keychain,
     const core::RouterInfo& routerInfo,
     core::SessionMap& sessions,
     core::RelayManager& relayManager,
     volatile bool& handle,
     util::ThroughputLogger* logger);
    ~PacketProcessor() = default;

    void process(std::condition_variable& var, std::atomic<bool>& readyToReceive);

    void flushResponses();

   private:
    const os::Socket& mSocket;
    const util::Clock& mRelayClock;
    const crypto::Keychain& mKeychain;
    const core::RouterInfo mRouterInfo;
    core::SessionMap& mSessionMap;
    core::RelayManager& mRelayManager;
    volatile bool& mShouldProcess;

    util::ThroughputLogger* mLogger;

    net::BufferedSender<1, 1> mSender;

    // Marks the first byte as a pong packet and sends it back
    void handleRelayPingPacket(GenericPacket& packet, const int size);

    // Processes the pong packet by increasing the sequence number and getting the time diff
    void handleRelayPongPacket(GenericPacket& packet, const int size);

    void handleRouteRequestPacket(GenericPacket& packet, const int size, net::Address& from);

    void handleRouteResponsePacket(GenericPacket& packet, const int size, net::Address& from);

    void handleContinueRequestPacket(GenericPacket& packet, const int size);

    void handleContinueResponsePacket(GenericPacket& packet, const int size);

    void handleClientToServerPacket(GenericPacket& packet, const int size);

    void handleServerToClientPacket(GenericPacket& packet, const int size);

    void handleSessionPingPacket(GenericPacket& packet, const int size);

    void handleSessionPongPacket(GenericPacket& packet, const int size);

    void handleNearPingPacket(GenericPacket& packet, const int size, net::Address& from);

    auto timestamp() -> uint64_t;

    auto tokenIsExpired(core::Token& token) -> bool;

    auto sessionIsExpired(core::SessionPtr session) -> bool;
  };

  inline void PacketProcessor::flushResponses()
  {
    mSender.autoSend();
  }
}  // namespace core
#endif