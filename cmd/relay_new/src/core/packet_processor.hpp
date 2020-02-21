#ifndef CORE_PACKET_PROCESSOR_HPP
#define CORE_PACKET_PROCESSOR_HPP

#include "session.hpp"

#include "core/relay_manager.hpp"
#include "core/router_info.hpp"

#include "crypto/keychain.hpp"

#include "os/platform.hpp"

#include "util/throughput_logger.hpp"

#include "relay/relay_route_token.hpp"
#include "relay/relay_continue_token.hpp"

namespace core
{
  class PacketProcessor
  {
   public:
    PacketProcessor(const util::Clock& relayClock,
     const crypto::Keychain& keychain,
     const core::RouterInfo& routerInfo,
     core::SessionMap& sessions,
     core::RelayManager& relayManager,
     volatile bool& handle,
     util::ThroughputLogger* logger);
    ~PacketProcessor() = default;

    void listen(os::Socket& socket, std::condition_variable& var, std::atomic<bool>& readyToReceive);

   private:
    const util::Clock& mRelayClock;
    const crypto::Keychain& mKeychain;
    const core::RouterInfo mRouterInfo;
    core::SessionMap& mSessionMap;
    core::RelayManager& mRelayManager;
    volatile bool& mShouldProcess;

    util::ThroughputLogger* mLogger;

    // Marks the first byte as a pong packet and sends it back
    void handleRelayPingPacket(
     os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from);

    // Processes the pong packet by increasing the sequence number and getting the time diff
    void handleRelayPongPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from);

    void handleRouteRequestPacket(
     os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from);

    void handleRouteResponsePacket(
     os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from);

    void handleContinueRequestPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleContinueResponsePacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleClientToServerPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleServerToClientPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleSessionPingPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleSessionPongPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleNearPingPacket(
     os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, net::Address& from);

    auto timestamp() -> uint64_t;
    auto tokenIsExpired(relay::relay_route_token_t& token) -> bool;
    auto tokenIsExpired(relay::relay_continue_token_t& token) -> bool;
    auto sessionIsExpired(core::SessionPtr session) -> bool;
  };

  inline auto PacketProcessor::timestamp() -> uint64_t
  {
    auto seconds_since_initialize = mRelayClock.elapsed<util::Second>();
    return mRouterInfo.InitalizeTimeInSeconds + seconds_since_initialize;
  }

  inline auto PacketProcessor::tokenIsExpired(relay::relay_route_token_t& token) -> bool
  {
    return token.expire_timestamp < timestamp();
  }

  inline auto PacketProcessor::tokenIsExpired(relay::relay_continue_token_t& token) -> bool
  {
    return token.expire_timestamp < timestamp();
  }

  inline auto PacketProcessor::sessionIsExpired(core::SessionPtr session) -> bool
  {
    return session->ExpireTimestamp < timestamp();
  }
}  // namespace core
#endif