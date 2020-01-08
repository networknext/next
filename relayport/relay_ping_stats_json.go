package relayport

// RelayPingStatsJSON is json for ping stats
type RelayPingStatsJSON struct {
	RelayID    uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
}
