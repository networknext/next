#include "includes.h"
#include "testing/test.hpp"

#include "encoding/write.hpp"

#include "core/packets/relay_ping_packet.hpp"

Test(core_packets_RelayPingPacket_general)
{
  uint8_t type = RELAY_PING_PACKET;
  auto seqnum = testing::Random<uint64_t>();
  net::Address addr = testing::RandomAddress();
  core::GenericPacket<> packet;

  // get test data into buffer
  size_t index = 0;
  encoding::WriteUint8(packet.Buffer, index, type);
  encoding::WriteUint64(packet.Buffer, index, seqnum);
  encoding::WriteAddress(packet.Buffer, index, addr);

  const int size = RELAY_PING_PACKET_BYTES;
  core::packets::RelayPingPacket pingPacket(packet, size);

  // make sure packet is passed values correctly
  check(pingPacket.Internal.Buffer[0] == RELAY_PING_PACKET);
  check(pingPacket.Size == size);

  // ensure getters work
  check(pingPacket.getSeqNum() == seqnum);
  check(pingPacket.getFromAddr() == addr);

  // create another addr
  addr = testing::RandomAddress();

  // write it to the buffer
  pingPacket.writeFromAddr(addr);

  // since the first packet has its addr filled, create another for reading purposes
  core::packets::RelayPingPacket followup(packet, size);

  // ensure the new addr was written to the buffer
  check(followup.getFromAddr() == addr);
}