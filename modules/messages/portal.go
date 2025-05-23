package messages

import (
	"github.com/networknext/next/modules/constants"
	"net"
)

// ----------------------------------------------------------------------------------------

type PortalClientRelayUpdateMessage struct {
	Timestamp             uint64
	BuyerId               uint64
	SessionId             uint64
	NumClientRelays       uint32
	ClientRelayId         [constants.MaxClientRelays]uint64
	ClientRelayRTT        [constants.MaxClientRelays]byte
	ClientRelayJitter     [constants.MaxClientRelays]byte
	ClientRelayPacketLoss [constants.MaxClientRelays]float32
}

// ----------------------------------------------------------------------------------------

type PortalServerRelayUpdateMessage struct {
	Timestamp             uint64
	BuyerId               uint64
	SessionId             uint64
	NumServerRelays       uint32
	ServerRelayId         [constants.MaxServerRelays]uint64
	ServerRelayRTT        [constants.MaxServerRelays]byte
	ServerRelayJitter     [constants.MaxServerRelays]byte
	ServerRelayPacketLoss [constants.MaxServerRelays]float32
}

// ----------------------------------------------------------------------------------------

type PortalRelayUpdateMessage struct {
	Timestamp                 uint64
	RelayName                 string
	RelayId                   uint64
	SessionCount              uint32
	MaxSessions               uint32
	EnvelopeBandwidthUpKbps   uint32
	EnvelopeBandwidthDownKbps uint32
	PacketsSentPerSecond      float32
	PacketsReceivedPerSecond  float32
	BandwidthSentKbps         float32
	BandwidthReceivedKbps     float32
	ClientPingsPerSecond      float32
	ServerPingsPerSecond      float32
	RelayPingsPerSecond       float32
	RelayFlags                uint64
	NumRoutable               uint32
	NumUnroutable             uint32
	StartTime                 uint64
	CurrentTime               uint64
	RelayAddress              net.UDPAddr
	RelayVersion              string
}

// ----------------------------------------------------------------------------------------

type PortalServerUpdateMessage struct {
	Timestamp        uint64
	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte
	BuyerId          uint64
	DatacenterId     uint64
	NumSessions      uint32
	Uptime           uint64
	ServerAddress    net.UDPAddr
}

// ----------------------------------------------------------------------------------------

type PortalSessionUpdateMessage struct {
	Timestamp uint64

	SDKVersion_Major byte
	SDKVersion_Minor byte
	SDKVersion_Patch byte

	SessionId      uint64
	UserHash       uint64
	StartTime      uint64
	BuyerId        uint64
	DatacenterId   uint64
	Latitude       float32
	Longitude      float32
	ClientAddress  net.UDPAddr
	ServerAddress  net.UDPAddr
	SliceNumber    uint32
	SessionFlags   uint64
	SessionEvents  uint64
	InternalEvents uint64
	ConnectionType uint8
	PlatformType   uint8

	DirectRTT        float32
	DirectJitter     float32
	DirectPacketLoss float32
	DirectKbpsUp     uint32
	DirectKbpsDown   uint32

	Next               bool
	NextRTT            float32
	NextJitter         float32
	NextPacketLoss     float32
	NextKbpsUp         uint32
	NextKbpsDown       uint32
	NextPredictedRTT   uint32
	NextNumRouteRelays uint32
	NextRouteRelayId   [constants.MaxRouteRelays]uint64

	RealJitter     float32
	RealPacketLoss float32
	RealOutOfOrder float32

	NumClientRelays       uint32
	ClientRelayId         [constants.MaxClientRelays]uint64
	ClientRelayRTT        [constants.MaxClientRelays]byte
	ClientRelayJitter     [constants.MaxClientRelays]byte
	ClientRelayPacketLoss [constants.MaxClientRelays]float32
	ClientRelayRoutable   [constants.MaxClientRelays]bool

	NumServerRelays       uint32
	ServerRelayId         [constants.MaxServerRelays]uint64
	ServerRelayRTT        [constants.MaxServerRelays]byte
	ServerRelayJitter     [constants.MaxServerRelays]byte
	ServerRelayPacketLoss [constants.MaxServerRelays]float32
	ServerRelayRoutable   [constants.MaxServerRelays]bool

	BestScore     uint32
	BestDirectRTT uint32
	BestNextRTT   uint32

	Retry            bool
	FallbackToDirect bool
	SendToPortal     bool
}

// ----------------------------------------------------------------------------------------
