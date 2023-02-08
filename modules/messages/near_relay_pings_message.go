package messages

import (
	"fmt"
	"net"

	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/encoding"
)

const (
	NearRelayPingsMessageVersion_Min   = 1
	NearRelayPingsMessageVersion_Max   = 1
	NearRelayPingsMessageVersion_Write = 1
)

type NearRelayPingsMessage struct {
	Version byte

	Timestamp uint64

	BuyerId        uint64
	SessionId      uint64
	UserHash       uint64
	Latitude       float32
	Longitude      float32
	ClientAddress  net.UDPAddr
	ConnectionType byte
	PlatformType   byte

	NumNearRelays       uint32
	NearRelayId         [constants.MaxNearRelays]uint64
	NearRelayRTT        [constants.MaxNearRelays]byte
	NearRelayJitter     [constants.MaxNearRelays]byte
	NearRelayPacketLoss [constants.MaxNearRelays]float32
}

func (message *NearRelayPingsMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < NearRelayPingsMessageVersion_Min || message.Version > NearRelayPingsMessageVersion_Max {
		panic(fmt.Sprintf("invalid near relay pings message version %d", message.Version))
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

	encoding.WriteUint32(buffer, &index, message.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		encoding.WriteUint64(buffer, &index, message.NearRelayId[i])
		encoding.WriteUint8(buffer, &index, message.NearRelayRTT[i])
		encoding.WriteUint8(buffer, &index, message.NearRelayJitter[i])
		encoding.WriteFloat32(buffer, &index, message.NearRelayPacketLoss[i])
	}

	return buffer[:index]
}

func (message *NearRelayPingsMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read near relay pings message version")
	}

	if message.Version < NearRelayPingsMessageVersion_Min || message.Version > NearRelayPingsMessageVersion_Max {
		return fmt.Errorf("invalid near relay pings message version %d", message.Version)
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
