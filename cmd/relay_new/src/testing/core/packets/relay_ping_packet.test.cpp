#include "includes.h"
#include "testing/test.hpp"

#include "core/packets/relay_ping_packet.hpp"
#include "core/packets/types.hpp"
#include "crypto/bytes.hpp"
#include "encoding/write.hpp"

Test(core_packets_RelayPingPacket_general)
{
  uint8_t type = static_cast<uint8_t>(core::packets::Type::RelayPing);
  auto seqnum = crypto::Random<uint64_t>();
  net::Address addr = testing::RandomAddress();
  core::GenericPacket<> packet;

  // get test data into buffer
  size_t index = 0;
  encoding::WriteUint8(packet.Buffer, index, type);
  encoding::WriteUint64(packet.Buffer, index, seqnum);
  encoding::WriteAddress(packet.Buffer, index, addr);

  packet.Len = core::packets::RelayPingPacket::ByteSize;
  core::packets::RelayPingPacket pingPacket(packet);

  // make sure packet is passed values correctly
  check(pingPacket.Internal.Buffer[0] == static_cast<uint8_t>(core::packets::Type::RelayPing));

  // ensure getters work
  check(pingPacket.getSeqNum() == seqnum);
  check(pingPacket.getFromAddr() == addr);

  // create another addr
  addr = testing::RandomAddress();

  // write it to the buffer
  pingPacket.writeFromAddr(addr);

  // since the first packet has its addr filled, create another for reading purposes
  core::packets::RelayPingPacket followup(packet);

  // ensure the new addr was written to the buffer
  check(followup.getFromAddr() == addr);
}