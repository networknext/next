#ifndef CORE_PACKET_HPP
#define CORE_PACKET_HPP
namespace core
{
  using GenericPacket = std::array<uint8_t, RELAY_MAX_PACKET_BYTES>;
}
#endif