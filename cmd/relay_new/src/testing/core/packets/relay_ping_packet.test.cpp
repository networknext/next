#include "includes.h"
#include "testing/test.hpp"

#include "encoding/write.hpp"

#include "core/packets/relay_ping_packet.hpp"

Test(core_packets_RelayPingPacket_general)
{
  uint8_t type = RELAY_PING_PACKET;
  auto seqnum = testing::Random<uint64_t>();
  net::Address addr = testing::RandomAddress();
  core::GenericPacket data;

  // get test data into buffer
  size_t index = 0;
  encoding::WriteUint8(data, index, type);
  encoding::WriteUint64(data, index, seqnum);
  encoding::WriteAddress(data, index, addr);

  const int size = RELAY_PING_PACKET_BYTES;
  core::packets::RelayPingPacket packet(data, size);

  // make sure packet is passed values correctly
  check(packet.Data[0] == RELAY_PING_PACKET);
  check(packet.Size == size);

  // ensure getters work
  check(packet.getSeqNum() == seqnum);
  check(packet.getFromAddr() == addr);

  // create another addr
  addr = testing::RandomAddress();

  // write it to the buffer
  packet.writeFromAddr(addr);

  // since the first packet has its addr filled, create another for reading purposes
  core::packets::RelayPingPacket followup(data, size);

  // ensure the new addr was written to the buffer
  check(followup.getFromAddr() == addr);
}