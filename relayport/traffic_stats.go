package relayport

// TrafficStats is a struct for a realay's traffic (and anything else?)
type TrafficStats struct {
	BytesPaidTx        uint64
	BytesPaidRx        uint64
	BytesManagementTx  uint64
	BytesManagementRx  uint64
	BytesMeasurementTx uint64
	BytesMeasurementRx uint64
	BytesInvalidRx     uint64
	SessionCount       uint64
	FlowCount          uint64
}

// Normalize is a "hack to handle rename from FlowCount -> SessionCount" quoted from Next repo
func (stats *TrafficStats) Normalize() {
	count := stats.SessionCount
	if stats.FlowCount > count {
		count = stats.FlowCount
	}
	stats.SessionCount = count
	stats.FlowCount = count
}
