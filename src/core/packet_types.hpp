#pragma once

#define PACKET_TYPE_SWITCH_MACRO(type) \
  case PacketType ::type: {            \
    str = #type;                       \
  } break

namespace core
{
  const size_t RELAY_PING_PACKET_SIZE = 1 + 8;  // type | sequence
  const size_t RELAY_MTU = 1300;
  const size_t RELAY_MAX_PACKET_BYTES = 1500;

  enum class PacketType : uint8_t
  {
    None = 0,
    RelayPing = 7,
    RelayPong = 8,
    RouteRequest4 = 100,
    RouteResponse4 = 101,
    ClientToServer4 = 102,
    ServerToClient4 = 103,
    SessionPing4 = 104,
    SessionPong4 = 105,
    ContinueRequest4 = 106,
    ContinueResponse4 = 107,
    NearPing4 = 116,
    NearPong4 = 117,
  };

  template <typename T>
  INLINE auto operator==(PacketType t, T other) -> bool
  {
    return t == static_cast<PacketType>(other);
  }

  template <typename T>
  INLINE auto operator!=(PacketType t, T other) -> bool
  {
    return !(t == other);
  }

  INLINE std::ostream& operator<<(std::ostream& os, const PacketType& type)
  {
    std::string str;
    switch (type) {
      PACKET_TYPE_SWITCH_MACRO(RelayPing);
      PACKET_TYPE_SWITCH_MACRO(RelayPong);
      PACKET_TYPE_SWITCH_MACRO(RouteRequest4);
      PACKET_TYPE_SWITCH_MACRO(RouteResponse4);
      PACKET_TYPE_SWITCH_MACRO(ClientToServer4);
      PACKET_TYPE_SWITCH_MACRO(ServerToClient4);
      PACKET_TYPE_SWITCH_MACRO(SessionPing4);
      PACKET_TYPE_SWITCH_MACRO(SessionPong4);
      PACKET_TYPE_SWITCH_MACRO(ContinueRequest4);
      PACKET_TYPE_SWITCH_MACRO(ContinueResponse4);
      PACKET_TYPE_SWITCH_MACRO(NearPing4);
      PACKET_TYPE_SWITCH_MACRO(NearPong4);
      default: {
        str = "Unknown";
      } break;
    }

    return os << str << " (" << static_cast<uint32_t>(type) << ')';
  }
}  // namespace core