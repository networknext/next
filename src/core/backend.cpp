#include "includes.h"
#include "backend.hpp"

#include "crypto/bytes.hpp"
#include "crypto/keychain.hpp"
#include "encoding/base64.hpp"
#include "encoding/read.hpp"
#include "encoding/write.hpp"
#include "net/http.hpp"
#include "os/system.hpp"

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

    if (!encoding::write_uint32(v, index, Magic)) {
      LOG(ERROR, "could not write init request magic");
      return false;
    }

    if (!encoding::write_uint32(v, index, Version)) {
      LOG(ERROR, "could not write init request version");
      return false;
    }

    if (!encoding::write_bytes(v, index, Nonce, Nonce.size())) {
      LOG(ERROR, "could not write init request nonce bytes");
      return false;
    }

    if (!encoding::write_string(v, index, Address)) {
      LOG(ERROR, "could not write init request address");
      return false;
    }

    if (!encoding::write_bytes(v, index, EncryptedToken, EncryptedToken.size())) {
      LOG(ERROR, "could not write init request token");
      return false;
    }

    if (!encoding::write_string(v, index, RelayVersion)) {
      LOG(ERROR, "could not write init request relay version");
      return false;
    }

    return true;
  }

  auto InitRequest::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    if (!encoding::read_uint32(v, index, this->Magic)) {
      return false;
    }
    if (!encoding::read_uint32(v, index, this->Version)) {
      return false;
    }
    if (!encoding::read_bytes(v, index, this->Nonce, this->Nonce.size())) {
      return false;
    }
    if (!encoding::read_string(v, index, this->Address)) {
      return false;
    }
    if (!encoding::read_bytes(v, index, this->EncryptedToken, this->EncryptedToken.size())) {
      return false;
    }
    if (!encoding::read_string(v, index, this->RelayVersion)) {
      return false;
    }
    return true;
  }

  // only used in tests
  auto InitResponse::into(std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    if (!encoding::write_uint32(v, index, Version)) {
      LOG(TRACE, "unable to write version");
      return false;
    }

    if (!encoding::write_uint64(v, index, Timestamp)) {
      LOG(TRACE, "unable to write timestamp");
      return false;
    }

    if (!encoding::write_bytes(v, index, PublicKey, PublicKey.size())) {
      LOG(TRACE, "unable to write public key");
      return false;
    }

    return true;
  }

  auto InitResponse::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    if (!encoding::read_uint32(v, index, this->Version)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->Timestamp)) {
      return false;
    }
    if (!encoding::read_bytes(v, index, PublicKey, PublicKey.size())) {
      return false;
    }
    return true;
  }

  auto UpdateRequest::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;
    if (!encoding::read_uint32(v, index, this->Version)) {
      return false;
    }
    if (!encoding::read_string(v, index, this->Address)) {
      return false;
    }
    if (!encoding::read_bytes(v, index, PublicKey, PublicKey.size())) {
      return false;
    }
    if (!encoding::read_uint32(v, index, this->PingStats.NumRelays)) {
      return false;
    }

    for (size_t i = 0; i < PingStats.NumRelays; i++) {
      if (!encoding::read_uint64(v, index, this->PingStats.IDs[i])) {
        return false;
      }
      if (!encoding::read_bytes(
           v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&PingStats.RTT[i]), sizeof(float), sizeof(float))) {
        return false;
      }
      if (!encoding::read_bytes(
           v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&PingStats.Jitter[i]), sizeof(float), sizeof(float))) {
        return false;
      }
      if (!encoding::read_bytes(
           v.data(), v.size(), index, reinterpret_cast<uint8_t*>(&PingStats.PacketLoss[i]), sizeof(float), sizeof(float))) {
        return false;
      }
    }
    if (!encoding::read_uint64(v, index, this->SessionCount)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->OutboundPingTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->RouteRequestRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->RouteRequestTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->RouteResponseRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->RouteResponseTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->ClientToServerRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->ClientToServerTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->ServerToClientRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->ServerToClientTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->InboundPingRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->InboundPingTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->PongRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->SessionPingRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->SessionPingTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->SessionPongRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->SessionPongTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->ContinueRequestRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->ContinueRequestTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->ContinueResponseRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->ContinueResponseTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->NearPingRx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->NearPingTx)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, this->UnknownRx)) {
      return false;
    }
    uint8_t shutdown_flag;
    if (!encoding::read_uint8(v, index, shutdown_flag)) {
      return false;
    }
    this->ShuttingDown = static_cast<bool>(shutdown_flag);

    if (!encoding::read_double(v, index, CPUUsage)) {
      return false;
    }
    if (!encoding::read_double(v, index, MemUsage)) {
      return false;
    }
    if (!encoding::read_string(v, index, RelayVersion)) {
      return false;
    }

    return true;
  }

  auto UpdateResponse::size() -> size_t
  {
    size_t size = 4 + 8 + 4 + NumRelays * (8 + 4);

    for (size_t i = 0; i < NumRelays; i++) {
      // only used in tests, so being lazy here;
      const auto& relay = Relays[i];
      size += relay.address.toString().length();
    }

    return size;
  }

  // only used in tests
  auto UpdateResponse::into(std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;

    if (!encoding::write_uint32(v, index, Version)) {
      LOG(TRACE, "could not write version");
      return false;
    }

    if (!encoding::write_uint64(v, index, Timestamp)) {
      LOG(TRACE, "could not write timestamp");
      return false;
    }

    if (!encoding::write_uint32(v, index, NumRelays)) {
      LOG(TRACE, "could not write num relays");
      return false;
    }

    for (size_t i = 0; i < NumRelays; i++) {
      const auto& relay = Relays[i];

      if (!encoding::write_uint64(v, index, relay.id)) {
        LOG(TRACE, "could not write relay id");
        return false;
      }

      if (!encoding::write_string(v, index, relay.address.toString())) {
        LOG(TRACE, "could not write relay address");
        return false;
      }
    }

    return true;
  }

  auto UpdateResponse::from(const std::vector<uint8_t>& v) -> bool
  {
    size_t index = 0;

    if (!encoding::read_uint32(v, index, Version)) {
      return false;
    }
    if (!encoding::read_uint64(v, index, Timestamp)) {
      return false;
    }
    if (!encoding::read_uint32(v, index, NumRelays)) {
      return false;
    }
    for (size_t i = 0; i < NumRelays; i++) {
      auto& relay = Relays[i];
      if (!encoding::read_uint64(v, index, relay.id)) {
        return false;
      }
      std::string addr;
      if (!encoding::read_string(v, index, addr)) {
        return false;
      }
      if (!relay.address.parse(addr)) {
        LOG(ERROR, "unable to parse relay address: ", addr);
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
   RelayManager& relayManager,
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
        mKeychain.backend_public_key.data(),
        mKeychain.relay_private_key.data()) != 0) {
        LOG(TRACE, "failed to encrypt init token");
        return false;
      }

      requestData.resize(request.size());
      if (!request.into(requestData)) {
        return false;
      }
    }

    // send request

    if (!mRequester.sendRequest(mHostname, "/relay_init", requestData, responseData)) {
      LOG(ERROR, "init request failed");
      return false;
    }

    // deserialize response
    {
      InitResponse response;
      if (!response.from(responseData)) {
        return false;
      }

      if (response.Version != InitResponseVersion) {
        LOG(ERROR, "error: bad relay init response version. expected ", InitResponseVersion, ", got ", response.Version);
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
          LOG(ERROR, "could not update relay, max attempts reached, aborting program");
          successfulRoutine = false;
          break;
        } else if (timeSinceLastUpdate > 30) {
          LOG(ERROR, "could not update relay for over 30 seconds, aborting program");
          successfulRoutine = false;
          break;
        }

        LOG(
         INFO,
         "could not update relay, attempts: ",
         static_cast<unsigned int>(updateAttempts),
         ", time since last update: ",
         timeSinceLastUpdate);
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
      mRelayManager.get_stats(stats);

      const size_t requestSize = 4 +                     // request version
                                 4 +                     // address length
                                 mAddressStr.length() +  // address
                                 crypto::KEY_SIZE +       // public key
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

      encoding::write_uint32(req, index, UpdateRequestVersion);
      encoding::write_string(req, index, mAddressStr);
      encoding::write_bytes(req, index, mKeychain.relay_public_key, mKeychain.relay_public_key.size());
      encoding::write_uint32(req, index, stats.NumRelays);

      for (unsigned int i = 0; i < stats.NumRelays; ++i) {
        encoding::write_uint64(req, index, stats.IDs[i]);
        encoding::write_bytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.RTT[i]), sizeof(float));
        encoding::write_bytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.Jitter[i]), sizeof(float));
        encoding::write_bytes(req.data(), req.size(), index, reinterpret_cast<uint8_t*>(&stats.PacketLoss[i]), sizeof(float));
      }

      encoding::write_uint64(req, index, mSessionMap.size());

      util::ThroughputRecorder trafficStats(std::move(recorder));

      encoding::write_uint64(req, index, trafficStats.outbound_ping_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.route_request_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.route_request_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.route_response_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.route_response_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.client_to_server_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.client_to_server_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.server_to_client_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.server_to_client_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.inbound_ping_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.inbound_ping_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.pong_rx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.session_ping_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.session_ping_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.session_pong_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.session_pong_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.continue_request_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.continue_request_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.continue_response_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.continue_response_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.near_ping_rx.num_bytes.load());
      encoding::write_uint64(req, index, trafficStats.near_ping_tx.num_bytes.load());

      encoding::write_uint64(req, index, trafficStats.unknown_rx.num_bytes.load());

      encoding::write_uint8(req, index, shutdown);

      auto sysStats = os::GetUsage();
      encoding::write_double(req, index, sysStats.CPU);
      encoding::write_double(req, index, sysStats.Mem);
      encoding::write_string(req, index, RELAY_VERSION);
    }

    if (!mRequester.sendRequest(mHostname, "/relay_update", req, res)) {
      LOG(ERROR, "update request failed");
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
        LOG(ERROR, "could not deserialize update response");
        return false;
      }

      if (response.Version != UpdateResponseVersion) {
        LOG(ERROR, "bad relay version response version. expected ", UpdateResponseVersion, ", got ", response.Version);
        return false;
      }

      mRouterInfo.setTimestamp(response.Timestamp);

      if (response.NumRelays > MAX_RELAYS) {
        LOG(ERROR, "too many relays to ping. max is ", MAX_RELAYS, ", got ", response.NumRelays, '\n');
        return false;
      }

      mRelayManager.update(response.NumRelays, response.Relays);
    }

    return true;
  }
}  // namespace core