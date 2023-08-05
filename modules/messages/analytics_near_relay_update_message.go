package messages

import (
	"fmt"
	"net"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/encoding"
)

const (
	AnalyticsNearRelayUpdateMessageVersion_Min   = 1
	AnalyticsNearRelayUpdateMessageVersion_Max   = 1
	AnalyticsNearRelayUpdateMessageVersion_Write = 1
)

type AnalyticsNearRelayUpdateMessage struct {
	Version             byte
	Timestamp           uint64
	BuyerId             uint64
	SessionId           uint64
	MatchId             uint64
	UserHash            uint64
	Latitude            float32
	Longitude           float32
	ClientAddress       net.UDPAddr
	ConnectionType      byte
	PlatformType        byte
	NumNearRelays       uint32
	NearRelayId         [constants.MaxNearRelays]uint64
	NearRelayRTT        [constants.MaxNearRelays]byte
	NearRelayJitter     [constants.MaxNearRelays]byte
	NearRelayPacketLoss [constants.MaxNearRelays]float32
}

func (message *AnalyticsNearRelayUpdateMessage) GetMaxSize() int {
	return 128 + (8+1+1+4)*constants.MaxNearRelays
}

func (message *AnalyticsNearRelayUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < AnalyticsNearRelayUpdateMessageVersion_Min || message.Version > AnalyticsNearRelayUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid analytics near relay pings message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)

	encoding.WriteUint64(buffer, &index, message.Timestamp)

	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteUint64(buffer, &index, message.MatchId)
	encoding.WriteUint64(buffer, &index, message.UserHash)
	encoding.WriteFloat32(buffer, &index, message.Latitude)
	encoding.WriteFloat32(buffer, &index, message.Longitude)
	encoding.WriteAddress(buffer, &index, &message.ClientAddress)
	encoding.WriteUint8(buffer, &index, message.ConnectionType)
	encoding.WriteUint8(buffer, &index, message.PlatformType)

	encoding.WriteUint32(buffer, &index, message.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		encoding.WriteUint64(buffer, &index, message.NearRelayId[i])
		encoding.WriteUint8(buffer, &index, message.NearRelayRTT[i])
		encoding.WriteUint8(buffer, &index, message.NearRelayJitter[i])
		encoding.WriteFloat32(buffer, &index, message.NearRelayPacketLoss[i])
	}

	return buffer[:index]
}

func (message *AnalyticsNearRelayUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read analytics near relay pings message version")
	}

	if message.Version < AnalyticsNearRelayUpdateMessageVersion_Min || message.Version > AnalyticsNearRelayUpdateMessageVersion_Max {
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

	if !encoding.ReadUint64(buffer, &index, &message.MatchId) {
		return fmt.Errorf("failed to read match id")
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

	if !encoding.ReadUint32(buffer, &index, &message.NumNearRelays) {
		return fmt.Errorf("failed to read num near relays")
	}

	for i := 0; i < int(message.NumNearRelays); i++ {

		if !encoding.ReadUint64(buffer, &index, &message.NearRelayId[i]) {
			return fmt.Errorf("failed to read near relay id")
		}

		if !encoding.ReadUint8(buffer, &index, &message.NearRelayRTT[i]) {
			return fmt.Errorf("failed to read near relay rtt")
		}

		if !encoding.ReadUint8(buffer, &index, &message.NearRelayJitter[i]) {
			return fmt.Errorf("failed to read near relay jitter")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.NearRelayPacketLoss[i]) {
			return fmt.Errorf("failed to read near relay packet loss")
		}
	}

	return nil
}

func (message *AnalyticsNearRelayUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	bigquery_entry := make(map[string]bigquery.Value)

	bigquery_entry["timestamp"] = int(message.Timestamp)
	bigquery_entry["buyer_id"] = int(message.BuyerId)
	bigquery_entry["session_id"] = int(message.SessionId)
	bigquery_entry["match_id"] = int(message.MatchId)
	bigquery_entry["user_hash"] = int(message.UserHash)
	bigquery_entry["latitude"] = float64(message.Latitude)
	bigquery_entry["longitude"] = float64(message.Longitude)
	bigquery_entry["client_address"] = message.ClientAddress.String()
	bigquery_entry["connection_type"] = int(message.ConnectionType)
	bigquery_entry["platform_type"] = int(message.PlatformType)

	near_relay_id := make([]bigquery.Value, message.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		near_relay_id[i] = int(message.NearRelayId[i])
	}
	bigquery_entry["near_relay_id"] = near_relay_id

	near_relay_rtt := make([]bigquery.Value, message.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		near_relay_rtt[i] = int(message.NearRelayRTT[i])
	}
	bigquery_entry["near_relay_rtt"] = near_relay_rtt

	near_relay_jitter := make([]bigquery.Value, message.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		near_relay_jitter[i] = int(message.NearRelayJitter[i])
	}
	bigquery_entry["near_relay_jitter"] = near_relay_jitter

	near_relay_packet_loss := make([]bigquery.Value, message.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		near_relay_packet_loss[i] = float64(message.NearRelayPacketLoss[i])
	}
	bigquery_entry["near_relay_packet_loss"] = near_relay_packet_loss

	return bigquery_entry, "", nil
}
