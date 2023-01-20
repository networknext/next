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
	DirectBandwidthUpKbps     float32
	DirectBandwidthUpDownKbps float32
	Next                      bool
	NextRTT                   float32
	NextJitter                float32
	NextPacketLoss            float32
	NextBandwidthUpKbps       float32
	NextBandwidthDownKbps     float32
	RealJitter                float32
	RealPacketLoss            float32
	RealOutOfOrder            float32
	PredictedRTT              uint32

	Reported         bool
	FallbackToDirect bool

	NumRouteRelays    int
	RouteRelayId      [MaxRouteRelays]uint64
	RouteRelayAddress [MaxRouteRelays]net.UDPAddr

	NumNearRelays       int
	NearRelayId         [MaxNearRelays]uint64
	NearRelayRTT        [MaxNearRelays]byte
	NearRelayJitter     [MaxNearRelays]byte
	NearRelayPacketLoss [MaxNearRelays]float32
	NearRelayRoutable   [MaxNearRelays]bool
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

func (message *PortalMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < PortalMessageVersion_Min || message.Version > PortalMessageVersion_Max {
		panic(fmt.Sprintf("invalid portal message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)

	// todo

	return buffer[:index]
}
