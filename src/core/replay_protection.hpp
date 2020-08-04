#ifndef CORE_REPLAY_PROTECTION_HPP
#define CORE_REPLAY_PROTECTION_HPP

namespace core
{
  class ReplayProtection
  {
   public:
    ReplayProtection() = default;
    ~ReplayProtection() = default;

    void reset();
    bool alreadyReceived(uint64_t seq);
    void advanceSeq(uint64_t seq);

    const uint64_t& mostRecentSeq();

   private:
    uint64_t mRecentSeq;
    std::array<uint64_t, RELAY_REPLAY_PROTECTION_BUFFER_SIZE> mReceivedPacket;
  };

  inline const uint64_t& ReplayProtection::mostRecentSeq()
  {
    return mRecentSeq;
  }
}  // namespace core

namespace legacy
{
  struct relay_replay_protection_t
  {
    uint64_t most_recent_sequence;
    uint64_t received_packet[RELAY_REPLAY_PROTECTION_BUFFER_SIZE];
  };

  void relay_replay_protection_reset(relay_replay_protection_t* replay_protection);

  int relay_replay_protection_already_received(relay_replay_protection_t* replay_protection, uint64_t sequence);

  void relay_replay_protection_advance_sequence(relay_replay_protection_t* replay_protection, uint64_t sequence);
}  // namespace legacy
#endif
