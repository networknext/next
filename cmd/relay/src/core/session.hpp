#ifndef CORE_SESSION_HPP
#define CORE_SESSION_HPP

#include <sodium.h>
#include <cinttypes>

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
    uint8_t private_key[crypto_box_SECRETKEYBYTES];
    ReplayProtection ServerToClientProtection;
    ReplayProtection ClientToServerProtection;
  };
}  // namespace core
#endif
