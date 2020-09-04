#pragma once

#define PACKET_TYPE_SWITCH_MACRO(type) \
  case Type ::type: {                  \
    str = #type;                       \
  } break

namespace core
{
  const size_t RELAY_PING_PACKET_SIZE = 1 + 8;  // type | sequence
  const size_t RELAY_MTU = 1300;
  const size_t RELAY_MAX_PACKET_BYTES = 1500;

  enum class Type : uint8_t
  {
    None = 0,
    RouteRequest = 1,
    RouteResponse = 2,
    ClientToServer = 3,
    ServerToClient = 4,
    RelayPing = 7,
    RelayPong = 8,
    SessionPing = 11,
    SessionPong = 12,
    ContinueRequest = 13,
    ContinueResponse = 14,
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
      PACKET_TYPE_SWITCH_MACRO(RelayPing);
      PACKET_TYPE_SWITCH_MACRO(RelayPong);
      PACKET_TYPE_SWITCH_MACRO(SessionPing);
      PACKET_TYPE_SWITCH_MACRO(SessionPong);
      PACKET_TYPE_SWITCH_MACRO(ContinueRequest);
      PACKET_TYPE_SWITCH_MACRO(ContinueResponse);
      PACKET_TYPE_SWITCH_MACRO(NearPing);
      PACKET_TYPE_SWITCH_MACRO(NearPong);
      default: {
        str = "Unknown";
      } break;
    }

    return os << str << " (" << static_cast<uint32_t>(type) << ')';
  }
}  // namespace core