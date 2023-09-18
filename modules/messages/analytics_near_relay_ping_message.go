package messages

import (
	"fmt"
	"net"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsNearRelayPingMessageVersion_Min   = 0
	AnalyticsNearRelayPingMessageVersion_Max   = 0
	AnalyticsNearRelayPingMessageVersion_Write = 0
)

type AnalyticsNearRelayPingMessage struct {
	Version             byte
	Timestamp           uint64
	BuyerId             uint64
	SessionId           uint64
	UserHash            uint64
	Latitude            float32
	Longitude           float32
	ClientAddress       net.UDPAddr
	ConnectionType      byte
	PlatformType        byte
	NearRelayId         uint64
	NearRelayRTT        byte
	NearRelayJitter     byte
	NearRelayPacketLoss float32
}

func (message *AnalyticsNearRelayPingMessage) GetMaxSize() int {
	return 128
}

func (message *AnalyticsNearRelayPingMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsNearRelayPingMessageVersion_Min || message.Version > AnalyticsNearRelayPingMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics near relay ping message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)

	encoding.WriteUint64(buffer, &index, message.Timestamp)

	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteUint64(buffer, &index, message.UserHash)
	encoding.WriteFloat32(buffer, &index, message.Latitude)
	encoding.WriteFloat32(buffer, &index, message.Longitude)
	encoding.WriteAddress(buffer, &index, &message.ClientAddress)
	encoding.WriteUint8(buffer, &index, message.ConnectionType)
	encoding.WriteUint8(buffer, &index, message.PlatformType)
	encoding.WriteUint64(buffer, &index, message.NearRelayId)
	encoding.WriteUint8(buffer, &index, message.NearRelayRTT)
	encoding.WriteUint8(buffer, &index, message.NearRelayJitter)
	encoding.WriteFloat32(buffer, &index, message.NearRelayPacketLoss)

	return buffer[:index]
}

func (message *AnalyticsNearRelayPingMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics near relay pings message version")
	}

	if message.Version < AnalyticsNearRelayPingMessageVersion_Min || message.Version > AnalyticsNearRelayPingMessageVersion_Max {
		return fmt.Errorf("invalid analytics near relay pings message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if !encoding.ReadUint64(buffer, &index, &message.BuyerId) {
		return fmt.Errorf("failed to read buyer id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionId) {
		return fmt.Errorf("failed to read session id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.UserHash) {
		return fmt.Errorf("failed to read user hash")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.Latitude) {
		return fmt.Errorf("failed to read latitude")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.Longitude) {
		return fmt.Errorf("failed to read longitude")
	}

	if !encoding.ReadAddress(buffer, &index, &message.ClientAddress) {
		return fmt.Errorf("failed to read client address")
	}

	if !encoding.ReadUint8(buffer, &index, &message.ConnectionType) {
		return fmt.Errorf("failed to read connection type")
	}

	if !encoding.ReadUint8(buffer, &index, &message.PlatformType) {
		return fmt.Errorf("failed to read platform type")
	}

	if !encoding.ReadUint64(buffer, &index, &message.NearRelayId) {
		return fmt.Errorf("failed to read near relay id")
	}

	if !encoding.ReadUint8(buffer, &index, &message.NearRelayRTT) {
		return fmt.Errorf("failed to read near relay rtt")
	}

	if !encoding.ReadUint8(buffer, &index, &message.NearRelayJitter) {
		return fmt.Errorf("failed to read near relay jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.NearRelayPacketLoss) {
		return fmt.Errorf("failed to read near relay packet loss")
	}

	return nil
}

func (message *AnalyticsNearRelayPingMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_entry := make(map[string]bigquery.Value)

	bigquery_entry["timestamp"] = int(message.Timestamp)
	bigquery_entry["buyer_id"] = int(message.BuyerId)
	bigquery_entry["session_id"] = int(message.SessionId)
	bigquery_entry["user_hash"] = int(message.UserHash)
	bigquery_entry["latitude"] = float64(message.Latitude)
	bigquery_entry["longitude"] = float64(message.Longitude)
	bigquery_entry["client_address"] = message.ClientAddress.String()
	bigquery_entry["connection_type"] = int(message.ConnectionType)
	bigquery_entry["platform_type"] = int(message.PlatformType)
	bigquery_entry["near_relay_id"] = int(message.NearRelayId)
	bigquery_entry["near_relay_rtt"] = int(message.NearRelayRTT)
	bigquery_entry["near_relay_jitter"] = int(message.NearRelayJitter)
	bigquery_entry["near_relay_packet_loss"] = float64(message.NearRelayPacketLoss)

	return bigquery_entry, "", nil
}
