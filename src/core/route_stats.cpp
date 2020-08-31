#include "includes.h"
#include "route_stats.hpp"

namespace core
{
  RouteStats::RouteStats(const core::PingHistory& ph, double start, double end, double safety)
   : mRTT(0), mJitter(-1.0), mPacketLoss(-1.0)
  {
    // Packet loss calc
    // and RTT calc

    size_t numPingsSent = 0u;
    size_t numPongsReceived = 0u;

    double totalRTT = 0.0;
    int numPings = 0;
    int numPongs = 0;

    for (const auto& entry : ph.mEntries) {
      if (entry.TimePingSent >= start) {
        if (entry.TimePingSent <= end - safety) {
          numPingsSent++;

          if (entry.TimePongReceived >= entry.TimePingSent) {
            numPongsReceived++;
          }
        }

        if (entry.TimePingSent <= end) {
          numPings++;

          if (entry.TimePongReceived > entry.TimePingSent) {
            totalRTT += 1000.0 * (entry.TimePongReceived - entry.TimePingSent);
            numPongs++;
          }
        }
      }
    }

    double meanRTT = (numPongs > 0) ? (totalRTT / numPongs) : 10000.0;

    if (numPingsSent > 0) {
      mPacketLoss = static_cast<float>(100.0 * (1.0 - (double(numPongsReceived) / double(numPingsSent))));
    }

    mRTT = static_cast<float>(meanRTT);

    // Jitter calc

    auto numJitterSamples = 0u;
    auto stdDevRTT = 0.0;

    for (const auto& entry : ph.mEntries) {
      // if the entry is within the window and the pong was received
      if (entry.TimePingSent >= start && entry.TimePingSent <= end && entry.TimePongReceived > entry.TimePingSent) {
        double rtt = 1000.0 * (entry.TimePongReceived - entry.TimePingSent);
        if (rtt >= meanRTT) {
          double error = rtt - meanRTT;
          stdDevRTT += error * error;
          numJitterSamples++;
        }
      }
    }

    if (numJitterSamples > 0) {
      mJitter = 3.0f * static_cast<float>(std::sqrt(stdDevRTT / numJitterSamples));
    }
  }
}  // namespace core
