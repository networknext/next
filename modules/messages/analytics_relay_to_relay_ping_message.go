package messages

import (
	"fmt"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/accelerate/modules/encoding"
)

const (
	AnalyticsRelayToRelayPingMessageVersion_Min   = 4
	AnalyticsRelayToRelayPingMessageVersion_Max   = 4
	AnalyticsRelayToRelayPingMessageVersion_Write = 4

	MaxAnalyticsRelayToRelayPingMessageSize = 128
)

type AnalyticsRelayToRelayPingMessage struct {
	Version    uint8
	Timestamp  uint64
	RelayA     uint64
	RelayB     uint64
	RTT        uint8
	Jitter     uint8
	PacketLoss float32
}

func (message *AnalyticsRelayToRelayPingMessage) GetMaxSize() int {
	return 64
}

func (message *AnalyticsRelayToRelayPingMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics relay to relay ping message version")
	}

	if message.Version < AnalyticsRelayToRelayPingMessageVersion_Min || message.Version > AnalyticsRelayToRelayPingMessageVersion_Max {
		return fmt.Errorf("invalid analytics relay to realy ping message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayA) {
		return fmt.Errorf("failed to read relay a")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayB) {
		return fmt.Errorf("failed to read relay b")
	}

	if !encoding.ReadUint8(buffer, &index, &message.RTT) {
		return fmt.Errorf("failed to read rtt")
	}

	if !encoding.ReadUint8(buffer, &index, &message.Jitter) {
		return fmt.Errorf("failed to read jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.PacketLoss) {
		return fmt.Errorf("failed to read packet loss")
	}

	return nil
}

func (message *AnalyticsRelayToRelayPingMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsRelayToRelayPingMessageVersion_Min || message.Version > AnalyticsRelayToRelayPingMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics relay to relay ping message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.RelayA)
	encoding.WriteUint64(buffer, &index, message.RelayB)
	encoding.WriteUint8(buffer, &index, message.RTT)
	encoding.WriteUint8(buffer, &index, message.Jitter)
	encoding.WriteFloat32(buffer, &index, message.PacketLoss)

	return buffer[:index]
}

func (message *AnalyticsRelayToRelayPingMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_message := make(map[string]bigquery.Value)

	bigquery_message["timestamp"] = int(message.Timestamp)
	bigquery_message["relay_a"] = int(message.RelayA)
	bigquery_message["relay_b"] = int(message.RelayB)
	bigquery_message["rtt"] = int(message.RTT)
	bigquery_message["jitter"] = int(message.Jitter)
	bigquery_message["packet_loss"] = float64(message.PacketLoss)

	return bigquery_message, "", nil
}
