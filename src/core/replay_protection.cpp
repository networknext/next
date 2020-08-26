#include "includes.h"
#include "replay_protection.hpp"

#include "util/logger.hpp"

namespace core
{
  void ReplayProtection::reset()
  {
    mRecentSeq = 0;
    mReceivedPacket.fill(INVALID_SEQUENCE_NUMBER);
  }

  bool ReplayProtection::alreadyReceived(uint64_t seq)
  {
    if (seq + RELAY_REPLAY_PROTECTION_BUFFER_SIZE <= mRecentSeq) {
      return true;
    }

    auto index = seq % RELAY_REPLAY_PROTECTION_BUFFER_SIZE;

    if (mReceivedPacket[index] == INVALID_SEQUENCE_NUMBER) {
      return false;
    }

    if (mReceivedPacket[index] >= seq) {
      return true;
    }

    return false;
  }

  void ReplayProtection::advanceSeq(uint64_t seq)
  {
    if (seq > mRecentSeq) {
      mRecentSeq = seq;
    }

    auto index = seq % RELAY_REPLAY_PROTECTION_BUFFER_SIZE;

    mReceivedPacket[index] = seq;
  }
}  // namespace core

namespace legacy
{
  void relay_replay_protection_reset(relay_replay_protection_t* replay_protection)
  {
    assert(replay_protection);
    replay_protection->most_recent_sequence = 0;
    memset(replay_protection->received_packet, 0xFF, sizeof(replay_protection->received_packet));
  }

  int relay_replay_protection_already_received(relay_replay_protection_t* replay_protection, uint64_t sequence)
  {
    assert(replay_protection);

    if (sequence + RELAY_REPLAY_PROTECTION_BUFFER_SIZE <= replay_protection->most_recent_sequence) {
      return 1;
    }

    int index = (int)(sequence % RELAY_REPLAY_PROTECTION_BUFFER_SIZE);

    if (replay_protection->received_packet[index] == 0xFFFFFFFFFFFFFFFFLL) {
      return 0;
    }

    if (replay_protection->received_packet[index] >= sequence) {
      return 1;
    }

    return 0;
  }

  void relay_replay_protection_advance_sequence(relay_replay_protection_t* replay_protection, uint64_t sequence)
  {
    assert(replay_protection);

    if (sequence > replay_protection->most_recent_sequence) {
      replay_protection->most_recent_sequence = sequence;
    }

    int index = (int)(sequence % RELAY_REPLAY_PROTECTION_BUFFER_SIZE);

    replay_protection->received_packet[index] = sequence;
  }
}  // namespace legacy
