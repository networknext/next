#pragma once

#include "backend_request.hpp"
#include "backend_response.hpp"
#include "backend_token.hpp"
#include "core/packet.hpp"
#include "core/relay_manager.hpp"
#include "crypto/keychain.hpp"
#include "net/address.hpp"
#include "os/platform.hpp"
#include "traffic_stats.hpp"
#include "util/channel.hpp"
#include "util/clock.hpp"
#include "util/env.hpp"
#include "util/json.hpp"
#include "core/session_map.hpp"

namespace legacy
{
  namespace v3
  {
    // Legacy support for the old v3 backend
    class Backend
    {
     public:
      Backend(
       volatile bool& shouldComm,
       util::Receiver<core::GenericPacket<>>& receiver,
       util::Env& env,
       const uint64_t relayID,
       os::Socket& socket,
       const util::Clock& relayClock,
       TrafficStats& stats,
       core::RelayManager<core::V3Relay>& manager,
       const size_t speed,
       std::atomic<ResponseState>& state,
       const crypto::Keychain& keychain,
       const core::SessionMap& sessions);
      ~Backend() = default;

      auto init() -> bool;
      auto config() -> bool;

      auto updateCycle(const volatile bool& handle) -> bool;

     private:
      volatile bool& mShouldCommunicate;
      util::Receiver<core::GenericPacket<>>& mReceiver;
      const util::Env& mEnv;
      os::Socket& mSocket;
      const util::Clock& mClock;
      TrafficStats& mStats;
      core::RelayManager<core::V3Relay>& mRelayManager;
      const size_t mSpeed;  // Relay nic speed in bits/second
      BackendToken mToken;
      uint64_t mInitTimestamp;  // in seconds
      uint64_t mInitReceived;   // in nanoseconds
      const uint64_t mRelayID;
      std::string mGroup;
      uint64_t mGroupID;
      std::string mPingKey;
      std::atomic<ResponseState>& mState;
      const crypto::Keychain& mKeychain;
      const core::SessionMap& mSessions;

      auto tryInit() -> bool;
      auto update(bool shuttingDown) -> bool;

      auto buildConfigJSON(util::JSON& doc) -> bool;
      auto buildUpdateJSON(util::JSON& doc, bool shuttingDown) -> bool;

      auto sendBinRecvJSON(
       BackendRequest& request, std::vector<uint8_t>& reqData, BackendResponse& response, util::JSON& respBuff)
       -> std::tuple<bool, std::string>;

      auto sendJSONRecvJSON(BackendRequest& request, util::JSON& reqData, BackendResponse& response, util::JSON& respBuff)
       -> std::tuple<bool, std::string>;

      auto signRequest(util::JSON& doc) -> bool;
      auto timestamp() -> uint64_t;

      auto buildCompleteResponse(std::vector<uint8_t>& completeBuffer, util::JSON& doc) -> std::tuple<bool, std::string>;
    };
  }  // namespace v3
}  // namespace legacy
