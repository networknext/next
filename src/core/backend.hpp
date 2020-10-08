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

using crypto::GenericKey;
using crypto::KEY_SIZE;

// forward declare test names to allow private functions to be visible them
namespace testing
{
  class _test_core_Backend_update_valid_;
  class _test_core_Backend_update_shutting_down_true_;
}  // namespace testing

namespace core
{
  const uint32_t INIT_REQUEST_MAGIC = 0x9083708f;

  const uint32_t INIT_REQUEST_VERSION = 1;
  const uint32_t INIT_RESPONSE_VERSION = 0;
  const uint32_t UPDATE_REQUEST_VERSION = 2;
  const uint32_t UPDATE_RESPONSE_VERSION = 0;

  const uint8_t MAX_UPDATE_ATTEMPTS = 11;  // 1 initial + 10 more for failures

  // | magic | version | nonce | address | encrypted token | relay version |
  struct InitRequest
  {
    uint32_t magic = INIT_REQUEST_MAGIC;
    uint32_t version = INIT_REQUEST_VERSION;
    std::array<uint8_t, crypto_box_NONCEBYTES> nonce;
    std::string address;
    std::array<uint8_t, RELAY_TOKEN_BYTES + crypto_box_MACBYTES> encrypted_token;
    std::string relay_version = RELAY_VERSION;

    auto size() -> size_t;
    auto into(std::vector<uint8_t>& v) -> bool;
    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  struct InitResponse
  {
    static const size_t SIZE_OF = 4 + 8 + KEY_SIZE;
    uint32_t version;
    uint64_t timestamp;
    crypto::GenericKey public_key;

    auto into(std::vector<uint8_t>& v) -> bool;
    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  struct UpdateRequest
  {
    uint32_t version = UPDATE_REQUEST_VERSION;
    std::string address;
    GenericKey public_key;
    RelayStats ping_stats;
    uint64_t session_count;
    uint64_t outbound_ping_tx;
    uint64_t route_request_rx;
    uint64_t route_request_tx;
    uint64_t route_response_rx;
    uint64_t route_response_tx;
    uint64_t client_to_server_rx;
    uint64_t client_to_server_tx;
    uint64_t server_to_client_rx;
    uint64_t server_to_client_tx;
    uint64_t inbound_ping_rx;
    uint64_t inbound_ping_tx;
    uint64_t pong_rx;
    uint64_t session_ping_rx;
    uint64_t session_ping_tx;
    uint64_t session_pong_rx;
    uint64_t session_pong_tx;
    uint64_t continue_request_rx;
    uint64_t continue_request_tx;
    uint64_t continue_response_rx;
    uint64_t continue_response_tx;
    uint64_t near_ping_rx;
    uint64_t near_ping_tx;
    uint64_t unknown_rx;
    bool shutting_down;
    double cpu_usage;
    double mem_usage;

    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  struct UpdateResponse
  {
    uint32_t version;
    uint64_t timestamp;
    uint32_t num_relays;
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

    enum class UpdateResult
    {
      Success,
      FailureMaxAttemptsReached,
      FailureTimeoutReached,
      FailureOther,
    };

   public:
    Backend(
     const std::string hostname,
     const std::string address,
     const crypto::Keychain& keychain,
     RouterInfo& router_info,
     RelayManager& relay_manager,
     std::string base64_relay_public_key,
     const core::SessionMap& sessions,
     net::IHttpClient& client);
    ~Backend() = default;

    auto init() -> bool;

    /*
     * Updates the relay in a loop once per second until should_loop is false
     * Returns true as long as the relay doesn't reach the max number of failed update attempts
     */
    auto update_loop(
     volatile bool& should_loop,
     const volatile bool& should_clean_shutdown,
     util::ThroughputRecorder& logger,
     core::SessionMap& sessions) -> bool;

   private:
    const std::string hostname;
    const std::string relay_address;
    const crypto::Keychain& keychain;
    RouterInfo& router_info;
    RelayManager& relay_manager;
    const std::string base64_relay_public_key;
    const core::SessionMap& session_map;
    net::IHttpClient& http_client;

    // this is the public key in the actual backend
    // for reference backends it's random so has to
    // be saved from the init response
    GenericKey update_token;

    auto update(util::ThroughputRecorder& recorder, bool shutdown, const volatile bool& should_retry) -> UpdateResult;
  };
}  // namespace core
