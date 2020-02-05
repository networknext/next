#ifndef CORE_PING_HISTORY_HPP
#define CORE_PING_HISTORY_HPP

#include <cinttypes>

namespace core
{
  class PingHistory
  {
   public:
    PingHistory() = default;
    ~PingHistory() = default;

    void clear();

    uint64_t pingSent(double time);

    void pongReceived(uint64_t seq, double time);

   private:
    uint64_t mSeq;
    double mTimePingSent;
    double mTimePongRecieved;
  };
}  // namespace core
#endif