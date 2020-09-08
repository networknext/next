#pragma once

#include "util/console.hpp"
#include "util/macros.hpp"

namespace util
{
  struct ThroughputStats
  {
    ThroughputStats() = default;
    ThroughputStats(ThroughputStats&& other);

    void add(size_t count);

    std::atomic<size_t> num_packets = 0;
    std::atomic<size_t> num_bytes = 0;
  };

  class ThroughputRecorder
  {
   public:
    ThroughputRecorder() = default;
    ThroughputRecorder(ThroughputRecorder&& other);
    ~ThroughputRecorder() = default;

    ThroughputStats outbound_ping_tx;

    ThroughputStats route_request_rx;
    ThroughputStats route_request_tx;

    ThroughputStats route_response_rx;
    ThroughputStats route_response_tx;

    ThroughputStats client_to_server_rx;
    ThroughputStats client_to_server_tx;

    ThroughputStats server_to_client_rx;
    ThroughputStats server_to_client_tx;

    ThroughputStats inbound_ping_rx;
    ThroughputStats inbound_ping_tx;

    ThroughputStats pong_rx;

    ThroughputStats session_ping_rx;
    ThroughputStats session_ping_tx;

    ThroughputStats session_pong_rx;
    ThroughputStats session_pong_tx;

    ThroughputStats continue_request_rx;
    ThroughputStats continue_request_tx;

    ThroughputStats continue_response_rx;
    ThroughputStats continue_response_tx;

    ThroughputStats near_ping_rx;
    ThroughputStats near_ping_tx;

    ThroughputStats unknown_rx;
  };

  INLINE ThroughputStats::ThroughputStats(ThroughputStats&& other)
   : num_packets(other.num_packets.exchange(0)), num_bytes(other.num_bytes.exchange(0))
  {}

  INLINE ThroughputRecorder::ThroughputRecorder(ThroughputRecorder&& other)
   : outbound_ping_tx(std::move(other.outbound_ping_tx)),
     route_request_rx(std::move(other.route_request_rx)),
     route_request_tx(std::move(other.route_request_tx)),
     route_response_rx(std::move(other.route_response_rx)),
     route_response_tx(std::move(other.route_response_tx)),
     client_to_server_rx(std::move(other.client_to_server_rx)),
     client_to_server_tx(std::move(other.client_to_server_tx)),
     server_to_client_rx(std::move(other.server_to_client_rx)),
     server_to_client_tx(std::move(other.server_to_client_tx)),
     inbound_ping_rx(std::move(other.inbound_ping_rx)),
     inbound_ping_tx(std::move(other.inbound_ping_tx)),
     pong_rx(std::move(other.pong_rx)),
     session_ping_rx(std::move(other.session_ping_rx)),
     session_ping_tx(std::move(other.session_ping_tx)),
     session_pong_rx(std::move(other.session_pong_rx)),
     session_pong_tx(std::move(other.session_pong_tx)),
     continue_request_rx(std::move(other.continue_request_rx)),
     continue_request_tx(std::move(other.continue_request_tx)),
     continue_response_rx(std::move(other.continue_response_rx)),
     continue_response_tx(std::move(other.continue_response_tx)),
     near_ping_rx(std::move(other.near_ping_rx)),
     near_ping_tx(std::move(other.near_ping_tx)),
     unknown_rx(std::move(other.unknown_rx))
  {}

  INLINE void ThroughputStats::add(size_t count)
  {
    this->num_bytes += count;
    this->num_packets++;
  }
}  // namespace util
