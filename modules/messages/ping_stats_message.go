package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/encoding"
)

const (
	PingStatsMessageVersion_Min   = 4
	PingStatsMessageVersion_Max   = 4
	PingStatsMessageVersion_Write = 4

	MaxPingStatsMessageSize = 128
)

type PingStatsMessage struct {
	Version    uint8
	Timestamp  uint64
	RelayA     uint64
	RelayB     uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
	Routable   bool
}

func (message *PingStatsMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read ping stats version")
	}

	if message.Version < PingStatsMessageVersion_Min || message.Version > PingStatsMessageVersion_Max {
		return fmt.Errorf("invalid ping stats message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read ping stats timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayA) {
		return fmt.Errorf("failed to read ping stats relay a")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayB) {
		return fmt.Errorf("failed to read ping stats relay b")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RTT) {
		return fmt.Errorf("failed to read ping stats rtt")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.Jitter) {
		return fmt.Errorf("failed to read ping stats jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.PacketLoss) {
		return fmt.Errorf("failed to read ping stats packet loss")
	}

	if !encoding.ReadBool(buffer, &index, &message.Routable) {
		return fmt.Errorf("failed to read ping stats routable")
	}

	return nil
}

func (message *PingStatsMessage) Write(buffer []byte) []byte {

	index := 0

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.RelayA)
	encoding.WriteUint64(buffer, &index, message.RelayB)
	encoding.WriteFloat32(buffer, &index, message.RTT)
	encoding.WriteFloat32(buffer, &index, message.Jitter)
	encoding.WriteFloat32(buffer, &index, message.PacketLoss)
	encoding.WriteBool(buffer, &index, message.Routable)

	return buffer[:index]
}

func (message *PingStatsMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["relay_a"] = int(message.RelayA)
	bigquery_message["relay_b"] = int(message.RelayB)
	bigquery_message["rtt"] = message.RTT
	bigquery_message["jitter"] = message.Jitter
	bigquery_message["packet_loss"] = message.PacketLoss
	bigquery_message["routable"] = message.Routable

	return bigquery_message, "", nil
}
