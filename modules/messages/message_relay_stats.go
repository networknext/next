package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/encoding"
)

const (
	RelayStatsMessageVersion = uint8(3)
	MaxRelayStatsMessageSize = 128
)

type RelayStatsMessage struct {
	Version byte

	Timestamp uint64

	ID uint64

	NumSessions uint32
	MaxSessions uint32

	NumRoutable   uint32
	NumUnroutable uint32

	Full bool

	CPUUsage float32

	// percent = (sent||received) / nic speed
	BandwidthSentPercent     float32
	BandwidthReceivedPercent float32

	// percent = bandwidth_(sent||received) / envelope_(sent||received)
	EnvelopeSentPercent     float32
	EnvelopeReceivedPercent float32

	BandwidthSentMbps     float32
	BandwidthReceivedMbps float32

	EnvelopeSentMbps     float32
	EnvelopeReceivedMbps float32

	// all of below are deprecated
	Tx                        uint64
	Rx                        uint64
	PeakSessions              uint64
	PeakSentBandwidthMbps     float32
	PeakReceivedBandwidthMbps float32
	MemUsage                  float32
}

func (message *RelayStatsMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read relay stat Version")
	}

	if message.Version < 2 {
		return fmt.Errorf("deprecated version")
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read relay stat Version")
	}

	if !encoding.ReadUint64(buffer, &index, &message.ID) {
		return fmt.Errorf("failed to read relay stat ID")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.CPUUsage) {
		return fmt.Errorf("failed to read relay stat CPUUsage")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.MemUsage) {
		return fmt.Errorf("failed to read relay stat MemUsage")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthSentPercent) {
		return fmt.Errorf("failed to read relay stat BandwidthSentPercent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthReceivedPercent) {
		return fmt.Errorf("failed to read relay stat BandwidthReceivedPercent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.EnvelopeSentPercent) {
		return fmt.Errorf("failed to read relay stat EnvelopeSentPercent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.EnvelopeReceivedPercent) {
		return fmt.Errorf("failed to read relay stat EnvelopeReceivedPercent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthSentMbps) {
		return fmt.Errorf("failed to read relay stat BandwidthSentMbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthReceivedMbps) {
		return fmt.Errorf("failed to read relay stat BandwidthReceivedMbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.EnvelopeSentMbps) {
		return fmt.Errorf("failed to read relay stat EnvelopeSentMbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.EnvelopeReceivedMbps) {
		return fmt.Errorf("failed to read relay stat EnvelopeReceivedMbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumSessions) {
		return fmt.Errorf("failed to read relay stat NumSessions")
	}

	if !encoding.ReadUint32(buffer, &index, &message.MaxSessions) {
		return fmt.Errorf("failed to read relay stat MaxSessions")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumRoutable) {
		return fmt.Errorf("failed to read relay stat NumRoutable")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumUnroutable) {
		return fmt.Errorf("failed to read relay stat NumUnroutable")
	}

	if message.Version >= 3 {
		if !encoding.ReadBool(buffer, &index, &message.Full) {
			return fmt.Errorf("failed to read relay stat Full")
		}
	}

	return nil
}

func (message *RelayStatsMessage) Write(buffer []byte) []byte {

	index := 0

	encoding.WriteUint8(buffer, &index, RelayStatsMessageVersion)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.ID)
	encoding.WriteFloat32(buffer, &index, message.CPUUsage)
	encoding.WriteFloat32(buffer, &index, message.MemUsage)
	encoding.WriteFloat32(buffer, &index, message.BandwidthSentPercent)
	encoding.WriteFloat32(buffer, &index, message.BandwidthReceivedPercent)
	encoding.WriteFloat32(buffer, &index, message.EnvelopeSentPercent)
	encoding.WriteFloat32(buffer, &index, message.EnvelopeReceivedPercent)
	encoding.WriteFloat32(buffer, &index, message.BandwidthSentMbps)
	encoding.WriteFloat32(buffer, &index, message.BandwidthReceivedMbps)
	encoding.WriteFloat32(buffer, &index, message.EnvelopeSentMbps)
	encoding.WriteFloat32(buffer, &index, message.EnvelopeReceivedMbps)
	encoding.WriteUint32(buffer, &index, message.NumSessions)
	encoding.WriteUint32(buffer, &index, message.MaxSessions)
	encoding.WriteUint32(buffer, &index, message.NumRoutable)
	encoding.WriteUint32(buffer, &index, message.NumUnroutable)
	encoding.WriteBool(buffer, &index, message.Full)

	return buffer[:index]
}

func (message *RelayStatsMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["relay_id"] = int(message.ID)
	bigquery_message["cpu_percent"] = message.CPUUsage
	bigquery_message["memory_percent"] = message.MemUsage
	bigquery_message["actual_bandwidth_send_percent"] = message.BandwidthSentPercent
	bigquery_message["actual_bandwidth_receive_percent"] = message.BandwidthReceivedPercent
	bigquery_message["envelope_bandwidth_send_percent"] = message.EnvelopeSentPercent
	bigquery_message["envelope_bandwidth_receive_percent"] = message.EnvelopeReceivedPercent
	bigquery_message["actual_bandwidth_send_mbps"] = message.BandwidthSentMbps
	bigquery_message["actual_bandwidth_receive_mbps"] = message.BandwidthReceivedMbps
	bigquery_message["envelope_bandwidth_send_mbps"] = message.EnvelopeSentMbps
	bigquery_message["envelope_bandwidth_receive_mbps"] = message.EnvelopeReceivedMbps
	bigquery_message["num_sessions"] = int(message.NumSessions)
	bigquery_message["max_sessions"] = int(message.MaxSessions)
	bigquery_message["num_routable"] = int(message.NumRoutable)
	bigquery_message["num_unroutable"] = int(message.NumUnroutable)

	if message.Full {
		bigquery_message["full"] = message.Full
	}

	return bigquery_message, "", nil
}
