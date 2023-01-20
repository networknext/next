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

	MaxNearRelays = 32
	MaxRouteRelays = 5
)

type PortalMessage struct {
	Version          byte

	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte

	SessionId        uint64
	BuyerId          uint64
	DatacenterId     uint64
	Latitude         float32
	Longitude        float32
	ClientAddress    net.UDPAddr
	ServerAddress    net.UDPAddr

	SliceNumber      	      uint32
	
	DirectRTT                 float32
	DirectJitter              float32
	DirectPacketLoss          float32
	DirectKbpsUp              uint32
	DirectKbpsDown            uint32
	
	Next                      bool
	NextRTT                   float32
	NextJitter                float32
	NextPacketLoss            float32
	NextKbpsUp                uint32
	NextKbpsDown              uint32
	NextBandwidthOverLimit    bool
	NextPredictedRTT          uint32
	NextNumRouteRelays        uint32
	NextRouteRelayId          [MaxRouteRelays]uint64

	RealJitter                float32
	RealPacketLoss            float32
	RealOutOfOrder            float32

	Reported         bool
	FallbackToDirect bool

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

	encoding.WriteFloat32(buffer, &index, message.DirectRTT)
	encoding.WriteFloat32(buffer, &index, message.DirectJitter)
	encoding.WriteFloat32(buffer, &index, message.DirectPacketLoss)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsUp)
	encoding.WriteUint32(buffer, &index, message.DirectKbpsDown)

	encoding.WriteBool(buffer, &index, message.Next)
	if message.Next {
		encoding.WriteFloat32(buffer, &index, message.NextRTT)
		encoding.WriteFloat32(buffer, &index, message.NextJitter)
		encoding.WriteFloat32(buffer, &index, message.NextPacketLoss)
		encoding.WriteUint32(buffer, &index, message.NextKbpsUp)
		encoding.WriteUint32(buffer, &index, message.NextKbpsDown)
		encoding.WriteBool(buffer, &index, message.NextBandwidthOverLimit)
		encoding.WriteUint32(buffer, &index, message.NextPredictedRTT)
		encoding.WriteUint32(buffer, &index, message.NextNumRouteRelays)
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			encoding.WriteUint64(buffer, &index, message.NextRouteRelayId[i])
		}
	}

	encoding.WriteFloat32(buffer, &index, message.RealJitter)
	encoding.WriteFloat32(buffer, &index, message.RealPacketLoss)
	encoding.WriteFloat32(buffer, &index, message.RealOutOfOrder)

	encoding.WriteBool(buffer, &index, message.Reported)
	encoding.WriteBool(buffer, &index, message.FallbackToDirect)

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
		return fmt.Errorf("invalid server portal message version %d", message.Version)
	}

	// todo

	return nil
}
