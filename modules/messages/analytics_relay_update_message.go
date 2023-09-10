package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsRelayUpdateMessageVersion_Min   = 0
	AnalyticsRelayUpdateMessageVersion_Max   = 0
	AnalyticsRelayUpdateMessageVersion_Write = 0
)

type AnalyticsRelayUpdateMessage struct {
	Version                   uint8
	Timestamp                 uint64
	RelayId                   uint64
	SessionCount              uint32
	MaxSessions               uint32
	EnvelopeBandwidthUpKbps   uint32
	EnvelopeBandwidthDownKbps uint32
	PacketsSentPerSecond      float32
	PacketsReceivedPerSecond  float32
	BandwidthSentKbps         float32
	BandwidthReceivedKbps     float32
	NearPingsPerSecond        float32
	RelayPingsPerSecond       float32
	RelayFlags                uint64
	NumRoutable               uint32
	NumUnroutable             uint32
	StartTime                 uint64
	CurrentTime               uint64
	NumRelayCounters          uint32
	RelayCounters             [constants.NumRelayCounters]uint64
}

func (message *AnalyticsRelayUpdateMessage) GetMaxSize() int {
	return 256 + 8*constants.NumRelayCounters
}

func (message *AnalyticsRelayUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to analytics read relay update version")
	}

	if message.Version < AnalyticsRelayUpdateMessageVersion_Min || message.Version > AnalyticsRelayUpdateMessageVersion_Max {
		return fmt.Errorf("invalid analytics relay update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayId) {
		return fmt.Errorf("failed to read relay id")
	}

	if !encoding.ReadUint32(buffer, &index, &message.SessionCount) {
		return fmt.Errorf("failed to read session count")
	}

	if !encoding.ReadUint32(buffer, &index, &message.MaxSessions) {
		return fmt.Errorf("failed to read max sessions")
	}

	if !encoding.ReadUint32(buffer, &index, &message.EnvelopeBandwidthUpKbps) {
		return fmt.Errorf("failed to read envelope bandwidth up kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.EnvelopeBandwidthDownKbps) {
		return fmt.Errorf("failed to read envelope bandwidth down kbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.PacketsSentPerSecond) {
		return fmt.Errorf("failed to read packets sent per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.PacketsReceivedPerSecond) {
		return fmt.Errorf("failed to read packets received per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthSentKbps) {
		return fmt.Errorf("failed to read bandwidth sent kbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthReceivedKbps) {
		return fmt.Errorf("failed to read bandwidth received kbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.NearPingsPerSecond) {
		return fmt.Errorf("failed to read near pings per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RelayPingsPerSecond) {
		return fmt.Errorf("failed to read relay pings per-second")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayFlags) {
		return fmt.Errorf("failed to read relay flags")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumRoutable) {
		return fmt.Errorf("failed to read num routable")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumUnroutable) {
		return fmt.Errorf("failed to read num unroutable")
	}

	if !encoding.ReadUint64(buffer, &index, &message.StartTime) {
		return fmt.Errorf("failed to read num start time")
	}

	if !encoding.ReadUint64(buffer, &index, &message.CurrentTime) {
		return fmt.Errorf("failed to read num current time")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumRelayCounters) {
		return fmt.Errorf("failed to read num relay counters")
	}

	for i := 0; i < int(message.NumRelayCounters); i++ {
		if !encoding.ReadUint64(buffer, &index, &message.RelayCounters[i]) {
			return fmt.Errorf("failed to read relay counter")
		}
	}

	return nil
}

func (message *AnalyticsRelayUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsRelayUpdateMessageVersion_Min || message.Version > AnalyticsRelayUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics relay update message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.RelayId)
	encoding.WriteUint32(buffer, &index, message.SessionCount)
	encoding.WriteUint32(buffer, &index, message.MaxSessions)

	encoding.WriteUint32(buffer, &index, message.EnvelopeBandwidthUpKbps)
	encoding.WriteUint32(buffer, &index, message.EnvelopeBandwidthDownKbps)
	encoding.WriteFloat32(buffer, &index, message.PacketsSentPerSecond)
	encoding.WriteFloat32(buffer, &index, message.PacketsReceivedPerSecond)
	encoding.WriteFloat32(buffer, &index, message.BandwidthSentKbps)
	encoding.WriteFloat32(buffer, &index, message.BandwidthReceivedKbps)
	encoding.WriteFloat32(buffer, &index, message.NearPingsPerSecond)
	encoding.WriteFloat32(buffer, &index, message.RelayPingsPerSecond)

	encoding.WriteUint64(buffer, &index, message.RelayFlags)
	encoding.WriteUint32(buffer, &index, message.NumRoutable)
	encoding.WriteUint32(buffer, &index, message.NumUnroutable)

	encoding.WriteUint64(buffer, &index, message.StartTime)
	encoding.WriteUint64(buffer, &index, message.CurrentTime)

	encoding.WriteUint32(buffer, &index, message.NumRelayCounters)
	for i := 0; i < int(message.NumRelayCounters); i++ {
		encoding.WriteUint64(buffer, &index, message.RelayCounters[i])
	}

	return buffer[:index]
}

func (message *AnalyticsRelayUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["relay_id"] = int(message.RelayId)
	bigquery_message["session_count"] = int(message.SessionCount)
	if message.MaxSessions != 0 {
		bigquery_message["max_sessions"] = int(message.MaxSessions)
	}

	bigquery_message["envelope_bandwidth_up_kbps"] = int(message.EnvelopeBandwidthUpKbps)
	bigquery_message["envelope_bandwidth_down_kbps"] = int(message.EnvelopeBandwidthDownKbps)
	bigquery_message["packets_sent_per_second"] = float64(message.PacketsSentPerSecond)
	bigquery_message["packets_received_per_second"] = float64(message.PacketsReceivedPerSecond)
	bigquery_message["bandwidth_sent_kbps"] = float64(message.BandwidthSentKbps)
	bigquery_message["bandwidth_received_kbps"] = float64(message.BandwidthReceivedKbps)
	bigquery_message["near_pings_per_second"] = float64(message.NearPingsPerSecond)
	bigquery_message["relay_pings_per_second"] = float64(message.RelayPingsPerSecond)

	if message.RelayFlags != 0 {
		bigquery_message["relay_flags"] = int(message.RelayFlags)
	}
	bigquery_message["num_routable"] = int(message.NumRoutable)
	bigquery_message["num_unroutable"] = int(message.NumUnroutable)

	bigquery_message["start_time"] = int(message.StartTime)
	bigquery_message["current_time"] = int(message.CurrentTime)

	relay_counters := make([]bigquery.Value, message.NumRelayCounters)
	for i := 0; i < int(message.NumRelayCounters); i++ {
		relay_counters[i] = int(message.RelayCounters[i])
	}
	bigquery_message["relay_counters"] = relay_counters

	return bigquery_message, "", nil
}
