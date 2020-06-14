#ifndef CORE_SESSION_HPP
#define CORE_SESSION_HPP

#include "expireable.hpp"
#include "net/address.hpp"
#include "replay_protection.hpp"
#include "util/logger.hpp"
#include "router_info.hpp"

namespace core
{
  class Session: public Expireable
  {
   public:
    Session(const RouterInfo& routerInfo);
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

  inline Session::Session(const RouterInfo& routerInfo): Expireable(routerInfo) {}

  using SessionPtr = std::shared_ptr<Session>;

  inline std::ostream& operator<<(std::ostream& os, const Session& session)
  {
    return os << std::hex << session.SessionID << '.' << std::dec << static_cast<unsigned int>(session.SessionVersion);
  }
}  // namespace core
#endif
