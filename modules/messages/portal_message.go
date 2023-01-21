package messages

import (
	"fmt"
	"net"

	"github.com/networknext/backend/modules/encoding"
)

const (
	PortalMessageVersion_Min   = 1
	PortalMessageVersion_Max   = 1
	PortalMessageVersion_Write = 1

	// todo: constants module
	MaxNearRelays  = 32
	MaxRouteRelays = 5
)

type PortalMessage struct {
	Version byte

	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte

	SessionId     uint64
	BuyerId       uint64
	DatacenterId  uint64
	Latitude      float32
	Longitude     float32
	ClientAddress net.UDPAddr
	ServerAddress net.UDPAddr
	SliceNumber   uint32
	SessionFlags  uint64
	GameEvents    uint64

	DirectRTT        float32
	DirectJitter     float32
	DirectPacketLoss float32
	DirectKbpsUp     uint32
	DirectKbpsDown   uint32

	NextRTT            float32
	NextJitter         float32
	NextPacketLoss     float32
	NextKbpsUp         uint32
	NextKbpsDown       uint32
	NextPredictedRTT   uint32
	NextNumRouteRelays uint32
	NextRouteRelayId   [MaxRouteRelays]uint64

	RealJitter     float32
	RealPacketLoss float32
	RealOutOfOrder float32

	NumNearRelays       uint32
	NearRelayId         [MaxNearRelays]uint64
	NearRelayRTT        [MaxNearRelays]byte
	NearRelayJitter     [MaxNearRelays]byte
	NearRelayPacketLoss [MaxNearRelays]float32
	NearRelayRoutable   [MaxNearRelays]bool
}

func (message *PortalMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < PortalMessageVersion_Min || message.Version > PortalMessageVersion_Max {
		panic(fmt.Sprintf("invalid portal message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)

	encoding.WriteUint8(buffer, &index, message.SDKVersion_Major)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Minor)
	encoding.WriteUint8(buffer, &index, message.SDKVersion_Patch)

	encoding.WriteUint64(buffer, &index, message.SessionId)
	encoding.WriteUint64(buffer, &index, message.BuyerId)
	encoding.WriteUint64(buffer, &index, message.DatacenterId)
	encoding.WriteFloat32(buffer, &index, message.Latitude)
	encoding.WriteFloat32(buffer, &index, message.Longitude)
	encoding.WriteAddress(buffer, &index, &message.ClientAddress)
	encoding.WriteAddress(buffer, &index, &message.ServerAddress)
	encoding.WriteUint32(buffer, &index, message.SliceNumber)
	encoding.WriteUint64(buffer, &index, message.SessionFlags)
	encoding.WriteUint64(buffer, &index, message.GameEvents)

	encoding.WriteFloat32(buffer, &index, message.DirectRTT)
	encoding.WriteFloat32(buffer, &index, message.DirectJitter)
	encoding.WriteFloat32(buffer, &index, message.DirectPacketLoss)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsUp)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsDown)

	if (message.SessionFlags & SessionFlags_Next) != 0 {
		encoding.WriteFloat32(buffer, &index, message.NextRTT)
		encoding.WriteFloat32(buffer, &index, message.NextJitter)
		encoding.WriteFloat32(buffer, &index, message.NextPacketLoss)
		encoding.WriteUint32(buffer, &index, message.NextKbpsUp)
		encoding.WriteUint32(buffer, &index, message.NextKbpsDown)
		encoding.WriteUint32(buffer, &index, message.NextPredictedRTT)
		encoding.WriteUint32(buffer, &index, message.NextNumRouteRelays)
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			encoding.WriteUint64(buffer, &index, message.NextRouteRelayId[i])
		}
	}

	encoding.WriteFloat32(buffer, &index, message.RealJitter)
	encoding.WriteFloat32(buffer, &index, message.RealPacketLoss)
	encoding.WriteFloat32(buffer, &index, message.RealOutOfOrder)

	encoding.WriteUint32(buffer, &index, message.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		encoding.WriteUint64(buffer, &index, message.NearRelayId[i])
		encoding.WriteUint8(buffer, &index, message.NearRelayRTT[i])
		encoding.WriteUint8(buffer, &index, message.NearRelayJitter[i])
		encoding.WriteFloat32(buffer, &index, message.NearRelayPacketLoss[i])
		encoding.WriteBool(buffer, &index, message.NearRelayRoutable[i])
	}

	return buffer[:index]
}

func (message *PortalMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read portal message version")
	}

	if message.Version < PortalMessageVersion_Min || message.Version > PortalMessageVersion_Max {
		return fmt.Errorf("invalid portal message version %d", message.Version)
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Major) {
		return fmt.Errorf("failed to read sdk version major")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Minor) {
		return fmt.Errorf("failed to read sdk version minor")
	}

	if !encoding.ReadUint8(buffer, &index, &message.SDKVersion_Patch) {
		return fmt.Errorf("failed to read sdk version patch")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionId) {
		return fmt.Errorf("failed to read session id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.BuyerId) {
		return fmt.Errorf("failed to read buyer id")
	}

	if !encoding.ReadUint64(buffer, &index, &message.DatacenterId) {
		return fmt.Errorf("failed to read datacenter id")
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

	if !encoding.ReadAddress(buffer, &index, &message.ServerAddress) {
		return fmt.Errorf("failed to read server address")
	}

	if !encoding.ReadUint32(buffer, &index, &message.SliceNumber) {
		return fmt.Errorf("failed to read slice number")
	}

	if !encoding.ReadUint64(buffer, &index, &message.SessionFlags) {
		return fmt.Errorf("failed to read session flags")
	}

	if !encoding.ReadUint64(buffer, &index, &message.GameEvents) {
		return fmt.Errorf("failed to read game events")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectRTT) {
		return fmt.Errorf("failed to read direct rtt")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectJitter) {
		return fmt.Errorf("failed to read direct jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.DirectPacketLoss) {
		return fmt.Errorf("failed to read direct packet loss")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DirectKbpsUp) {
		return fmt.Errorf("failed to read direct kbps up")
	}

	if !encoding.ReadUint32(buffer, &index, &message.DirectKbpsDown) {
		return fmt.Errorf("failed to read direct kbps down")
	}

	if (message.SessionFlags & SessionFlags_Next) != 0 {

		if !encoding.ReadFloat32(buffer, &index, &message.NextRTT) {
			return fmt.Errorf("failed to read next rtt")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.NextJitter) {
			return fmt.Errorf("failed to read next jitter")
		}

		if !encoding.ReadFloat32(buffer, &index, &message.NextPacketLoss) {
			return fmt.Errorf("failed to read next packet loss")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextKbpsUp) {
			return fmt.Errorf("failed to read next kbps up")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextKbpsDown) {
			return fmt.Errorf("failed to read next kbps down")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextPredictedRTT) {
			return fmt.Errorf("failed to read next predicted rtt")
		}

		if !encoding.ReadUint32(buffer, &index, &message.NextNumRouteRelays) {
			return fmt.Errorf("failed to read next num route relays")
		}

		for i := 0; i < int(message.NextNumRouteRelays); i++ {

			if !encoding.ReadUint64(buffer, &index, &message.NextRouteRelayId[i]) {
				return fmt.Errorf("failed to read next route relay id")
			}
		}
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealJitter) {
		return fmt.Errorf("failed to read real jitter")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealPacketLoss) {
		return fmt.Errorf("failed to read real packet loss")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RealOutOfOrder) {
		return fmt.Errorf("failed to read real out of order")
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

		if !encoding.ReadBool(buffer, &index, &message.NearRelayRoutable[i]) {
			return fmt.Errorf("failed to read near relay packet routable")
		}
	}

	return nil
}
