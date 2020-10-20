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
    RouteRequest = 100,
    RouteResponse = 101,
    ClientToServer = 102,
    ServerToClient = 103,
    SessionPing = 104,
    SessionPong = 105,
    ContinueRequest = 106,
    ContinueResponse = 107,
    NearPing = 116,
    NearPong = 117,
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
      PACKET_TYPE_SWITCH_MACRO(RouteRequest);
      PACKET_TYPE_SWITCH_MACRO(RouteResponse);
      PACKET_TYPE_SWITCH_MACRO(ClientToServer);
      PACKET_TYPE_SWITCH_MACRO(ServerToClient);
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