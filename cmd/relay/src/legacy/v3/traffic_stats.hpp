#pragma once

namespace legacy
{
  namespace v3
  {
    // member names ripped from old repo, actual per/sec should be per 10 secs
    struct TrafficStats
    {
      std::atomic<size_t> BytesPerSecPaidTx;
      std::atomic<size_t> BytesPerSecPaidRx;
      std::atomic<size_t> BytesPerSecManagementTx;
      std::atomic<size_t> BytesPerSecManagementRx;
      std::atomic<size_t> BytesPerSecMeasurementTx;
      std::atomic<size_t> BytesPerSecMeasurementRx;
      std::atomic<size_t> BytesPerSecInvalidRx;
    };
  }  // namespace v3
}  // namespace legacy