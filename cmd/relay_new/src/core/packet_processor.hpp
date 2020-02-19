#ifndef CORE_PACKET_PROCESSOR_HPP
#define CORE_PACKET_PROCESSOR_HPP

#include "session.hpp"

#include "os/platform.hpp"

#include "util/throughput_logger.hpp"

#include "relay/relay.hpp"

namespace core
{
  class PacketProcessor
  {
   public:
    PacketProcessor(core::SessionMap& sessions, relay::relay_t& relay, volatile bool& handle, util::ThroughputLogger* logger);
    ~PacketProcessor() = default;

    void listen(os::Socket& socket);

   private:
    core::SessionMap& mSessionMap;
    relay::relay_t& mRelay;
    volatile bool& mShouldProcess;

    util::ThroughputLogger* mLogger;

    // Marks the first byte as a pong packet and sends it back
    void handleRelayPingPacket(
     os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);

    // Processes the pong packet by increasing the sequence number and getting the time diff
    void handleRelayPongPacket(
     std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);

    void handleRouteRequestPacket(
     os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);

    void handleRouteResponsePacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);

    void handleContinueRequestPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleContinueResponsePacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleClientToServerPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleServerToClientPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleSessionPingPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleSessionPongPacket(os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleNearPingPacket(
     os::Socket& socket, std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);
  };
}  // namespace core
#endif