#pragma once

namespace legacy
{
  namespace v3
  {
    struct TrafficStats
    {
      size_t BytesPerSecPaidTx;
      size_t BytesPerSecPaidRx;
      size_t BytesPerSecManagementTx;
      size_t BytesPerSecManagementRx;
      size_t BytesPerSecMeasurementTx;
      size_t BytesPerSecMeasurementRx;
      size_t BytesPerSecInvalidRx
    };
  }  // namespace v3
}  // namespace legacy