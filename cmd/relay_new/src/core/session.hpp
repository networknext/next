#ifndef CORE_SESSION_HPP
#define CORE_SESSION_HPP

#include "expireable.hpp"
#include "net/address.hpp"
#include "replay_protection.hpp"

namespace core
{
  class Session: public Expireable
  {
   public:
    Session(const util::Clock& relayClock, const core::RouterInfo& routerInfo);
    virtual ~Session() override = default;

    uint64_t SessionID;
    uint8_t SessionVersion;
    uint64_t ClientToServerSeq;
    uint64_t ServerToClientSeq;
    int KbpsUp;
    int KbpsDown;
    net::Address PrevAddr;
    net::Address NextAddr;
    std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;
    // Not tested or benchmarked yet, don't use
    // ReplayProtection ServerToClientProtection;
    // ReplayProtection ClientToServerProtection;
    legacy::relay_replay_protection_t ServerToClientProtection;
    legacy::relay_replay_protection_t ClientToServerProtection;
  };

  inline Session::Session(const util::Clock& relayClock, const core::RouterInfo& routerInfo): Expireable(relayClock, routerInfo)
  {}

  using SessionPtr = std::shared_ptr<Session>;
}  // namespace core
#endif
