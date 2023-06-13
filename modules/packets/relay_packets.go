package packets

import (
	"errors"
	"fmt"
	"net"

	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/crypto"
	"github.com/networknext/accelerate/modules/encoding"
)

const (
	RelayUpdateRequestPacket_VersionMin   = 1
	RelayUpdateRequestPacket_VersionMax   = 1
	RelayUpdateRequestPacket_VersionWrite = 1

	RelayUpdateResponsePacket_VersionMin   = 1
	RelayUpdateResponsePacket_VersionMax   = 1
	RelayUpdateResponsePacket_VersionWrite = 1
)

// --------------------------------------------------------------------------

type RelayPacket interface {
	Write(buffer []byte) []byte
	Read(buffer []byte) error
}

// --------------------------------------------------------------------------

type RelayUpdateRequestPacket struct {
	Version                   uint8
	Address                   net.UDPAddr
	StartTime                 uint64
	CurrentTime               uint64
	NumSamples                uint32
	SampleRelayId             [constants.MaxRelays]uint64
	SampleRTT                 [constants.MaxRelays]uint8  // [0,255] milliseconds
	SampleJitter              [constants.MaxRelays]uint8  // [0,255] milliseconds
	SamplePacketLoss          [constants.MaxRelays]uint16 // [0,65535] -> [0%,100%]
	SessionCount              uint32
	EnvelopeBandwidthUpKbps   uint32
	EnvelopeBandwidthDownKbps uint32
	PacketsSentPerSecond      float32
	PacketsReceivedPerSecond  float32
	BandwidthSentKbps         float32
	BandwidthReceivedKbps     float32
	NearPingsPerSecond        float32
	RelayPingsPerSecond       float32
	RelayFlags                uint64
	RelayVersion              string
	NumRelayCounters          uint32
	RelayCounters             [constants.NumRelayCounters]uint64
}

func (packet *RelayUpdateRequestPacket) Write(buffer []byte) []byte {

	index := 0

	if packet.Version < RelayUpdateRequestPacket_VersionMin || packet.Version > RelayUpdateRequestPacket_VersionMax {
		panic(fmt.Sprintf("invalid relay update request packet version %d", packet.Version))
	}

	encoding.WriteUint8(buffer, &index, packet.Version)
	encoding.WriteAddress(buffer, &index, &packet.Address)
	encoding.WriteUint64(buffer, &index, packet.CurrentTime)
	encoding.WriteUint64(buffer, &index, packet.StartTime)

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
	encoding.WriteFloat32(buffer, &index, packet.PacketsSentPerSecond)
	encoding.WriteFloat32(buffer, &index, packet.PacketsReceivedPerSecond)
	encoding.WriteFloat32(buffer, &index, packet.BandwidthSentKbps)
	encoding.WriteFloat32(buffer, &index, packet.BandwidthReceivedKbps)
	encoding.WriteFloat32(buffer, &index, packet.NearPingsPerSecond)
	encoding.WriteFloat32(buffer, &index, packet.RelayPingsPerSecond)

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

	if packet.Version < RelayUpdateRequestPacket_VersionMin && packet.Version > RelayUpdateRequestPacket_VersionMax {
		return errors.New("invalid relay update request packet version")
	}

	if !encoding.ReadAddress(buffer, &index, &packet.Address) {
		return errors.New("could not read relay address")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.CurrentTime) {
		return errors.New("could not read current time")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.StartTime) {
		return errors.New("could not read start time")
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

	if !encoding.ReadFloat32(buffer, &index, &packet.PacketsSentPerSecond) {
		return errors.New("could not read packets sent per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &packet.PacketsReceivedPerSecond) {
		return errors.New("could not read packets received per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &packet.BandwidthSentKbps) {
		return errors.New("could not read bandwidth sent kbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &packet.BandwidthReceivedKbps) {
		return errors.New("could not read bandwidth received kbps")
	}

	if !encoding.ReadFloat32(buffer, &index, &packet.NearPingsPerSecond) {
		return errors.New("could not read near pings per-second")
	}

	if !encoding.ReadFloat32(buffer, &index, &packet.RelayPingsPerSecond) {
		return errors.New("could not read relay pings per-second")
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

// --------------------------------------------------------------------------

type RelayUpdateResponsePacket struct {
	Version                       uint8
	Timestamp                     uint64
	NumRelays                     uint32
	RelayId                       [constants.MaxRelays]uint64
	RelayAddress                  [constants.MaxRelays]net.UDPAddr
	RelayInternal                 [constants.MaxRelays]byte
	TargetVersion                 string
	UpcomingMagic                 [constants.MagicBytes]byte
	CurrentMagic                  [constants.MagicBytes]byte
	PreviousMagic                 [constants.MagicBytes]byte
	ExpectedPublicAddress         net.UDPAddr
	ExpectedInternalAddress       net.UDPAddr
	ExpectedHasInternalAddress    uint8
	ExpectedRelayPublicKey        [crypto.Box_PublicKeySize]byte
	ExpectedRelayBackendPublicKey [crypto.Box_PublicKeySize]byte
	TestToken                     [constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES]byte
	PingKey                       [crypto.Auth_KeySize]byte
}

func (packet *RelayUpdateResponsePacket) GetMaxSize() int {
	size := 256
	size += int(packet.NumRelays) * (8 + 7 + 1)
	size += constants.MaxRelayVersionLength
	size += constants.MagicBytes * 3
	size += 7 * 2
	size += 1 + 2*crypto.Box_PublicKeySize
	size += constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES
	size += crypto.Auth_KeySize
	return size
}

func (packet *RelayUpdateResponsePacket) Write(buffer []byte) []byte {

	index := 0

	if packet.Version < RelayUpdateResponsePacket_VersionMin || packet.Version > RelayUpdateResponsePacket_VersionMax {
		panic(fmt.Sprintf("invalid relay update request packet version %d", packet.Version))
	}

	encoding.WriteUint8(buffer, &index, packet.Version)
	encoding.WriteUint64(buffer, &index, uint64(packet.Timestamp))
	encoding.WriteUint32(buffer, &index, uint32(packet.NumRelays))

	for i := 0; i < int(packet.NumRelays); i++ {
		encoding.WriteUint64(buffer, &index, packet.RelayId[i])
		encoding.WriteAddress(buffer, &index, &packet.RelayAddress[i])
		encoding.WriteUint8(buffer, &index, packet.RelayInternal[i])
	}

	encoding.WriteString(buffer, &index, packet.TargetVersion, constants.MaxRelayVersionLength)

	encoding.WriteBytes(buffer, &index, packet.UpcomingMagic[:], constants.MagicBytes)
	encoding.WriteBytes(buffer, &index, packet.CurrentMagic[:], constants.MagicBytes)
	encoding.WriteBytes(buffer, &index, packet.PreviousMagic[:], constants.MagicBytes)

	encoding.WriteAddress(buffer, &index, &packet.ExpectedPublicAddress)
	encoding.WriteUint8(buffer, &index, packet.ExpectedHasInternalAddress)
	if packet.ExpectedHasInternalAddress != 0 {
		encoding.WriteAddress(buffer, &index, &packet.ExpectedInternalAddress)
	}
	encoding.WriteBytes(buffer, &index, packet.ExpectedRelayPublicKey[:], crypto.Box_PublicKeySize)
	encoding.WriteBytes(buffer, &index, packet.ExpectedRelayBackendPublicKey[:], crypto.Box_PublicKeySize)

	encoding.WriteBytes(buffer, &index, packet.TestToken[:], constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	encoding.WriteBytes(buffer, &index, packet.PingKey[:], crypto.Auth_KeySize)

	return buffer[:index]
}

func (packet *RelayUpdateResponsePacket) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &packet.Version) {
		return errors.New("could not read version")
	}

	if packet.Version < RelayUpdateResponsePacket_VersionMin || packet.Version > RelayUpdateResponsePacket_VersionMax {
		return errors.New("invalid relay update response packet version")
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

	if !encoding.ReadBytes(buffer, &index, packet.UpcomingMagic[:], constants.MagicBytes) {
		return errors.New("could not read upcoming magic")
	}

	if !encoding.ReadBytes(buffer, &index, packet.CurrentMagic[:], constants.MagicBytes) {
		return errors.New("could not read current magic")
	}

	if !encoding.ReadBytes(buffer, &index, packet.PreviousMagic[:], constants.MagicBytes) {
		return errors.New("could not read previous magic")
	}

	if !encoding.ReadAddress(buffer, &index, &packet.ExpectedPublicAddress) {
		return errors.New("could not read expected public address")
	}

	if !encoding.ReadUint8(buffer, &index, &packet.ExpectedHasInternalAddress) {
		return errors.New("could not read expected has internal address")
	}

	if packet.ExpectedHasInternalAddress != 0 {
		if !encoding.ReadAddress(buffer, &index, &packet.ExpectedInternalAddress) {
			return errors.New("could not read expected internal address")
		}
	}

	if !encoding.ReadBytes(buffer, &index, packet.ExpectedRelayPublicKey[:], crypto.Box_PublicKeySize) {
		return errors.New("could not read expected relay public key")
	}

	if !encoding.ReadBytes(buffer, &index, packet.ExpectedRelayBackendPublicKey[:], crypto.Box_PublicKeySize) {
		return errors.New("could not read expected relay backend public key")
	}

	if !encoding.ReadBytes(buffer, &index, packet.TestToken[:], constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES) {
		return errors.New("could not read test token")
	}

	if !encoding.ReadBytes(buffer, &index, packet.PingKey[:], crypto.Auth_KeySize) {
		return errors.New("could not read ping key")
	}

	return nil
}
