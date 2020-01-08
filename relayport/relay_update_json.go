package relayport

type RelayUpdateJSON struct {
	Metadata     RelayDataJSON
	Timestamp    uint64
	Signature    []byte
	PingStats    []RelayPingStatsJSON
	Usage        float32
	TrafficStats *TrafficStats
}
