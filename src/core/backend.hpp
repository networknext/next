#ifndef CORE_BACKEND_HPP
#define CORE_BACKEND_HPP

#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "legacy/v3/traffic_stats.hpp"
#include "net/http.hpp"
#include "os/platform.hpp"
#include "relay_manager.hpp"
#include "router_info.hpp"
#include "session_map.hpp"
#include "testing/test.hpp"
#include "util/json.hpp"
#include "util/logger.hpp"
#include "util/throughput_recorder.hpp"

// forward declare test names to allow private functions to be visible them
namespace testing
{
  class _test_core_Backend_update_valid_;
  class _test_core_Backend_update_shutting_down_true_;
}  // namespace testing

namespace core
{
  const uint32_t InitRequestMagic = 0x9083708f;

  const uint32_t InitRequestVersion = 1;
  const uint32_t InitResponseVersion = 0;

  const uint32_t UpdateRequestVersion = 1;
  const uint32_t UpdateResponseVersion = 0;

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

  auto InitRequest::size() -> size_t
  {
    return 4 + 4 + Nonce.size() + 4 + Address.length() + EncryptedToken.size() + 4 + RelayVersion.length();
  }

  auto InitRequest::into(std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;

    if (!encoding::WriteUint32(v, index, Magic)) {
      Log("could not write init request magic");
      return false;
    }

    if (!encoding::WriteUint32(v, index, Version)) {
      Log("could not write init request version");
      return false;
    }

    if (!encoding::WriteBytes(v, index, Nonce, Nonce.size())) {
      Log("could not write init request nonce bytes");
      return false;
    }

    if (!encoding::WriteString(v, index, Address)) {
      Log("could not write init request address");
      return false;
    }

    if (!encoding::WriteBytes(v, index, EncryptedToken, EncryptedToken.size())) {
      Log("could not write init request token");
      return false;
    }

    if (!encoding::WriteString(v, index, RelayVersion)) {
      Log("could not write init request relay version");
      return false;
    }

    return true;
  }

  auto InitRequest::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    Magic = encoding::ReadUint32(v, index);
    Version = encoding::ReadUint32(v, index);
    encoding::ReadBytes(v, index, Nonce, Nonce.size());
    Address = encoding::ReadString(v, index);
    encoding::ReadBytes(v, index, EncryptedToken, EncryptedToken.size());
    RelayVersion = encoding::ReadString(v, index);
    return true;
  }

  struct InitResponse
  {
    static const size_t ByteSize = 4 + 8 + crypto::KeySize;
    uint32_t Version;
    uint64_t Timestamp;
    crypto::GenericKey PublicKey;

    auto into(std::vector<uint8_t>& v) -> bool;
    auto from(const std::vector<uint8_t>& v) -> bool;
  };

  auto InitResponse::into(std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    if (!encoding::WriteUint32(v, index, Version)) {
      return false;
    }

    if (!encoding::WriteUint64(v, index, Timestamp)) {
      return false;
    }

    if (!encoding::WriteBytes(v, index, PublicKey, PublicKey.size())) {
      return false;
    }

    return true;
  }

  auto InitResponse::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    Version = encoding::ReadUint32(v, index);
    Timestamp = encoding::ReadUint64(v, index);
    encoding::ReadBytes(v, index, PublicKey, PublicKey.size());
    return true;
  }

  struct UpdateRequest
  {
    UpdateRequest(const crypto::GenericKey& publicKey, const SessionMap& sessions);
    uint32_t Version;
    std::string Address;
    const crypto::GenericKey& PublicKey;
    RelayStats Stats;
    uint64_t SessionCount;
    uint64_t Tx;
    uint64_t Rx;
    float CPUUsage;
    float MemUsage;
  };

  UpdateRequest::UpdateRequest(const crypto::GenericKey& publicKey, const SessionMap& sessions)
   : PublicKey(publicKey), Sessions(sessions)
  {}

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
     RelayManager<Relay>& relayManager,
     std::string base64RelayPublicKey,
     const core::SessionMap& sessions,
     legacy::v3::TrafficStats& stats,
     net::IHttpRequester& requester);
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
    legacy::v3::TrafficStats& mStats;
    net::IHttpRequester& mRequester;

    auto update(util::ThroughputRecorder& recorder, bool shutdown) -> bool;
  };

  Backend::Backend(
   std::string hostname,
   std::string address,
   const crypto::Keychain& keychain,
   RouterInfo& routerInfo,
   RelayManager<Relay>& relayManager,
   std::string base64RelayPublicKey,
   const core::SessionMap& sessions,
   legacy::v3::TrafficStats& stats,
   net::IHttpRequester& requester)
   : mHostname(hostname),
     mAddressStr(address),
     mKeychain(keychain),
     mRouterInfo(routerInfo),
     mRelayManager(relayManager),
     mBase64RelayPublicKey(base64RelayPublicKey),
     mSessionMap(sessions),
     mStats(stats),
     mRequester(requester)
  {}

  auto Backend::init() -> bool
  {
    std::vector<uint8_t> requestData, responseData;

    // serialize request
    {
      InitRequest request;
      request.Address = mAddressStr;

      crypto::CreateNonceBytes(request.Nonce);

      // just has to be something the backend can decrypt
      std::array<uint8_t, RELAY_TOKEN_BYTES> token = {};
      crypto::RandomBytes(token, token.size());

      if (
       crypto_box_easy(
        request.EncryptedToken.data(),
        token.data(),
        token.size(),
        request.Nonce.data(),
        mKeychain.RouterPublicKey.data(),
        mKeychain.RelayPrivateKey.data()) != 0) {
        Log("failed to encrypt init token");
        return false;
      }

      requestData.resize(request.size());
      if (!request.into(requestData)) {
        return false;
      }
    }

    // send request

    if (!mRequester.sendRequest(mHostname, "/relay_init", requestData, responseData)) {
      Log("curl request failed in init");
      return false;
    }

    // deserialize response
    {
      InitResponse response;
      if (!response.from(responseData)) {
        return false;
      }

      if (response.Version != InitResponseVersion) {
        Log("error: bad relay init response version. expected ", InitResponseVersion, ", got ", response.Version);
        return false;
      }

      // for old relay compat the router sends this back in millis, so convert back to seconds
      mRouterInfo.setTimestamp(response.Timestamp / 1000);
    }

    return true;
  }

  bool Backend::updateCycle(
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

  auto Backend::update(util::ThroughputRecorder& recorder, bool shutdown) -> bool
  {
    std::vector<uint8_t> req, res;

    // serialize request
    {
      core::RelayStats stats;
      mRelayManager.getStats(stats);

      // | version | address length | address | public key | num stats | ping stats | session count | bytes sent | bytes
      // received | shutting down | cpu usage | memory usage |
      const size_t requestSize =
       4 + 4 + mAddressStr.length() + crypto::KeySize + 4 + 20 * stats.NumRelays + 8 + 8 + 8 + 1 + 8 + 8;
      req.resize(requestSize);

      size_t index = 0;

      encoding::WriteUint32(req, index, UpdateRequestVersion);
      encoding::WriteString(req, index, mAddressStr);
      encoding::WriteBytes(req, index, mKeychain.RelayPublicKey, mKeychain.RelayPublicKey.size());
      encoding::WriteUint32(req, index, stats.NumRelays);

      for (unsigned int i = 0; i < stats.NumRelays; ++i) {
        encoding::WriteUint64(req, index, stats.IDs[i]);
        encoding::WriteBytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.RTT[i]), sizeof(uint32_t));
        encoding::WriteBytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.Jitter[i]), sizeof(uint32_t));
        encoding::WriteBytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.PacketLoss[i]), sizeof(uint32_t));
      }

      encoding::WriteUint64(req, index, mSessionMap.size());

      util::ThroughputStatsCollection trafficStats(std::move(recorder.get()));

      encoding::WriteUint64(req, index, trafficStats.Sent.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.Received.ByteCount.load());

      auto sysStats = os::GetUsage();
      encoding::WriteBytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&sysStats.CPU), sizeof(uint64_t));
      encoding::WriteBytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&sysStats.Mem), sizeof(uint64_t));
    }

    if (!mRequester.sendRequest(mHostname, "/relay_update", req, res)) {
      Log("curl request failed in update");
      return false;
    }

    // early return if shutting down since the response won't be valid
    if (shutdown) {
      return true;
    }

    // parse response
    {
      size_t index = 0;

      uint32_t version = encoding::ReadUint32(res, index);
      if (version != UpdateResponseVersion) {
        Log("error: bad relay version response version. expected ", UpdateResponseVersion, ", got ", version);
        return false;
      }

      uint64_t timestamp = encoding::ReadUint64(res, index);
      mRouterInfo.setTimestamp(timestamp);

      size_t numRelays = encoding::ReadUint32(res, index);
      if (numRelays > MAX_RELAYS) {
        Log("error: too many relays to ping. max is ", MAX_RELAYS, ", got ", numRelays, '\n');
        return false;
      }

      std::array<Relay, MAX_RELAYS> incoming{};

      bool allValid = true;
      for (size_t i = 0; i < numRelays; i++) {
        auto& relay = incoming[i];

        uint64_t id = encoding::ReadUint64(res, index);
        std::string addr = encoding::ReadString(res, index);

        relay.ID = id;
        if (!relay.Addr.parse(addr)) {
          Log("failed to parse address for relay '", id, "': ", addr);
          allValid = false;
        }
      }

      if (!allValid) {
        Log("some or all of the update ping data was invalid");
      }

      mRelayManager.update(numRelays, incoming);
    }

    return true;
  }
}  // namespace core
#endif