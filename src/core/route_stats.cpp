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

    double meanRTT = 0.0;
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
            meanRTT += 1000.0 * (entry.TimePongReceived - entry.TimePingSent);
            numPongs++;
          }
        }
      }
    }

    meanRTT = (numPongs > 0) ? (meanRTT / numPongs) : 10000.0;
    assert(meanRTT >= 0.0);

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

namespace legacy
{
  void relay_route_stats_from_ping_history(
   const legacy::relay_ping_history_t* history, double start, double end, relay_route_stats_t* stats, double ping_safety)
  {
    assert(history);
    assert(stats);
    assert(start < end);

    stats->rtt = 0.0f;
    stats->jitter = 0.0f;
    stats->packet_loss = 0.0f;

    // calculate packet loss

    int num_pings_sent = 0;
    int num_pongs_received = 0;

    for (int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++) {
      const relay_ping_history_entry_t* entry = &history->entries[i];

      if (entry->time_ping_sent >= start && entry->time_ping_sent <= end - ping_safety) {
        num_pings_sent++;

        if (entry->time_pong_received >= entry->time_ping_sent)
          num_pongs_received++;
      }
    }

    if (num_pings_sent > 0) {
      stats->packet_loss = (float)(100.0 * (1.0 - (double(num_pongs_received) / double(num_pings_sent))));
    }

    // calculate mean RTT

    double mean_rtt = 0.0;
    int num_pings = 0;
    int num_pongs = 0;

    for (int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++) {
      const relay_ping_history_entry_t* entry = &history->entries[i];

      if (entry->time_ping_sent >= start && entry->time_ping_sent <= end) {
        if (entry->time_pong_received > entry->time_ping_sent) {
          mean_rtt += 1000.0 * (entry->time_pong_received - entry->time_ping_sent);
          num_pongs++;
        }
        num_pings++;
      }
    }

    mean_rtt = (num_pongs > 0) ? (mean_rtt / num_pongs) : 10000.0;

    assert(mean_rtt >= 0.0);

    stats->rtt = float(mean_rtt);

    // calculate jitter

    int num_jitter_samples = 0;

    double stddev_rtt = 0.0;

    for (int i = 0; i < RELAY_PING_HISTORY_ENTRY_COUNT; i++) {
      const relay_ping_history_entry_t* entry = &history->entries[i];

      if (entry->time_ping_sent >= start && entry->time_ping_sent <= end) {
        if (entry->time_pong_received > entry->time_ping_sent) {
          // pong received
          double rtt = 1000.0 * (entry->time_pong_received - entry->time_ping_sent);
          if (rtt >= mean_rtt) {
            double error = rtt - mean_rtt;
            stddev_rtt += error * error;
            num_jitter_samples++;
          }
        }
      }
    }

    if (num_jitter_samples > 0) {
      stats->jitter = 3.0f * (float)sqrt(stddev_rtt / num_jitter_samples);
    }
  }
}  // namespace legacy