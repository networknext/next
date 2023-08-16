package messages

import (
	"fmt"
	"net"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/encoding"
)

const (
	PortalRelayUpdateMessageVersion_Min   = 1
	PortalRelayUpdateMessageVersion_Max   = 2
	PortalRelayUpdateMessageVersion_Write = 1
)

type PortalRelayUpdateMessage struct {
	Version                   uint8
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
	NearPingsPerSecond        float32
	RelayPingsPerSecond       float32
	RelayFlags                uint64
	NumRoutable               uint32
	NumUnroutable             uint32
	StartTime                 uint64
	CurrentTime               uint64
	RelayAddress              net.UDPAddr
	RelayVersion              string
}

func (message *PortalRelayUpdateMessage) GetMaxSize() int {
	return 512
}

func (message *PortalRelayUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to portal read relay update version")
	}

	if message.Version < PortalRelayUpdateMessageVersion_Min || message.Version > PortalRelayUpdateMessageVersion_Max {
		return fmt.Errorf("invalid portal relay update message version %d", message.Version)
	}

	if !encoding.ReadUint64(buffer, &index, &message.Timestamp) {
		return fmt.Errorf("failed to read timestamp")
	}

	if message.Version >= 2 {
		if !encoding.ReadString(buffer, &index, &message.RelayName, constants.MaxRelayNameLength) {
			return fmt.Errorf("failed to read relay name")
		}
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayId) {
		return fmt.Errorf("failed to read relay id")
	}

	if !encoding.ReadUint32(buffer, &index, &message.SessionCount) {
		return fmt.Errorf("failed to read session count")
	}

	if !encoding.ReadUint32(buffer, &index, &message.MaxSessions) {
		return fmt.Errorf("failed to read max sessions")
	}

	if !encoding.ReadUint32(buffer, &index, &message.EnvelopeBandwidthUpKbps) {
		return fmt.Errorf("failed to read envelope bandwidth up kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.EnvelopeBandwidthDownKbps) {
		return fmt.Errorf("failed to read envelope bandwidth down kbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.PacketsSentPerSecond) {
		return fmt.Errorf("failed to read packets sent per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.PacketsReceivedPerSecond) {
		return fmt.Errorf("failed to read packets received per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthSentKbps) {
		return fmt.Errorf("failed to read bandwidth sent kbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.BandwidthReceivedKbps) {
		return fmt.Errorf("failed to read bandwidth received kbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.NearPingsPerSecond) {
		return fmt.Errorf("failed to read near pings per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &message.RelayPingsPerSecond) {
		return fmt.Errorf("failed to read relay pings per-second")
	}

	if !encoding.ReadUint64(buffer, &index, &message.RelayFlags) {
		return fmt.Errorf("failed to read relay flags")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumRoutable) {
		return fmt.Errorf("failed to read num routable")
	}

	if !encoding.ReadUint32(buffer, &index, &message.NumUnroutable) {
		return fmt.Errorf("failed to read num unroutable")
	}

	if !encoding.ReadUint64(buffer, &index, &message.StartTime) {
		return fmt.Errorf("failed to read start time")
	}

	if !encoding.ReadUint64(buffer, &index, &message.CurrentTime) {
		return fmt.Errorf("failed to read current time")
	}

	if !encoding.ReadAddress(buffer, &index, &message.RelayAddress) {
		return fmt.Errorf("failed to read relay address")
	}

	if !encoding.ReadString(buffer, &index, &message.RelayVersion, constants.MaxRelayVersionLength) {
		return fmt.Errorf("failed to read relay version")
	}

	return nil
}

func (message *PortalRelayUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < PortalRelayUpdateMessageVersion_Min || message.Version > PortalRelayUpdateMessageVersion_Max {
		panic(fmt.Sprintf("invalid portal relay update message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)
	encoding.WriteUint64(buffer, &index, message.Timestamp)
	if message.Version >= 2 {
		encoding.WriteString(buffer, &index, message.RelayName, constants.MaxRelayNameLength)
	}
	encoding.WriteUint64(buffer, &index, message.RelayId)
	encoding.WriteUint32(buffer, &index, message.SessionCount)
	encoding.WriteUint32(buffer, &index, message.MaxSessions)

	encoding.WriteUint32(buffer, &index, message.EnvelopeBandwidthUpKbps)
	encoding.WriteUint32(buffer, &index, message.EnvelopeBandwidthDownKbps)
	encoding.WriteFloat32(buffer, &index, message.PacketsSentPerSecond)
	encoding.WriteFloat32(buffer, &index, message.PacketsReceivedPerSecond)
	encoding.WriteFloat32(buffer, &index, message.BandwidthSentKbps)
	encoding.WriteFloat32(buffer, &index, message.BandwidthReceivedKbps)
	encoding.WriteFloat32(buffer, &index, message.NearPingsPerSecond)
	encoding.WriteFloat32(buffer, &index, message.RelayPingsPerSecond)

	encoding.WriteUint64(buffer, &index, message.RelayFlags)
	encoding.WriteUint32(buffer, &index, message.NumRoutable)
	encoding.WriteUint32(buffer, &index, message.NumUnroutable)
	encoding.WriteUint64(buffer, &index, message.StartTime)
	encoding.WriteUint64(buffer, &index, message.CurrentTime)
	encoding.WriteAddress(buffer, &index, &message.RelayAddress)
	encoding.WriteString(buffer, &index, message.RelayVersion, constants.MaxRelayVersionLength)

	return buffer[:index]
}
