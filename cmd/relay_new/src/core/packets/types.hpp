#pragma once

namespace core
{
  namespace packets
  {
    enum class Type : uint8_t
    {
      RouteRequest = 1,
      RouteResponse = 2,
      ClientToServer = 3,
      ServerToClient = 4,
      OldRelayPing = 5,
      OldRelayPong = 6,
      NewRelayPing = 7,
      NewRelayPong = 8,
      SessionPing = 11,
      SessionPong = 12,
      ContinueRequest = 13,
      ContinueResponse = 14,
      V3BackendUpdateResponse = 49,
      V3BackendConfigResponse = 51,
      V3BackendInitResponse = 52,
      NearPing = 73,  // client -> relay
      NearPong = 74,  // relay -> client
    };

    template <typename T>
    inline auto operator==(Type t, T other) -> bool
    {
      return t == static_cast<Type>(other);
    }

    template <typename T>
    inline auto operator!=(Type t, T other) -> bool
    {
      return !(t == other);
    }
  }  // namespace packets
}  // namespace core