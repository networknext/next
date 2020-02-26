#ifndef CORE_HANDLERS_BASE_HANDLER_HPP
#define CORE_HANDLERS_BASE_HANDLER_HPP

#include "core/packet.hpp"
#include "core/router_info.hpp"

#include "util/clock.hpp"

namespace core
{
  namespace handlers
  {
    class BaseHandler
    {
     protected:
      BaseHandler(const util::Clock& relayClock, const RouterInfo& routerInfo, GenericPacket& packet, const int packetSize);

      GenericPacket& mPacket;
      const int mPacketSize;

      inline auto tokenIsExpired(core::Token& token) -> bool;
      inline auto sessionIsExpired(core::SessionPtr session) -> bool;

     private:
      const util::Clock& mRelayClock;
      const RouterInfo& mRouterInfo;

      inline auto timestamp() -> uint64_t;
    };

    inline BaseHandler::BaseHandler(
     const util::Clock& relayClock, const RouterInfo& routerInfo, GenericPacket& packet, const int packetSize)
     : mPacket(packet), mPacketSize(packetSize), mRelayClock(relayClock), mRouterInfo(routerInfo)
    {}

    inline auto BaseHandler::tokenIsExpired(core::Token& token) -> bool
    {
      return token.ExpireTimestamp < timestamp();
    }

    inline auto BaseHandler::sessionIsExpired(core::SessionPtr session) -> bool
    {
      return session->ExpireTimestamp < timestamp();
    }

    inline auto BaseHandler::timestamp() -> uint64_t
    {
      auto seconds_since_initialize = mRelayClock.elapsed<util::Second>();
      return mRouterInfo.InitalizeTimeInSeconds + seconds_since_initialize;
    }
  }  // namespace handlers
}  // namespace core
#endif