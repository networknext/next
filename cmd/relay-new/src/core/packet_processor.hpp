#ifndef CORE_PACKET_PROCESSOR_HPP
#define CORE_PACKET_PROCESSOR_HPP

#include "os/platform.hpp"

#include "util/throughput_logger.hpp"

#include "relay/relay.hpp"

namespace core
{
  class PacketProcessor
  {
   public:
    PacketProcessor(os::Socket& socket, relay::relay_t& relay, volatile bool& handle, util::ThroughputLogger& logger);
    ~PacketProcessor();

    void listen();

    void stop();

   private:
    os::Socket& mSocket;
    relay::relay_t& mRelay;
    volatile bool& mHandle;

    util::ThroughputLogger& mLogger;

    // Marks the first byte as a pong packet and sends it back
    void handleRelayPingPacket(
     std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);

    // Processes the pong packet by increasing the sequence number and getting the time diff
    void handleRelayPongPacket(
     std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);

    void handleRouteRequestPacket(
     std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);

    void handleRouteResponsePacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleContinueRequestPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleContinueResponsePacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleClientToServerPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleServerToClientPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleSessionPingPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleSessionPongPacket(std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size);

    void handleNearPingPacket(
     std::array<uint8_t, RELAY_MAX_PACKET_BYTES>& packet, const int size, legacy::relay_address_t& from);
  };
}  // namespace core
#endif