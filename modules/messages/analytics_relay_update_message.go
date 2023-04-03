package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/encoding"
)

const (
	AnalyticsRelayUpdateMessageVersion_Min   = 3
	AnalyticsRelayUpdateMessageVersion_Max   = 3
	AnalyticsRelayUpdateMessageVersion_Write = 3
)

type AnalyticsRelayUpdateMessage struct {
	Version                   uint8
	Timestamp                 uint64
	RelayId                   uint64
	SessionCount              uint32
	MaxSessions               uint32
	EnvelopeBandwidthUpKbps   uint32
	EnvelopeBandwidthDownKbps uint32
	ActualBandwidthUpKbps     uint32
	ActualBandwidthDownKbps   uint32
	RelayFlags                uint64
	NumRoutable               uint32
	NumUnroutable             uint32
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

	if !encoding.ReadUint32(buffer, &index, &message.ActualBandwidthUpKbps) {
		return fmt.Errorf("failed to read actual bandwidth up kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.ActualBandwidthDownKbps) {
		return fmt.Errorf("failed to read actual bandwidth down kbps")
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
	encoding.WriteUint32(buffer, &index, message.ActualBandwidthUpKbps)
	encoding.WriteUint32(buffer, &index, message.ActualBandwidthDownKbps)
	encoding.WriteUint64(buffer, &index, message.RelayFlags)
	encoding.WriteUint32(buffer, &index, message.NumRoutable)
	encoding.WriteUint32(buffer, &index, message.NumUnroutable)
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
	bigquery_message["actual_bandwidth_up_kbps"] = int(message.ActualBandwidthUpKbps)
	bigquery_message["actual_bandwidth_down_kbps"] = int(message.ActualBandwidthDownKbps)
	if message.RelayFlags != 0 {
		bigquery_message["relay_flags"] = int(message.RelayFlags)
	}
	bigquery_message["num_routable"] = int(message.NumRoutable)
	bigquery_message["num_unroutable"] = int(message.NumUnroutable)

	relay_counters := make([]bigquery.Value, message.NumRelayCounters)
	for i := 0; i < int(message.NumRelayCounters); i++ {
		relay_counters[i] = int(message.RelayCounters[i])
	}
	bigquery_message["relay_counters"] = relay_counters

	return bigquery_message, "", nil
}
