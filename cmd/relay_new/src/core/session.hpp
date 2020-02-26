#ifndef CORE_SESSION_HPP
#define CORE_SESSION_HPP

#include "replay_protection.hpp"
#include "net/address.hpp"

namespace core
{
  struct Session
  {
    uint64_t ExpireTimestamp;
    uint64_t SessionID;
    uint8_t SessionVersion;
    uint64_t ClientToServerSeq;
    uint64_t ServerToClientSeq;
    int KbpsUp;
    int KbpsDown;
    net::Address PrevAddr;
    net::Address NextAddr;
    std::array<uint8_t, crypto_box_SECRETKEYBYTES> PrivateKey;
    // ReplayProtection ServerToClientProtection;
    // ReplayProtection ClientToServerProtection;
    legacy::relay_replay_protection_t ServerToClientProtection;
    legacy::relay_replay_protection_t ClientToServerProtection;
  };

  using SessionPtr = std::shared_ptr<Session>;
  class SessionMap: public std::map<uint64_t, SessionPtr>
  {
   public:
    std::mutex Lock;
  };
}  // namespace core
#endif
