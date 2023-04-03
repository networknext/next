package messages

import (
	"fmt"
	"net"

	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/encoding"
)

const (
	PortalRelayUpdateMessageVersion_Min   = 1
	PortalRelayUpdateMessageVersion_Max   = 1
	PortalRelayUpdateMessageVersion_Write = 1
)

type PortalRelayUpdateMessage struct {
	Version                   uint8
	Timestamp                 uint64
	RelayId                   uint64
	SessionCount              uint32
	MaxSessions               uint32
	EnvelopeBandwidthUpKbps   uint32
	EnvelopeBandwidthDownKbps uint32
	ActualBandwidthUpKbps     uint32
	ActualBandwidthDownKbps   uint32
	RelayFlags                uint64
	NumRoutable               uint32
	NumUnroutable             uint32
	StartTime                 uint64
	RelayAddress              net.UDPAddr
	RelayVersion              string
}

func (message *PortalRelayUpdateMessage) GetMaxSize() int {
	return 256
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

	if !encoding.ReadUint32(buffer, &index, &message.ActualBandwidthUpKbps) {
		return fmt.Errorf("failed to read actual bandwidth up kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &message.ActualBandwidthDownKbps) {
		return fmt.Errorf("failed to read actual bandwidth down kbps")
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
	encoding.WriteUint64(buffer, &index, message.RelayId)
	encoding.WriteUint32(buffer, &index, message.SessionCount)
	encoding.WriteUint32(buffer, &index, message.MaxSessions)
	encoding.WriteUint32(buffer, &index, message.EnvelopeBandwidthUpKbps)
	encoding.WriteUint32(buffer, &index, message.EnvelopeBandwidthDownKbps)
	encoding.WriteUint32(buffer, &index, message.ActualBandwidthUpKbps)
	encoding.WriteUint32(buffer, &index, message.ActualBandwidthDownKbps)
	encoding.WriteUint64(buffer, &index, message.RelayFlags)
	encoding.WriteUint32(buffer, &index, message.NumRoutable)
	encoding.WriteUint32(buffer, &index, message.NumUnroutable)
	encoding.WriteUint64(buffer, &index, message.StartTime)
	encoding.WriteAddress(buffer, &index, &message.RelayAddress)
	encoding.WriteString(buffer, &index, message.RelayVersion, constants.MaxRelayVersionLength)

	return buffer[:index]
}
