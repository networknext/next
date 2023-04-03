package messages

import (
	"fmt"

	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/encoding"
)

const (
	PortalNearRelayUpdateMessageVersion_Min   = 1
	PortalNearRelayUpdateMessageVersion_Max   = 1
	PortalNearRelayUpdateMessageVersion_Write = 1
)

type PortalNearRelayUpdateMessage struct {
	Version             byte
	Timestamp           uint64
	BuyerId             uint64
	SessionId           uint64
	NumNearRelays       uint32
	NearRelayId         [constants.MaxNearRelays]uint64
	NearRelayRTT        [constants.MaxNearRelays]byte
	NearRelayJitter     [constants.MaxNearRelays]byte
	NearRelayPacketLoss [constants.MaxNearRelays]float32
}

func (message *PortalNearRelayUpdateMessage) GetMaxSize() int {
	return 128 + (8+1+1+4)*constants.MaxNearRelays
}

func (message *PortalNearRelayUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < PortalNearRelayUpdateMessageVersion_Min || message.Version > PortalNearRelayUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid portal near relay pings message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteUint32(buffer, &index, message.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		encoding.WriteUint64(buffer, &index, message.NearRelayId[i])
		encoding.WriteUint8(buffer, &index, message.NearRelayRTT[i])
		encoding.WriteUint8(buffer, &index, message.NearRelayJitter[i])
		encoding.WriteFloat32(buffer, &index, message.NearRelayPacketLoss[i])
	}

	return buffer[:index]
}

func (message *PortalNearRelayUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read portal near relay pings message version")
	}

	if message.Version < PortalNearRelayUpdateMessageVersion_Min || message.Version > PortalNearRelayUpdateMessageVersion_Max {
		return fmt.Errorf("invalid portal near relay pings message version %d", message.Version)
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
