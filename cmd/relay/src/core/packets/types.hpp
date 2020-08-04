#pragma once

#define PACKET_TYPE_SWITCH_MACRO(type) \
  case Type ::type: {                  \
    str = #type;                       \
  } break

namespace core
{
  namespace packets
  {
    enum class Type : uint8_t
    {
      None = 0,
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
      V3InitRequest = 43,
      V3UpdateRequest = 48,
      V3BackendUpdateResponse = 49,
      V3ConfigRequest = 50,
      V3BackendConfigResponse = 51,
      V3BackendInitResponse = 52,
      NearPing = 73,
      NearPong = 74,
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

    inline std::ostream& operator<<(std::ostream& os, const Type& type)
    {
      std::string str;
      switch (type) {
        PACKET_TYPE_SWITCH_MACRO(RouteRequest);
        PACKET_TYPE_SWITCH_MACRO(RouteResponse);
        PACKET_TYPE_SWITCH_MACRO(ClientToServer);
        PACKET_TYPE_SWITCH_MACRO(ServerToClient);
        PACKET_TYPE_SWITCH_MACRO(OldRelayPing);
        PACKET_TYPE_SWITCH_MACRO(OldRelayPong);
        PACKET_TYPE_SWITCH_MACRO(NewRelayPing);
        PACKET_TYPE_SWITCH_MACRO(NewRelayPong);
        PACKET_TYPE_SWITCH_MACRO(SessionPing);
        PACKET_TYPE_SWITCH_MACRO(SessionPong);
        PACKET_TYPE_SWITCH_MACRO(ContinueRequest);
        PACKET_TYPE_SWITCH_MACRO(ContinueResponse);
        PACKET_TYPE_SWITCH_MACRO(V3InitRequest);
        PACKET_TYPE_SWITCH_MACRO(V3UpdateRequest);
        PACKET_TYPE_SWITCH_MACRO(V3BackendUpdateResponse);
        PACKET_TYPE_SWITCH_MACRO(V3ConfigRequest);
        PACKET_TYPE_SWITCH_MACRO(V3BackendConfigResponse);
        PACKET_TYPE_SWITCH_MACRO(V3BackendInitResponse);
        PACKET_TYPE_SWITCH_MACRO(NearPing);
        PACKET_TYPE_SWITCH_MACRO(NearPong);
        default: {
          str = "Unknown";
        } break;
      }

      return os << str << " (" << static_cast<uint32_t>(type) << ')';
    }
  }  // namespace packets
}  // namespace core