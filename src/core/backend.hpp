#pragma once

#include "core/throughput_recorder.hpp"
#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "net/http.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "testing/test.hpp"
#include "util/logger.hpp"

// forward declare test names to allow private functions to be visible them
namespace testing
{
  class _test_core_Backend_update_valid_;
  class _test_core_Backend_update_shutting_down_true_;
}  // namespace testing

namespace core
{
  const uint32_t InitRequestMagic = 0x9083708f;

  const uint32_t InitRequestVersion = 0;

  const uint8_t MaxUpdateAttempts = 11;  // 1 initial + 10 more for failures
  // | magic | version | nonce | address | encrypted token | relay version |
  struct InitRequest
  {
    uint32_t Magic = InitRequestMagic;
    uint32_t Version = InitRequestVersion;
    std::array<uint8_t, crypto_box_NONCEBYTES> Nonce;
    std::string Address;
    std::array<uint8_t, RELAY_TOKEN_BYTES + crypto_box_MACBYTES> EncryptedToken;
    std::string RelayVersion = RELAY_VERSION;

    auto size() -> size_t;
    auto into(std::vector<uint8_t>& v) -> bool;
    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  struct InitResponse
  {
    static const size_t ByteSize = 4 + 8 + crypto::KEY_SIZE;
    uint32_t Version;
    uint64_t Timestamp;
    crypto::GenericKey PublicKey;

    auto into(std::vector<uint8_t>& v) -> bool;
    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  struct UpdateRequest
  {
    uint32_t Version;
    std::string Address;
    std::array<uint8_t, crypto::KEY_SIZE> PublicKey;
    RelayStats PingStats;
    uint64_t SessionCount;
    uint64_t OutboundPingTx;
    uint64_t RouteRequestRx;
    uint64_t RouteRequestTx;
    uint64_t RouteResponseRx;
    uint64_t RouteResponseTx;
    uint64_t ClientToServerRx;
    uint64_t ClientToServerTx;
    uint64_t ServerToClientRx;
    uint64_t ServerToClientTx;
    uint64_t InboundPingRx;
    uint64_t InboundPingTx;
    uint64_t PongRx;
    uint64_t SessionPingRx;
    uint64_t SessionPingTx;
    uint64_t SessionPongRx;
    uint64_t SessionPongTx;
    uint64_t ContinueRequestRx;
    uint64_t ContinueRequestTx;
    uint64_t ContinueResponseRx;
    uint64_t ContinueResponseTx;
    uint64_t NearPingRx;
    uint64_t NearPingTx;
    uint64_t UnknownRx;
    bool ShuttingDown;
    double CPUUsage;
    double MemUsage;
    std::string RelayVersion;

    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  struct UpdateResponse
  {
    uint32_t Version;
    uint64_t Timestamp;
    uint32_t NumRelays;
    std::array<RelayPingInfo, MAX_RELAYS> Relays;

    auto size() -> size_t;
    auto into(std::vector<uint8_t>& v) -> bool;
    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  /*
   * A class that's responsible for backend related tasks
   * where T should be anything that defines a static SendTo function
   * with the same signature as net::BeastWrapper
   */
  class Backend
  {
    friend testing::_test_core_Backend_update_valid_;
    friend testing::_test_core_Backend_update_shutting_down_true_;

   public:
    Backend(
     const std::string hostname,
     const std::string address,
     const crypto::Keychain& keychain,
     RouterInfo& routerInfo,
     RelayManager& relayManager,
     std::string base64RelayPublicKey,
     const core::SessionMap& sessions,
     net::IHttpClient& client);
    ~Backend() = default;

    auto init() -> bool;

    /*
     * Updates the relay in a loop once per second until loopHandle is false
     * Returns true as long as the relay doesn't reach the max number of failed update attempts
     */
    auto updateCycle(
     const volatile bool& loopHandle,
     const volatile bool& shouldCleanShutdown,
     util::ThroughputRecorder& logger,
     core::SessionMap& sessions) -> bool;

   private:
    const std::string mHostname;
    const std::string mAddressStr;
    const crypto::Keychain& mKeychain;
    RouterInfo& mRouterInfo;
    RelayManager& mRelayManager;
    const std::string mBase64RelayPublicKey;
    const core::SessionMap& mSessionMap;
    net::IHttpClient& mRequester;

    auto update(util::ThroughputRecorder& recorder, bool shutdown) -> bool;
  };
}  // namespace core
