#pragma once

namespace core
{
  namespace packets
  {
    enum class Type : uint8_t
    {
      RelayPing = 75,
      RelayPong = 76,
      RouteRequest = 1,
      RouteResponse = 2,
      ClientToServer = 3,
      ServerToClient = 4,
      SessionPing = 11,
      SessionPong = 12,
      ContinueRequest = 13,
      ContinueResponse = 14,
      NearPing = 73,
      NearPong = 74,

      V3BackendUpdateResponse = 49,
      V3BackendConfigResponse = 51,
      V3BackendInitResponse = 52,
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