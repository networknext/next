#pragma once

#include "util/console.hpp"

using namespace std::chrono_literals;

namespace util
{
  struct ThroughputStats
  {
    ThroughputStats() = default;
    ThroughputStats(ThroughputStats&& other);

    void add(size_t count);

    std::atomic<size_t> PacketCount = 0;
    std::atomic<size_t> ByteCount = 0;
  };

  struct ThroughputStatsCollection
  {
    ThroughputStatsCollection() = default;
    ThroughputStatsCollection(ThroughputStatsCollection&& other);

    ThroughputStats Sent;
    ThroughputStats Received;
    ThroughputStats Unknown;
  };

  class ThroughputRecorder
  {
   public:
    ThroughputRecorder() = default;
    ~ThroughputRecorder() = default;

    void addToSent(size_t count);
    void addToReceived(size_t count);
    void addToUnknown(size_t count);

    auto get() -> ThroughputStatsCollection&;

   private:
    ThroughputStatsCollection mStats;
  };

  inline ThroughputStats::ThroughputStats(ThroughputStats&& other)
   : PacketCount(other.PacketCount.exchange(0)), ByteCount(other.ByteCount.exchange(0))
  {}

  inline ThroughputStatsCollection::ThroughputStatsCollection(ThroughputStatsCollection&& other)
   : Sent(std::move(other.Sent)), Received(std::move(other.Received)), Unknown(std::move(other.Unknown))
  {}

  [[gnu::always_inline]] inline void ThroughputStats::add(size_t count)
  {
    this->ByteCount += count;
    this->PacketCount++;
  }

  [[gnu::always_inline]] inline auto ThroughputRecorder::get() -> ThroughputStatsCollection&
  {
    return mStats;
  }

  [[gnu::always_inline]] inline void ThroughputRecorder::addToSent(size_t count)
  {
    mStats.Sent.add(count);
  }

  [[gnu::always_inline]] inline void ThroughputRecorder::addToReceived(size_t count)
  {
    mStats.Received.add(count);
  }

  [[gnu::always_inline]] inline void ThroughputRecorder::addToUnknown(size_t count)
  {
    mStats.Unknown.add(count);
  }
}  // namespace util
