package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/encoding"
)

const (
	RelayUpdateMessageVersion_Min   = 3
	RelayUpdateMessageVersion_Max   = 3
	RelayUpdateMessageVersion_Write = 3

	MaxRelayUpdateMessageSize = 2048
)

type RelayUpdateMessage struct {
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
	NumRelayCounters          uint32
	RelayCounters             [constants.NumRelayCounters]uint64
}

func (message *RelayUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read relay update version")
	}

	if message.Version < RelayUpdateMessageVersion_Min || message.Version > RelayUpdateMessageVersion_Max {
		return fmt.Errorf("invalid relay update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read relay update timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayId) {
		return fmt.Errorf("failed to read relay update relay id")
	}

	if !encoding.ReadUint32(buffer, &index, &message.SessionCount) {
		return fmt.Errorf("failed to read relay update session count")
	}

	if !encoding.ReadUint32(buffer, &index, &message.MaxSessions) {
		return fmt.Errorf("failed to read relay update max sessions")
	}

	if !encoding.ReadUint32(buffer, &index, &message.EnvelopeBandwidthUpKbps) {
		return fmt.Errorf("failed to read relay update envelope bandwidth up kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.EnvelopeBandwidthDownKbps) {
		return fmt.Errorf("failed to read relay update envelope bandwidth down kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.ActualBandwidthUpKbps) {
		return fmt.Errorf("failed to read relay update actual bandwidth up kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.ActualBandwidthDownKbps) {
		return fmt.Errorf("failed to read relay update actual bandwidth down kbps")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayFlags) {
		return fmt.Errorf("failed to read relay update relay flags")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumRelayCounters) {
		return fmt.Errorf("failed to read relay update num relay counters")
	}

	for i := 0; i < int(message.NumRelayCounters); i++ {
		if !encoding.ReadUint64(buffer, &index, &message.RelayCounters[i]) {
			return fmt.Errorf("failed to read relay update relay counter")
		}
	}

	return nil
}

func (message *RelayUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < RelayUpdateMessageVersion_Min || message.Version > RelayUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid relay update message version %d", message.Version))
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
	encoding.WriteUint32(buffer, &index, message.NumRelayCounters)

	for i := 0; i < int(message.NumRelayCounters); i++ {
		encoding.WriteUint64(buffer, &index, message.RelayCounters[i])
	}

	return buffer[:index]
}

func (message *RelayUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["relay_id"] = int(message.RelayId)
	bigquery_message["session_count"] = int(message.SessionCount)
	bigquery_message["max_sessions"] = int(message.MaxSessions)
	bigquery_message["envelope_bandwidth_up_kbps"] = int(message.EnvelopeBandwidthUpKbps)
	bigquery_message["envelope_bandwidth_down_kbps"] = int(message.EnvelopeBandwidthDownKbps)
	bigquery_message["actual_bandwidth_up_kbps"] = int(message.ActualBandwidthUpKbps)
	bigquery_message["actual_bandwidth_down_kbps"] = int(message.ActualBandwidthDownKbps)
	bigquery_message["relay_flags"] = int(message.RelayFlags)

	relay_counters := make([]bigquery.Value, message.NumRelayCounters)
	for i := 0; i < int(message.NumRelayCounters); i++ {
		relay_counters[i] = int(message.RelayCounters[i])
	}
	bigquery_message["relay_counters"] = relay_counters

	return bigquery_message, "", nil
}
