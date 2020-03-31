#ifndef CORE_BACKEND_HPP
#define CORE_BACKEND_HPP

#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "net/curl.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "testing/test.hpp"
#include "util/json.hpp"
#include "util/logger.hpp"
#include "util/throughput_logger.hpp"

namespace testing
{
  class _test_core_Backend_update_valid_;
}

namespace core
{
  const uint32_t InitRequestMagic = 0x9083708f;

  const uint32_t InitRequestVersion = 0;
  const uint32_t InitResponseVersion = 0;

  const uint32_t UpdateRequestVersion = 0;
  const uint32_t UpdateResponseVersion = 0;

  /*
   * A class that's responsible for backend related tasks
   * where T should be anything that defines a static SendTo function
   * with the same signature as net::CurlWrapper
   */
  template <typename T>
  class Backend
  {
   public:
    Backend(
     const std::string hostname,
     const std::string address,
     const crypto::Keychain& keychain,
     RouterInfo& routerInfo,
     RelayManager& relayManager,
     std::string base64RelayPublicKey,
     const core::SessionMap& sessions);
    ~Backend() = default;

    auto init() -> bool;

    void updateCycle(
     volatile bool& loopHandle, util::ThroughputLogger& logger, core::SessionMap& sessions, const util::Clock& relayClock);

   private:
    friend testing::_test_core_Backend_update_valid_;
    const std::string mHostname;
    const std::string mAddressStr;
    const crypto::Keychain& mKeychain;
    RouterInfo& mRouterInfo;
    RelayManager& mRelayManager;
    const std::string mBase64RelayPublicKey;
    const core::SessionMap& mSessionMap;

    auto update(uint64_t bytesReceived, bool shutdown) -> bool;
    auto buildInitRequest(util::JSON& doc) -> std::tuple<bool, const char*>;
    auto buildUpdateRequest(util::JSON& doc, uint64_t bytesReceived, bool shutdown) -> std::tuple<bool, const char*>;
  };

  template <typename T>
  Backend<T>::Backend(
   std::string hostname,
   std::string address,
   const crypto::Keychain& keychain,
   RouterInfo& routerInfo,
   RelayManager& relayManager,
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
      Log(err);
      return false;
    }
    std::string request = doc.toString();
    std::vector<char> response;

    LogDebug("init request: ", doc.toPrettyString());

    if (!T::SendTo(mHostname, "/relay_init", request, response)) {
      Log("curl request failed in init");
      return false;
    }

    if (!doc.parse(response)) {
      Log("could not parse json response in init: ", doc.err());
      return false;
    }

    LogDebug("init response: ", doc.toPrettyString());

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
        mRouterInfo.InitializeTimeInSeconds = doc.get<uint64_t>("Timestamp") / 1000;
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
  void Backend<T>::updateCycle(
   volatile bool& loopHandle, util::ThroughputLogger& logger, core::SessionMap& sessions, const util::Clock& relayClock)
  {
    std::vector<uint8_t> update_response_memory;
    update_response_memory.resize(RESPONSE_MAX_BYTES);
    while (loopHandle) {
      auto bytesReceived = logger.print();
      bool updated = false;

      for (int i = 0; i < 10; i++) {
        if (update(bytesReceived, false)) {
          updated = true;
          break;
        }
      }

      if (!updated) {
        std::cout << "error: could not update relay\n";
        break;
      }

      sessions.purge(relayClock.unixTime<util::Second>());
      std::this_thread::sleep_for(1s);
    }

    std::atomic<bool> shouldWait60 = true, shouldWait30 = true, waited60 = false;
    auto fut = std::async([&shouldWait60, &waited60] {
      for (uint seconds = 0; seconds < 60; seconds++) {
        std::this_thread::sleep_for(1s);
        if (!shouldWait60) {
          return;
        }
      }

      waited60 = true;
    });

    // keep living for another 30 seconds
    // no more updates allows the backend to remove
    // this relay from the route decisions
    while (!update(0, true) && !waited60) {
      std::this_thread::sleep_for(1s);
    }
    shouldWait60 = false;
    if (!waited60) {
      std::this_thread::sleep_for(30s);
    }

    fut.wait();
  }

  template <typename T>
  auto Backend<T>::update(uint64_t bytesReceived, bool shutdown) -> bool
  {
    util::JSON doc;
    auto [ok, err] = buildUpdateRequest(doc, bytesReceived, shutdown);
    if (!ok) {
      Log(err);
      return false;
    }
    std::string request = doc.toString();
    std::vector<char> response;

    LogDebug("update request: ", doc.toPrettyString());

    if (!T::SendTo(mHostname, "/relay_update", request, response)) {
      Log("curl request failed in update");
      return false;
    }

    if (!doc.parse(response)) {
      Log("could not parse json response in update: ", doc.err());
      return false;
    }

    LogDebug("update response: ", doc.toPrettyString());

    if (doc.memberExists("version")) {
      if (doc.memberIs(util::JSON::Type::Number, "version")) {
        auto version = doc.get<uint32_t>("version");
        if (version != UpdateResponseVersion) {
          Log("error: bad relay version response version. expected ", UpdateResponseVersion, ", got ", version);
          return false;
        }
      } else {
        Log("warning: update version not number");
      }
    } else {
      Log("warning, version number missing in update response");
    }

    bool allValid = true;
    auto relays = doc.get<util::JSON>("ping_data");
    if (relays.isArray()) {
      size_t count = 0;
      std::array<uint64_t, MAX_RELAYS> relayIDs = {};
      std::array<net::Address, MAX_RELAYS> relayAddresses;
      relays.foreach([&allValid, &count, &relayIDs, &relayAddresses](rapidjson::Value& relayData) {
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

        relayIDs[count] = id;
        if (!relayAddresses[count].parse(address)) {
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

      mRelayManager.update(count, relayIDs, relayAddresses);
    } else if (relays.memberIs(util::JSON::Type::Null)) {
      Log("no relays received from backend, ping data is null");
    } else {
      Log("update ping data not array");
      // TODO how to handle
    }

    if (!allValid) {
      Log("some or all of the update ping data was invalid");
    }

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
  auto Backend<T>::buildUpdateRequest(util::JSON& doc, uint64_t bytesReceived, bool shutdown) -> std::tuple<bool, const char*>
  {
    // TODO once the other stats are finally added, pull out the json parts that are always the same, no sense rebuilding those
    // parts of the document
    doc.set(shutdown, "shutting_down");
    doc.set(UpdateRequestVersion, "version");
    doc.set(mAddressStr, "relay_address");
    doc.set(mBase64RelayPublicKey, "Metadata", "PublicKey");

    util::JSON trafficStats;
    trafficStats.set(bytesReceived, "BytesMeasurementRx");
    trafficStats.set(mSessionMap.size(), "SessionCount");

    doc.set(trafficStats, "TrafficStats");

    core::RelayStats stats;
    mRelayManager.getStats(stats);
    util::JSON pingStats;
    pingStats.setArray();

    // pushing behaves really weird when the pushed value goes out of scope, must be declared outside of for loop
    util::JSON pingStat;
    for (unsigned int i = 0; i < stats.NumRelays; ++i) {
      pingStat.set(stats.IDs[i], "RelayId");
      pingStat.set(stats.RTT[i], "RTT");
      pingStat.set(stats.Jitter[i], "Jitter");
      pingStat.set(stats.PacketLoss[i], "PacketLoss");

      if (!pingStats.push(pingStat)) {
        return {false, "ping stats not array! can't update!"};
      }
    }

    // performs a deep copy, so it's ok for things to go out of scope after this, regular move seems to be weird due to the
    // allocator concept
    doc.set(pingStats, "PingStats");

    return {true, nullptr};
  }
}  // namespace core
#endif