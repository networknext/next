#include "includes.h"
#include "backend.hpp"

#include "encoding/base64.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"

namespace
{
  const uint32_t InitResponseVersion = 0;

  const uint32_t UpdateRequestVersion = 1;
  const uint32_t UpdateResponseVersion = 0;
}  // namespace

namespace core
{
  using namespace std::chrono_literals;

  auto InitRequest::size() -> size_t
  {
    return 4 + 4 + Nonce.size() + 4 + Address.length() + EncryptedToken.size() + 4 + RelayVersion.length();
  }

  auto InitRequest::into(std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;

    if (!encoding::WriteUint32(v, index, Magic)) {
      LOG("could not write init request magic");
      return false;
    }

    if (!encoding::WriteUint32(v, index, Version)) {
      LOG("could not write init request version");
      return false;
    }

    if (!encoding::WriteBytes(v, index, Nonce, Nonce.size())) {
      LOG("could not write init request nonce bytes");
      return false;
    }

    if (!encoding::WriteString(v, index, Address)) {
      LOG("could not write init request address");
      return false;
    }

    if (!encoding::WriteBytes(v, index, EncryptedToken, EncryptedToken.size())) {
      LOG("could not write init request token");
      return false;
    }

    if (!encoding::WriteString(v, index, RelayVersion)) {
      LOG("could not write init request relay version");
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

  // only used in tests
  auto InitResponse::into(std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    if (!encoding::WriteUint32(v, index, Version)) {
      LogTest("unable to write version");
      return false;
    }

    if (!encoding::WriteUint64(v, index, Timestamp)) {
      LogTest("unable to write timestamp");
      return false;
    }

    if (!encoding::WriteBytes(v, index, PublicKey, PublicKey.size())) {
      LogTest("unable to write public key");
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

  auto UpdateRequest::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    Version = encoding::ReadUint32(v, index);
    Address = encoding::ReadString(v, index);
    encoding::ReadBytes(v, index, PublicKey, PublicKey.size());

    PingStats.NumRelays = encoding::ReadUint32(v, index);

    for (size_t i = 0; i < PingStats.NumRelays; i++) {
      PingStats.IDs[i] = encoding::ReadUint64(v, index);
      encoding::ReadBytes(
       v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&PingStats.RTT[i]), sizeof(float), sizeof(float));
      encoding::ReadBytes(
       v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&PingStats.Jitter[i]), sizeof(float), sizeof(float));
      encoding::ReadBytes(
       v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&PingStats.PacketLoss[i]), sizeof(float), sizeof(float));
    }

    SessionCount = encoding::ReadUint64(v, index);
    OutboundPingTx = encoding::ReadUint64(v, index);
    RouteRequestRx = encoding::ReadUint64(v, index);
    RouteRequestTx = encoding::ReadUint64(v, index);
    RouteResponseRx = encoding::ReadUint64(v, index);
    RouteResponseTx = encoding::ReadUint64(v, index);
    ClientToServerRx = encoding::ReadUint64(v, index);
    ClientToServerTx = encoding::ReadUint64(v, index);
    ServerToClientRx = encoding::ReadUint64(v, index);
    ServerToClientTx = encoding::ReadUint64(v, index);
    InboundPingRx = encoding::ReadUint64(v, index);
    InboundPingTx = encoding::ReadUint64(v, index);
    PongRx = encoding::ReadUint64(v, index);
    SessionPingRx = encoding::ReadUint64(v, index);
    SessionPingTx = encoding::ReadUint64(v, index);
    SessionPongRx = encoding::ReadUint64(v, index);
    SessionPongTx = encoding::ReadUint64(v, index);
    ContinueRequestRx = encoding::ReadUint64(v, index);
    ContinueRequestTx = encoding::ReadUint64(v, index);
    ContinueResponseRx = encoding::ReadUint64(v, index);
    ContinueResponseTx = encoding::ReadUint64(v, index);
    NearPingRx = encoding::ReadUint64(v, index);
    NearPingTx = encoding::ReadUint64(v, index);
    UnknownRx = encoding::ReadUint64(v, index);
    ShuttingDown = static_cast<bool>(encoding::ReadUint8(v, index));

    CPUUsage = encoding::ReadDouble(v, index);
    MemUsage = encoding::ReadDouble(v, index);
    RelayVersion = encoding::ReadString(v, index);

    return true;
  }

  auto UpdateResponse::size() -> size_t
  {
    size_t size = 4 + 8 + 4 + NumRelays * (8 + 4);

    for (size_t i = 0; i < NumRelays; i++) {
      // only used in tests, so being lazy here;
      const auto& relay = Relays[i];
      size += relay.Addr.toString().length();
    }

    return size;
  }

  // only used in tests
  auto UpdateResponse::into(std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;

    if (!encoding::WriteUint32(v, index, Version)) {
      LogTest("could not write version");
      return false;
    }

    if (!encoding::WriteUint64(v, index, Timestamp)) {
      LogTest("could not write timestamp");
      return false;
    }

    if (!encoding::WriteUint32(v, index, NumRelays)) {
      LogTest("could not write num relays");
      return false;
    }

    for (size_t i = 0; i < NumRelays; i++) {
      const auto& relay = Relays[i];

      if (!encoding::WriteUint64(v, index, relay.ID)) {
        LogTest("could not write relay id");
        return false;
      }

      if (!encoding::WriteString(v, index, relay.Addr.toString())) {
        LogTest("could not write relay address");
        return false;
      }
    }

    return true;
  }

  auto UpdateResponse::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;

    Version = encoding::ReadUint32(v, index);
    Timestamp = encoding::ReadUint64(v, index);
    NumRelays = encoding::ReadUint32(v, index);
    for (size_t i = 0; i < NumRelays; i++) {
      auto& relay = Relays[i];
      relay.ID = encoding::ReadUint64(v, index);
      std::string addr = encoding::ReadString(v, index);
      if (!relay.Addr.parse(addr)) {
        LOG("unable to parse relay address: ", addr);
        return false;
      }
    }

    return true;
  }

  Backend::Backend(
   std::string hostname,
   std::string address,
   const crypto::Keychain& keychain,
   RouterInfo& routerInfo,
   RelayManager<Relay>& relayManager,
   std::string base64RelayPublicKey,
   const core::SessionMap& sessions,
   net::IHttpClient& client)
   : mHostname(hostname),
     mAddressStr(address),
     mKeychain(keychain),
     mRouterInfo(routerInfo),
     mRelayManager(relayManager),
     mBase64RelayPublicKey(base64RelayPublicKey),
     mSessionMap(sessions),
     mRequester(client)
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
        LOG("failed to encrypt init token");
        return false;
      }

      requestData.resize(request.size());
      if (!request.into(requestData)) {
        return false;
      }
    }

    // send request

    if (!mRequester.sendRequest(mHostname, "/relay_init", requestData, responseData)) {
      LOG("init request failed");
      return false;
    }

    // deserialize response
    {
      InitResponse response;
      if (!response.from(responseData)) {
        return false;
      }

      if (response.Version != InitResponseVersion) {
        LOG("error: bad relay init response version. expected ", InitResponseVersion, ", got ", response.Version);
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
          LOG("could not update relay, max attempts reached, aborting program");
          successfulRoutine = false;
          break;
        } else if (timeSinceLastUpdate > 30) {
          LOG("could not update relay for over 30 seconds, aborting program");
          successfulRoutine = false;
          break;
        }

        LOG(
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

      const size_t requestSize = 4 +                     // request version
                                 4 +                     // address length
                                 mAddressStr.length() +  // address
                                 crypto::KeySize +       // public key
                                 4 +                     // number of relay ping stats
                                 stats.NumRelays * 20 +  // relay ping stats
                                 8 +                     // session count
                                 8 +                     // outbound ping tx
                                 8 +                     // route request rx
                                 8 +                     // route request tx
                                 8 +                     // route response rx
                                 8 +                     // route response tx
                                 8 +                     // client to server rx
                                 8 +                     // client to server tx
                                 8 +                     // server to client rx
                                 8 +                     // server to client tx
                                 8 +                     // inbound ping rx
                                 8 +                     // inbound ping tx
                                 8 +                     // pong rx
                                 8 +                     // session ping rx
                                 8 +                     // session ping tx
                                 8 +                     // session pong rx
                                 8 +                     // session pong tx
                                 8 +                     // continue request rx
                                 8 +                     // continue request tx
                                 8 +                     // continue response rx
                                 8 +                     // continue response tx
                                 8 +                     // near ping rx
                                 8 +                     // near ping tx
                                 8 +                     // unknown Rx
                                 1 +                     // shut down flag
                                 8 +                     // cpu usage
                                 8 +                     // memory usage
                                 4 +                     // relay version length
                                 strlen(RELAY_VERSION);  // relay version string
      req.resize(requestSize);

      size_t index = 0;

      encoding::WriteUint32(req, index, UpdateRequestVersion);
      encoding::WriteString(req, index, mAddressStr);
      encoding::WriteBytes(req, index, mKeychain.RelayPublicKey, mKeychain.RelayPublicKey.size());
      encoding::WriteUint32(req, index, stats.NumRelays);

      for (unsigned int i = 0; i < stats.NumRelays; ++i) {
        encoding::WriteUint64(req, index, stats.IDs[i]);
        encoding::WriteBytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.RTT[i]), sizeof(float));
        encoding::WriteBytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.Jitter[i]), sizeof(float));
        encoding::WriteBytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.PacketLoss[i]), sizeof(float));
      }

      encoding::WriteUint64(req, index, mSessionMap.size());

      util::ThroughputRecorder trafficStats(std::move(recorder));

      encoding::WriteUint64(req, index, trafficStats.OutboundPingTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.RouteRequestRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.RouteRequestTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.RouteResponseRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.RouteResponseTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.ClientToServerRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.ClientToServerTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.ServerToClientRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.ServerToClientTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.InboundPingRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.InboundPingTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.PongRx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.SessionPingRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.SessionPingTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.SessionPongRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.SessionPongTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.ContinueRequestRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.ContinueRequestTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.ContinueResponseRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.ContinueResponseTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.NearPingRx.ByteCount.load());
      encoding::WriteUint64(req, index, trafficStats.NearPingTx.ByteCount.load());

      encoding::WriteUint64(req, index, trafficStats.UnknownRx.ByteCount.load());

      encoding::WriteUint8(req, index, shutdown);

      auto sysStats = os::GetUsage();
      encoding::WriteDouble(req, index, sysStats.CPU);
      encoding::WriteDouble(req, index, sysStats.Mem);
      encoding::WriteString(req, index, RELAY_VERSION);
    }

    if (!mRequester.sendRequest(mHostname, "/relay_update", req, res)) {
      LOG("update request failed");
      return false;
    }

    // early return if shutting down since the response won't be valid
    if (shutdown) {
      return true;
    }

    // parse response
    {
      UpdateResponse response;
      if (!response.from(res)) {
        LOG("could not deserialize update response");
        return false;
      }

      if (response.Version != UpdateResponseVersion) {
        LOG("error: bad relay version response version. expected ", UpdateResponseVersion, ", got ", response.Version);
        return false;
      }

      mRouterInfo.setTimestamp(response.Timestamp);

      if (response.NumRelays > MAX_RELAYS) {
        LOG("error: too many relays to ping. max is ", MAX_RELAYS, ", got ", response.NumRelays, '\n');
        return false;
      }

      mRelayManager.update(response.NumRelays, response.Relays);
    }

    return true;
  }
}  // namespace core