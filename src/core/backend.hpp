#ifndef CORE_BACKEND_HPP
#define CORE_BACKEND_HPP

#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "net/http.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "testing/test.hpp"
#include "util/json.hpp"
#include "util/logger.hpp"
#include "util/throughput_recorder.hpp"
#include "os/platform.hpp"

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
  const uint32_t InitResponseVersion = 0;

  const uint32_t UpdateRequestVersion = 0;
  const uint32_t UpdateResponseVersion = 0;

  const uint8_t MaxUpdateAttempts = 11;  // 1 initial + 10 more for failures

  /*
   * A class that's responsible for backend related tasks
   * where T should be anything that defines a static SendTo function
   * with the same signature as net::CurlWrapper
   */
  template <typename T>
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
     RelayManager<Relay>& relayManager,
     std::string base64RelayPublicKey,
     const core::SessionMap& sessions);
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
    RelayManager<Relay>& mRelayManager;
    const std::string mBase64RelayPublicKey;
    const core::SessionMap& mSessionMap;

    auto update(util::ThroughputRecorder& recorder, bool shutdown) -> bool;
    auto buildInitRequest(util::JSON& doc) -> std::tuple<bool, const char*>;
    auto buildUpdateRequest(util::JSON& doc, util::ThroughputRecorder& recorder, bool shutdown)
     -> std::tuple<bool, const char*>;
  };

  template <typename T>
  Backend<T>::Backend(
   std::string hostname,
   std::string address,
   const crypto::Keychain& keychain,
   RouterInfo& routerInfo,
   RelayManager<Relay>& relayManager,
   std::string base64RelayPublicKey,
   const core::SessionMap& sessions)
   : mHostname(hostname),
     mAddressStr(address),
     mKeychain(keychain),
     mRouterInfo(routerInfo),
     mRelayManager(relayManager),
     mBase64RelayPublicKey(base64RelayPublicKey),
     mSessionMap(sessions)
  {}

  template <typename T>
  auto Backend<T>::init() -> bool
  {
    util::JSON doc;
    auto [ok, err] = buildInitRequest(doc);
    if (!ok) {
      Log("error building init request: ", err);
      return false;
    }
    std::string request = doc.toString();
    std::string response;

    if (!T::SendTo(mHostname, "/relay_init", request, response)) {
      Log("curl request failed in init");
      return false;
    }

    if (!doc.parse(response)) {
      Log("could not parse json response in init: ", doc.err(), "\nResponse: ", std::string(response.begin(), response.end()));
      return false;
    }

    if (doc.memberExists("version")) {
      if (doc.memberIs(util::JSON::Type::Number, "version")) {
        auto version = doc.get<uint32_t>("version");
        if (version != InitResponseVersion) {
          Log("error: bad relay init response version. expected ", InitResponseVersion, ", got ", version);
          return false;
        }
      } else {
        Log("warning, init version response not a number");
      }
    } else {
      Log("warning, version number missing in init response");
    }

    if (doc.memberExists("Timestamp")) {
      if (doc.memberIs(util::JSON::Type::Number, "Timestamp")) {
        // for old relay compat the router sends this back in millis, so convert back to seconds
        mRouterInfo.setTimestamp(doc.get<uint64_t>("Timestamp") / 1000);
      } else {
        Log("init timestamp not a number");
        return false;
      }
    } else {
      Log("response json missing member 'Timestamp'");
      return false;
    }

    return true;
  }

  template <typename T>
  bool Backend<T>::updateCycle(
   const volatile bool& loopHandle,
   const volatile bool& shouldCleanShutdown,
   util::ThroughputRecorder& recorder,
   core::SessionMap& sessions)
  {
    bool successfulRoutine = true;
    std::vector<uint8_t> update_response_memory;
    update_response_memory.resize(RESPONSE_MAX_BYTES);

    // update once every 10 seconds
    // if the update fails, try again, once per second for (MaxUpdateAttempts - 1) seconds
    // if there's still no successful update, exit the loop and return false, and skip the clean shutdown
    uint8_t updateAttempts = 0;

    util::Clock backendTimeout;
    while (loopHandle) {
      if (update(recorder, false)) {
        updateAttempts = 0;
        backendTimeout.reset();
      } else {
        auto timeSinceLastUpdate = backendTimeout.elapsed<util::Second>();
        if (++updateAttempts == MaxUpdateAttempts) {
          Log("could not update relay, max attempts reached, aborting program");
          successfulRoutine = false;
          break;
        } else if (timeSinceLastUpdate > 30) {
          Log("could not update relay for over 30 seconds, aborting program");
          successfulRoutine = false;
          break;
        }

        Log(
         "could not update relay, attempts: ", (unsigned int)updateAttempts, ", time since last update: ", timeSinceLastUpdate);
      }

      sessions.purge(mRouterInfo.currentTime());

      std::this_thread::sleep_for(1s);
    }

    if (shouldCleanShutdown) {
      unsigned int seconds = 0;
      while (seconds++ < 60 && !update(recorder, true)) {
        std::this_thread::sleep_for(1s);
      }

      if (seconds < 60) {
        std::this_thread::sleep_for(30s);
      }
    }

    return successfulRoutine;
  }

  template <typename T>
  auto Backend<T>::update(util::ThroughputRecorder& recorder, bool shutdown) -> bool
  {
    util::JSON doc;
    auto [ok, err] = buildUpdateRequest(doc, recorder, shutdown);
    if (!ok) {
      Log("error building update request: ", err);
      return false;
    }
    std::string request = doc.toString();
    std::string response;

    LogDebug("Sending new: ", doc.toPrettyString());

    if (!T::SendTo(mHostname, "/relay_update", request, response)) {
      Log("curl request failed in update");
      return false;
    }

    // early return if shutting down since the response won't be valid
    if (shutdown) {
      return true;
    }

    if (!doc.parse(response)) {
      Log("could not parse json response in update: ", doc.err(), "\nReponse: ", std::string(response.begin(), response.end()));
      return false;
    }

    LogDebug("Received new: ", doc.toPrettyString());

    if (doc.memberExists("version")) {
      if (doc.memberIs(util::JSON::Type::Number, "version")) {
        auto version = doc.get<uint32_t>("version");
        if (version != UpdateResponseVersion) {
          Log("error: bad relay version response version. expected ", UpdateResponseVersion, ", got ", version);
          return false;
        }
      } else {
        Log("warning, update version not number");
      }
    } else {
      Log("warning, version number missing in update response");
    }

    if (doc.memberExists("timestamp")) {
      if (doc.memberIs(util::JSON::Type::Number, "timestamp")) {
        mRouterInfo.setTimestamp(doc.get<int64_t>("timestamp"));
      } else {
        Log("init timestamp not a number");
        return false;
      }
    } else {
      Log("response json missing member 'timestamp'");
      return false;
    }

    size_t count = 0;
    std::array<Relay, MAX_RELAYS> incoming{};

    bool allValid = true;
    auto relays = doc.get<util::JSON>("ping_data");
    if (relays.isArray()) {
      // 'return' functions like 'continue' within the lambda
      relays.foreach([&allValid, &count, &incoming](rapidjson::Value& relayData) {
        if (!relayData.HasMember("relay_id")) {
          Log("ping data missing 'relay_id'");
          allValid = false;
          return;
        }

        auto idMember = std::move(relayData["relay_id"]);
        if (idMember.GetType() != rapidjson::Type::kNumberType) {
          Log("id from ping not number type");
          allValid = false;
          return;
        }

        auto id = idMember.GetUint64();

        if (!relayData.HasMember("relay_address")) {
          Log("ping data missing member 'relay_address' for relay id: ", id);
          allValid = false;
          return;
        }

        auto addrMember = std::move(relayData["relay_address"]);
        if (addrMember.GetType() != rapidjson::Type::kStringType) {
          Log("relay address is not a string in ping data for relay with id: ", id);
          allValid = false;
          return;
        }

        std::string address = addrMember.GetString();

        incoming[count].ID = id;
        if (!incoming[count].Addr.parse(address)) {
          Log("failed to parse address for relay '", id, "': ", address);
          allValid = false;
          return;
        }

        count++;
      });

      if (count > MAX_RELAYS) {
        Log("error: too many relays to ping. max is ", MAX_RELAYS, ", got ", count, '\n');
        return false;
      }
    } else if (relays.memberIs(util::JSON::Type::Null)) {
      LogDebug("no relays received from new backend, ping data is null");
    } else {
      Log("update ping data not array");
      // TODO how to handle
    }

    if (!allValid) {
      Log("some or all of the update ping data was invalid");
    }

    mRelayManager.update(count, incoming);

    return true;
  }

  template <typename T>
  auto Backend<T>::buildInitRequest(util::JSON& doc) -> std::tuple<bool, const char*>
  {
    std::string base64NonceStr;
    std::array<uint8_t, crypto_box_NONCEBYTES> nonce = {};
    crypto::CreateNonceBytes(nonce);
    std::vector<char> b64Nonce(nonce.size() * 2);

    auto len = encoding::base64::Encode(nonce, b64Nonce);
    if (len < nonce.size()) {
      return {false, "failed to encode base64 nonce for init"};
    }

    // greedy method but gets the job done, plus init is done once so who cares if it's a few nanos slower
    base64NonceStr = std::string(b64Nonce.begin(), b64Nonce.begin() + len);

    std::string base64TokenStr;
    // just has to be something the backend can decrypt
    std::array<uint8_t, RELAY_TOKEN_BYTES> token = {};
    crypto::RandomBytes(token, token.size());

    std::array<uint8_t, RELAY_TOKEN_BYTES + crypto_box_MACBYTES> encryptedToken = {};
    std::vector<char> b64EncryptedToken(encryptedToken.size() * 2);

    if (
     crypto_box_easy(
      encryptedToken.data(),
      token.data(),
      token.size(),
      nonce.data(),
      mKeychain.RouterPublicKey.data(),
      mKeychain.RelayPrivateKey.data()) != 0) {
      return {false, "failed to encrypt init token"};
    }

    len = encoding::base64::Encode(encryptedToken, b64EncryptedToken);
    if (len < encryptedToken.size()) {
      return {false, "failed to encode base64 token for init"};
    }

    base64TokenStr = std::string(b64EncryptedToken.begin(), b64EncryptedToken.begin() + len);

    doc.set(InitRequestMagic, "magic_request_protection");
    doc.set(InitRequestVersion, "version");
    doc.set(mAddressStr, "relay_address");
    doc.set(base64NonceStr, "nonce");
    doc.set(base64TokenStr, "encrypted_token");

    return {true, nullptr};
  }

  template <typename T>
  auto Backend<T>::buildUpdateRequest(util::JSON& doc, util::ThroughputRecorder& recorder, bool shutdown)
   -> std::tuple<bool, const char*>
  {
    // TODO once the other stats are finally added, pull out the json parts that are always the same, no sense rebuilding those
    // parts of the document
    doc.set(shutdown, "shutting_down");
    doc.set(UpdateRequestVersion, "version");
    doc.set(mAddressStr, "relay_address");
    doc.set(mBase64RelayPublicKey, "Metadata", "PublicKey");
    doc.set(RELAY_VERSION, "relay_version");

    // traffic stats
    {
      util::JSON trafficStats;

      util::ThroughputStatsCollection stats(std::move(recorder.get()));
      trafficStats.set(stats.Sent.ByteCount.load(), "BytesMeasurementTx");
      trafficStats.set(stats.Received.ByteCount.load(), "BytesMeasurementRx");
      trafficStats.set(mSessionMap.size(), "SessionCount");
      doc.set(trafficStats, "TrafficStats");
    }

    // ping stats
    {
      util::JSON pingStats;

      core::RelayStats stats;
      mRelayManager.getStats(stats);
      pingStats.setArray();

      for (unsigned int i = 0; i < stats.NumRelays; ++i) {
        util::JSON pingStat;
        pingStat.set(stats.IDs[i], "RelayId");
        pingStat.set(stats.RTT[i], "RTT");
        pingStat.set(stats.Jitter[i], "Jitter");
        pingStat.set(stats.PacketLoss[i], "PacketLoss");

        if (!pingStats.push(pingStat)) {
          return {false, "ping stats not array! can't update!"};
        }
      }

      doc.set(pingStats, "PingStats");
    }

    // sys
    {
      util::JSON sysStats;

      auto stats = os::GetUsage();

      sysStats.set(stats.CPU, "cpu_usage");
      sysStats.set(stats.Mem, "mem_usage");

      doc.set(sysStats, "sys_stats");
    }

    return {true, nullptr};
  }
}  // namespace core
#endif