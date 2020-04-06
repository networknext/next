#pragma once

#include "util/console.hpp"

using namespace std::chrono_literals;

namespace util
{
  struct ThroughputStats
  {
    ThroughputStats() = default;

    void add(size_t count);

    size_t PacketCount = 0;
    size_t ByteCount = 0;

    void reset();
  };

  struct ThroughputStatsCollection
  {
    ThroughputStats Sent;
    ThroughputStats Received;
    ThroughputStats Unknown;

    void reset();
  };

  class ThroughputRecorder
  {
   public:
    ThroughputRecorder() = default;
    ~ThroughputRecorder() = default;

    void addToSent(size_t count);
    void addToReceived(size_t count);
    void addToUnknown(size_t count);

    auto get() const -> ThroughputStatsCollection;

    void reset();

   private:
    std::mutex mLock;

    ThroughputStatsCollection mStats;
  };

  inline void ThroughputStats::add(size_t count)
  {
    this->ByteCount += count;
    this->PacketCount++;
  }

  inline void ThroughputStats::reset()
  {
    ByteCount = 0;
    PacketCount = 0;
  }

  inline void ThroughputStatsCollection::reset()
  {
    Sent.reset();
    Received.reset();
    Unknown.reset();
  }

  inline auto ThroughputRecorder::get() const -> ThroughputStatsCollection
  {
    return mStats;
  }

  inline void ThroughputRecorder::reset()
  {
    mStats.reset();
  }

  inline void ThroughputRecorder::addToSent(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mStats.Sent.add(count);
  }

  inline void ThroughputRecorder::addToReceived(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mStats.Received.add(count);
  }

  inline void ThroughputRecorder::addToUnknown(size_t count)
  {
    std::lock_guard<std::mutex> lk(mLock);
    mStats.Unknown.add(count);
  }
}  // namespace util
