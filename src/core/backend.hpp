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
#include "util/clock.hpp"

using core::SessionMap;
using crypto::GenericKey;
using crypto::KEY_SIZE;
using crypto::Keychain;
using crypto::Nonce;
using net::CurlWrapper;
using util::Clock;
using util::ThroughputRecorder;

// forward declare test names to allow private functions to be visible them
namespace testing
{
  class _test_core_Backend_update_valid_;
  class _test_core_Backend_update_shutting_down_true_;
}  // namespace testing

namespace core
{
  extern const char* RELAY_VERSION;

  const uint32_t UPDATE_REQUEST_VERSION = 5;
  const uint32_t UPDATE_RESPONSE_VERSION = 0;

  const uint8_t MAX_UPDATE_ATTEMPTS = 11;  // 1 initial + 10 more for failures

  struct UpdateRequest
  {
    uint32_t version = UPDATE_REQUEST_VERSION;
    std::string address;
    GenericKey public_key;
    RelayStats ping_stats;
    uint64_t session_count;
    bool shutting_down;
    std::string relay_version = RELAY_VERSION;
    uint8_t cpu_usage;
    uint64_t envelope_up;
    uint64_t envelope_down;
    double bandwidth_sent;
    double bandwidth_recv;

    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  struct UpdateResponse
  {
    uint32_t version;
    uint64_t timestamp;
    uint32_t num_relays;
    std::array<RelayPingInfo, MAX_RELAYS> relays;
    std::string target_version;

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

    enum class UpdateResult
    {
      Success,
      Failure,
    };

   public:
    Backend(
     const std::string hostname,
     const std::string address,
     const Keychain& keychain,
     RouterInfo& router_info,
     RelayManager& relay_manager,
     const SessionMap& sessions,
     CurlWrapper & curl);
    ~Backend() = default;

    auto init() -> bool;

    /*
     * Updates the relay in a loop once per second until should_loop is false
     * Returns true as long as the relay doesn't reach the max number of failed update attempts
     */
    auto update_loop(
     volatile bool& should_loop, const volatile bool& should_clean_shutdown, ThroughputRecorder& logger, SessionMap& sessions)
     -> bool;

   private:
    const std::string hostname;
    const std::string relay_address;
    const Keychain& keychain;
    RouterInfo& router_info;
    RelayManager& relay_manager;
    const SessionMap& session_map;
    CurlWrapper& http_client;

    // this is the public key in the actual backend
    // for reference backends it's random so has to
    // be saved from the init response
    GenericKey update_token;

    auto update(util::ThroughputRecorder& recorder, bool shutdown, const volatile bool& should_retry) -> UpdateResult;
  };
}  // namespace core
