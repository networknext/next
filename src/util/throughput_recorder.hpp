#pragma once

#include "util/console.hpp"

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

  class ThroughputRecorder
  {
   public:
    ThroughputRecorder() = default;
    ThroughputRecorder(ThroughputRecorder&& other);
    ~ThroughputRecorder() = default;

    // packets sent out via ping processor
    ThroughputStats OutboundPingTx;

    ThroughputStats RouteRequestRx;
    ThroughputStats RouteRequestTx;

    ThroughputStats RouteResponseRx;
    ThroughputStats RouteResponseTx;

    ThroughputStats ClientToServerRx;
    ThroughputStats ClientToServerTx;

    ThroughputStats ServerToClientRx;
    ThroughputStats ServerToClientTx;

    ThroughputStats InboundPingRx;
    ThroughputStats InboundPingTx;

    ThroughputStats PongRx;

    ThroughputStats SessionPingRx;
    ThroughputStats SessionPingTx;

    ThroughputStats SessionPongRx;
    ThroughputStats SessionPongTx;

    ThroughputStats ContinueRequestRx;
    ThroughputStats ContinueRequestTx;

    ThroughputStats ContinueResponseRx;
    ThroughputStats ContinueResponseTx;

    ThroughputStats NearPingRx;
    ThroughputStats NearPingTx;

    ThroughputStats UnknownRx;
  };

  inline ThroughputStats::ThroughputStats(ThroughputStats&& other)
   : PacketCount(other.PacketCount.exchange(0)), ByteCount(other.ByteCount.exchange(0))
  {}

  inline ThroughputRecorder::ThroughputRecorder(ThroughputRecorder&& other)
   : OutboundPingTx(std::move(other.OutboundPingTx)),
     RouteRequestRx(std::move(other.RouteRequestRx)),
     RouteRequestTx(std::move(other.RouteRequestTx)),
     RouteResponseRx(std::move(other.RouteResponseRx)),
     RouteResponseTx(std::move(other.RouteResponseTx)),
     ClientToServerRx(std::move(other.ClientToServerRx)),
     ClientToServerTx(std::move(other.ClientToServerTx)),
     ServerToClientRx(std::move(other.ServerToClientRx)),
     ServerToClientTx(std::move(other.ServerToClientTx)),
     InboundPingRx(std::move(other.InboundPingRx)),
     InboundPingTx(std::move(other.InboundPingTx)),
     PongRx(std::move(other.PongRx)),
     SessionPingRx(std::move(other.SessionPingRx)),
     SessionPingTx(std::move(other.SessionPingTx)),
     SessionPongRx(std::move(other.SessionPongRx)),
     SessionPongTx(std::move(other.SessionPongTx)),
     ContinueRequestRx(std::move(other.ContinueRequestRx)),
     ContinueRequestTx(std::move(other.ContinueRequestTx)),
     ContinueResponseRx(std::move(other.ContinueResponseRx)),
     ContinueResponseTx(std::move(other.ContinueResponseTx)),
     NearPingRx(std::move(other.NearPingRx)),
     NearPingTx(std::move(other.NearPingTx)),
     UnknownRx(std::move(other.UnknownRx))
  {}

  [[gnu::always_inline]] inline void ThroughputStats::add(size_t count)
  {
    this->ByteCount += count;
    this->PacketCount++;
  }
}  // namespace util
