package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/encoding"
)

const (
	RelayStatsMessageVersion_Min   = 3
	RelayStatsMessageVersion_Max   = 3
	RelayStatsMessageVersion_Write = 3

	MaxRelayStatsMessageSize = 128
)

type RelayStatsMessage struct {
	Version       uint8
	Timestamp     uint64

	// todo: update all this to the latest stats we really have from each relay
	
	ID            uint64
	NumSessions   uint32
	MaxSessions   uint32
	NumRoutable   uint32
	NumUnroutable uint32
	Full          bool
	CPUUsage      float32

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
		return fmt.Errorf("failed to read relay stat version")
	}

	if message.Version < RelayStatsMessageVersion_Min || message.Version > RelayStatsMessageVersion_Max {
		return fmt.Errorf("invalid relay stats message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read relay stat timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.ID) {
		return fmt.Errorf("failed to read relay stat id")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.CPUUsage) {
		return fmt.Errorf("failed to read relay stat cpu usage")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.MemUsage) {
		return fmt.Errorf("failed to read relay stat mem usage")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthSentPercent) {
		return fmt.Errorf("failed to read relay stat bandwidth sent percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthReceivedPercent) {
		return fmt.Errorf("failed to read relay stat bandwidth received percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.EnvelopeSentPercent) {
		return fmt.Errorf("failed to read relay stat envelope sent percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.EnvelopeReceivedPercent) {
		return fmt.Errorf("failed to read relay stat envelope received percent")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthSentMbps) {
		return fmt.Errorf("failed to read relay stat bandwidth sent mbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthReceivedMbps) {
		return fmt.Errorf("failed to read relay stat bandwidth received mbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.EnvelopeSentMbps) {
		return fmt.Errorf("failed to read relay stat envelope sent mbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.EnvelopeReceivedMbps) {
		return fmt.Errorf("failed to read relay stat envelope received mbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumSessions) {
		return fmt.Errorf("failed to read relay stat num sessions")
	}

	if !encoding.ReadUint32(buffer, &index, &message.MaxSessions) {
		return fmt.Errorf("failed to read relay stat max sessions")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumRoutable) {
		return fmt.Errorf("failed to read relay stat num routable")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumUnroutable) {
		return fmt.Errorf("failed to read relay stat num unroutable")
	}

	if message.Version >= 3 {
		if !encoding.ReadBool(buffer, &index, &message.Full) {
			return fmt.Errorf("failed to read relay stat full")
		}
	}

	return nil
}

func (message *RelayStatsMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < RelayStatsMessageVersion_Min || message.Version > RelayStatsMessageVersion_Max {
		panic(fmt.Sprintf("invalid relay stats message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
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
