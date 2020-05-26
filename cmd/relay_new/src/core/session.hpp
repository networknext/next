#ifndef CORE_SESSION_HPP
#define CORE_SESSION_HPP

#include "expireable.hpp"
#include "net/address.hpp"
#include "replay_protection.hpp"
#include "util/logger.hpp"

namespace core
{
  class Session: public Expireable
  {
   public:
    Session(const util::Clock& relayClock);
    virtual ~Session() override = default;

    uint64_t SessionID;
    uint8_t SessionVersion;
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

    void setClientToServerSeq(uint64_t seq);
    void setServerToClientSeq(uint64_t seq);

    auto getClientToServerSeq() -> uint64_t;
    auto getServerToClientSeq() -> uint64_t;

   private:
    uint64_t mClientToServer = 0;
    uint64_t mServerToClientSeq = 0;
  };

  inline Session::Session(const util::Clock& relayClock)
   : Expireable(relayClock), mServerToClientSeq(0), mClientToServer(0)
  {}

  inline void Session::setClientToServerSeq(uint64_t seq)
  {
    LogDebug("setting session ping seq: ", seq);
    mClientToServer = seq;
  }

  inline void Session::setServerToClientSeq(uint64_t seq)
  {
    LogDebug("setting server to client seq: ", seq);
    mServerToClientSeq = seq;
  }

  inline auto Session::getClientToServerSeq() -> uint64_t
  {
    return mClientToServer;
  }

  inline auto Session::getServerToClientSeq() -> uint64_t
  {
    return mServerToClientSeq;
  }

  using SessionPtr = std::shared_ptr<Session>;
}  // namespace core
#endif
