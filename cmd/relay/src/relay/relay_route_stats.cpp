#include "includes.h"
#include "relay_route_stats.hpp"

namespace relay
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
      const legacy::relay_ping_history_entry_t* entry = &history->entries[i];

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
      const legacy::relay_ping_history_entry_t* entry = &history->entries[i];

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
      const legacy::relay_ping_history_entry_t* entry = &history->entries[i];

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
}  // namespace relay
