#pragma once

#include "util/macros.hpp"
#include "util/logger.hpp"
#include "ping_history.hpp"

namespace testing
{
  class _test_core_ReplayProtection_all_;
}

namespace core
{
  const size_t REPLAY_PROTECTION_BUFFER_SIZE = 256UL;

  class ReplayProtection
  {
    friend testing::_test_core_ReplayProtection_all_;

   public:
    ReplayProtection() = default;
    ~ReplayProtection() = default;

    void reset();
    bool is_already_received(uint64_t incoming_sequence);
    void advance_sequence_to(uint64_t incoming_sequence);

   private:
    uint64_t most_recent_sequence;
    std::array<uint64_t, REPLAY_PROTECTION_BUFFER_SIZE> received_packets;
  };

  INLINE void ReplayProtection::reset()
  {
    this->most_recent_sequence = 0;
    this->received_packets.fill(INVALID_SEQUENCE_NUMBER);
  }

  INLINE bool ReplayProtection::is_already_received(uint64_t incoming_sequence)
  {
    if (incoming_sequence + REPLAY_PROTECTION_BUFFER_SIZE <= this->most_recent_sequence) {
      return true;
    }

    uint64_t index = incoming_sequence % REPLAY_PROTECTION_BUFFER_SIZE;

    if (this->received_packets[index] == INVALID_SEQUENCE_NUMBER) {
      return false;
    }

    if (this->received_packets[index] >= incoming_sequence) {
      return true;
    }

    return false;
  }

  INLINE void ReplayProtection::advance_sequence_to(uint64_t incoming_sequence)
  {
    if (incoming_sequence > this->most_recent_sequence) {
      this->most_recent_sequence = incoming_sequence;
    }

    auto index = incoming_sequence % REPLAY_PROTECTION_BUFFER_SIZE;

    this->received_packets[index] = incoming_sequence;
  }
}  // namespace core
