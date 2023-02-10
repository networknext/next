package packets

import (
	"errors"
	"fmt"
	"net"

	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/encoding"
)

const (
	VersionNumberRelayUpdateRequest  = 1
	VersionNumberRelayUpdateResponse = 1
)

// --------------------------------------------------------------------------

type RelayPacket interface {
	Write(buffer []byte) []byte
	Read(buffer []byte) error
}

// --------------------------------------------------------------------------

type RelayUpdateRequestPacket struct {
	Version                   uint8
	Timestamp                 uint64
	Address                   net.UDPAddr
	NumSamples                uint32
	SampleRelayId             [constants.MaxRelays]uint64
	SampleRTT                 [constants.MaxRelays]uint8  // [0,255] milliseconds
	SampleJitter              [constants.MaxRelays]uint8  // [0,255] milliseconds
	SamplePacketLoss          [constants.MaxRelays]uint16 // [0,65535] -> [0%,100%]
	SessionCount              uint32
	EnvelopeBandwidthUpKbps   uint32
	EnvelopeBandwidthDownKbps uint32
	ActualBandwidthUpKbps     uint32
	ActualBandwidthDownKbps   uint32
	RelayFlags                uint64
	RelayVersion              string
	NumRelayCounters          uint32
	RelayCounters             [constants.NumRelayCounters]uint64
}

func (packet *RelayUpdateRequestPacket) Write(buffer []byte) []byte {

	index := 0

	encoding.WriteUint8(buffer, &index, packet.Version)
	encoding.WriteUint64(buffer, &index, packet.Timestamp)
	encoding.WriteAddress(buffer, &index, &packet.Address)

	encoding.WriteUint32(buffer, &index, packet.NumSamples)
	for i := 0; i < int(packet.NumSamples); i++ {
		encoding.WriteUint64(buffer, &index, packet.SampleRelayId[i])
		encoding.WriteUint8(buffer, &index, packet.SampleRTT[i])
		encoding.WriteUint8(buffer, &index, packet.SampleJitter[i])
		encoding.WriteUint16(buffer, &index, packet.SamplePacketLoss[i])
	}

	encoding.WriteUint32(buffer, &index, packet.SessionCount)
	encoding.WriteUint32(buffer, &index, packet.EnvelopeBandwidthUpKbps)
	encoding.WriteUint32(buffer, &index, packet.EnvelopeBandwidthDownKbps)
	encoding.WriteUint32(buffer, &index, packet.ActualBandwidthUpKbps)
	encoding.WriteUint32(buffer, &index, packet.ActualBandwidthDownKbps)

	encoding.WriteUint64(buffer, &index, packet.RelayFlags)
	encoding.WriteString(buffer, &index, packet.RelayVersion, constants.MaxRelayVersionLength)

	encoding.WriteUint32(buffer, &index, packet.NumRelayCounters)
	for i := 0; i < int(packet.NumRelayCounters); i++ {
		encoding.WriteUint64(buffer, &index, packet.RelayCounters[i])
	}

	return buffer[:index]
}

func (packet *RelayUpdateRequestPacket) Read(buffer []byte) error {

	index := 0

	encoding.ReadUint8(buffer, &index, &packet.Version)

	if packet.Version != VersionNumberRelayUpdateRequest {
		return errors.New("invalid relay update request packet version")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.Timestamp) {
		return errors.New("could not read timestamp")
	}

	if !encoding.ReadAddress(buffer, &index, &packet.Address) {
		return errors.New("could not read relay address")
	}

	if !encoding.ReadUint32(buffer, &index, &packet.NumSamples) {
		return errors.New("could not read num samples")
	}

	if packet.NumSamples < 0 || packet.NumSamples > constants.MaxRelays {
		return errors.New("invalid num samples")
	}

	for i := 0; i < int(packet.NumSamples); i++ {

		if !encoding.ReadUint64(buffer, &index, &packet.SampleRelayId[i]) {
			return errors.New("could not read sample relay id")
		}

		if !encoding.ReadUint8(buffer, &index, &packet.SampleRTT[i]) {
			return errors.New("could not read sample rtt")
		}

		if !encoding.ReadUint8(buffer, &index, &packet.SampleJitter[i]) {
			return errors.New("could not read sample jitter")
		}

		if !encoding.ReadUint16(buffer, &index, &packet.SamplePacketLoss[i]) {
			return errors.New("could not read sample packet loss")
		}
	}

	if !encoding.ReadUint32(buffer, &index, &packet.SessionCount) {
		return errors.New("could not read session count")
	}

	if !encoding.ReadUint32(buffer, &index, &packet.EnvelopeBandwidthUpKbps) {
		return errors.New("could not read envelope bandwidth up kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &packet.EnvelopeBandwidthDownKbps) {
		return errors.New("could not read envelope bandwidth down kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &packet.ActualBandwidthUpKbps) {
		return errors.New("could not read actual bandwidth up kbps")
	}

	if !encoding.ReadUint32(buffer, &index, &packet.ActualBandwidthDownKbps) {
		return errors.New("could not read actual bandwidth down kbps")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.RelayFlags) {
		return errors.New("could not read relay flags")
	}

	if !encoding.ReadString(buffer, &index, &packet.RelayVersion, constants.MaxRelayVersionLength) {
		return errors.New("could not read relay version string")
	}

	if !encoding.ReadUint32(buffer, &index, &packet.NumRelayCounters) {
		return errors.New("could not read num relay counters")
	}

	if packet.NumRelayCounters != constants.NumRelayCounters {
		return fmt.Errorf("wrong number of relay counters. expected %d, got %d", constants.NumRelayCounters, packet.NumRelayCounters)
	}

	for i := 0; i < int(packet.NumRelayCounters); i++ {
		if !encoding.ReadUint64(buffer, &index, &packet.RelayCounters[i]) {
			return errors.New("could not read relay counter")
		}
	}

	return nil
}

func (packet *RelayUpdateRequestPacket) Peek(buffer []byte) error {

	index := 0

	encoding.ReadUint8(buffer, &index, &packet.Version)

	if packet.Version != VersionNumberRelayUpdateRequest {
		return errors.New("invalid relay update request packet version")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.Timestamp) {
		return errors.New("could not read timestamp")
	}

	if !encoding.ReadAddress(buffer, &index, &packet.Address) {
		return errors.New("could not read relay address")
	}

	return nil
}

// --------------------------------------------------------------------------

type RelayUpdateResponsePacket struct {
	Version       uint8
	Timestamp     uint64
	NumRelays     uint32
	RelayId       [constants.MaxRelays]uint64
	RelayAddress  [constants.MaxRelays]net.UDPAddr
	RelayInternal [constants.MaxRelays]byte
	TargetVersion string
	UpcomingMagic []byte
	CurrentMagic  []byte
	PreviousMagic []byte
}

func (packet *RelayUpdateResponsePacket) Write(buffer []byte) []byte {

	index := 0

	encoding.WriteUint8(buffer, &index, packet.Version)
	encoding.WriteUint64(buffer, &index, uint64(packet.Timestamp))
	encoding.WriteUint32(buffer, &index, uint32(packet.NumRelays))

	for i := 0; i < int(packet.NumRelays); i++ {
		encoding.WriteUint64(buffer, &index, packet.RelayId[i])
		encoding.WriteAddress(buffer, &index, &packet.RelayAddress[i])
		encoding.WriteUint8(buffer, &index, packet.RelayInternal[i])
	}

	encoding.WriteString(buffer, &index, packet.TargetVersion, constants.MaxRelayVersionLength)

	encoding.WriteBytes(buffer, &index, packet.UpcomingMagic, 8)
	encoding.WriteBytes(buffer, &index, packet.CurrentMagic, 8)
	encoding.WriteBytes(buffer, &index, packet.PreviousMagic, 8)

	return buffer[:index]
}

func (packet *RelayUpdateResponsePacket) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &packet.Version) {
		return errors.New("could not read version")
	}

	if packet.Version > VersionNumberRelayUpdateResponse {
		return errors.New("invalid relay update response version")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.Timestamp) {
		return errors.New("could not read timestamp")
	}

	if !encoding.ReadUint32(buffer, &index, &packet.NumRelays) {
		return errors.New("could not read num relays")
	}

	if packet.NumRelays < 0 || packet.NumRelays > constants.MaxRelays {
		return errors.New("invalid num relays")
	}

	for i := 0; i < int(packet.NumRelays); i++ {

		if !encoding.ReadUint64(buffer, &index, &packet.RelayId[i]) {
			return errors.New("could not read relay id")
		}

		if !encoding.ReadAddress(buffer, &index, &packet.RelayAddress[i]) {
			return errors.New("could not read relay address")
		}

		if !encoding.ReadUint8(buffer, &index, &packet.RelayInternal[i]) {
			return errors.New("could not read relay internal")
		}
	}

	if !encoding.ReadString(buffer, &index, &packet.TargetVersion, constants.MaxRelayVersionLength) {
		return errors.New("could not read target version")
	}

	packet.UpcomingMagic = make([]byte, 8)
	packet.CurrentMagic = make([]byte, 8)
	packet.PreviousMagic = make([]byte, 8)

	if !encoding.ReadBytes(buffer, &index, &packet.UpcomingMagic, 8) {
		return errors.New("could not read upcoming magic")
	}

	if !encoding.ReadBytes(buffer, &index, &packet.CurrentMagic, 8) {
		return errors.New("could not read current magic")
	}

	if !encoding.ReadBytes(buffer, &index, &packet.PreviousMagic, 8) {
		return errors.New("could not read previous magic")
	}

	return nil
}
