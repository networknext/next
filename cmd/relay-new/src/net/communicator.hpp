#ifndef NET_COMMUNICATOR_HPP
#define NET_COMMUNICATOR_HPP

#include "relay/relay.hpp"
#include "util/throughput_logger.hpp"

namespace net
{
  class Communicator
  {
   public:
    Communicator(relay::relay_t& relay, volatile bool& handle, std::ostream& ouptut = std::cout);
    ~Communicator();

    void stop();

   private:
    relay::relay_t& mRelay;
    volatile bool& mHandle;

    std::unique_ptr<std::thread> mPingThread;
    std::unique_ptr<std::thread> mRecvThread;

    util::ThroughputLogger mLogger;

    void initPingThread();
    void initRecvThread();

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
}  // namespace net
#endif