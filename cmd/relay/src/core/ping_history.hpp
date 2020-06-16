#ifndef CORE_PING_HISTORY_HPP
#define CORE_PING_HISTORY_HPP

namespace core
{
  struct HistoryEntry
  {
    uint64_t Sequence = INVALID_SEQUENCE_NUMBER;
    double TimePingSent = -1.0;
    double TimePongReceived = -1.0;
  };

  class PingHistory
  {
   public:
    PingHistory() = default;
    PingHistory(const PingHistory& other);
    ~PingHistory() = default;

    void clear();

    auto pingSent(double time) -> uint64_t;

    void pongReceived(uint64_t seq, double time);

    // helper for testing only
    auto seq() -> uint64_t;

    auto operator[](size_t i) -> const HistoryEntry&;
    auto operator=(const PingHistory& other) -> PingHistory&;

   private:
    uint64_t mSeq = 0;
    std::array<HistoryEntry, RELAY_PING_HISTORY_ENTRY_COUNT> mEntries;

    friend class RouteStats;
  };

  [[gnu::always_inline]] inline PingHistory::PingHistory(const PingHistory& other)
  {
    *this = other;
  }

  inline void PingHistory::clear()
  {
    GCC_NO_OPT_OUT;
    mSeq = 0;

    mEntries.fill(HistoryEntry());
  }

  inline auto PingHistory::seq() -> uint64_t
  {
    return mSeq;
  }

  inline auto PingHistory::operator[](size_t i) -> const HistoryEntry&
  {
    return mEntries[i % mEntries.size()];
  }

  [[gnu::always_inline]] inline auto PingHistory::operator=(const PingHistory& other) -> PingHistory&
  {
    this->mSeq = other.mSeq;
    std::copy(other.mEntries.begin(), other.mEntries.end(), this->mEntries.begin());
    return *this;
  }
}  // namespace core

namespace legacy
{
  struct relay_ping_history_entry_t
  {
    uint64_t sequence;
    double time_ping_sent;
    double time_pong_received;
  };

  struct relay_ping_history_t
  {
    uint64_t sequence;
    relay_ping_history_entry_t entries[RELAY_PING_HISTORY_ENTRY_COUNT];
  };

  void relay_ping_history_clear(relay_ping_history_t* history);

  uint64_t relay_ping_history_ping_sent(relay_ping_history_t* history, double time);

  void relay_ping_history_pong_received(relay_ping_history_t* history, uint64_t sequence, double time);
}  // namespace legacy
#endif
