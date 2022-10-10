package packets

import (
	"net"
	"fmt"
	"errors"

	"github.com/networknext/backend/modules/encoding"
)

const (
	VersionNumberRelayUpdateRequest  = 5
	VersionNumberRelayUpdateResponse = 1
	MaxRelayVersionStringLength      = 32
	MaxRelayAddressLength            = 256
	RelayTokenSize                   = 32
	MaxRelays                        = 1024
)

// --------------------------------------------------------------------------

type RelayPacket interface {

	Write(buffer []byte) []byte

	Read(buffer []byte) error
}

// --------------------------------------------------------------------------

type RelayUpdateRequestPacket struct {
	Version           uint32
	Address           net.UDPAddr
	Token             []byte
	NumSamples        uint32
	SampleRelayId     [MaxRelays]uint64
	SampleRTT         [MaxRelays]float32
	SampleJitter      [MaxRelays]float32
	SamplePacketLoss  [MaxRelays]float32
	SessionCount      uint64
	ShuttingDown      bool
	RelayVersion      string
	CPU               uint8
	EnvelopeUpKbps    uint64
	EnvelopeDownKbps  uint64
	BandwidthSentKbps uint64
	BandwidthRecvKbps uint64
}

func (packet *RelayUpdateRequestPacket) Write(buffer []byte) []byte {

	index := 0

	encoding.WriteUint32(buffer, &index, packet.Version)
	encoding.WriteString(buffer, &index, packet.Address.String(), MaxRelayAddressLength)
	encoding.WriteBytes(buffer, &index, packet.Token, RelayTokenSize)

	encoding.WriteUint32(buffer, &index, packet.NumSamples)

	for i := 0; i < int(packet.NumSamples); i++ {
		encoding.WriteUint64(buffer, &index, packet.SampleRelayId[i])
		encoding.WriteFloat32(buffer, &index, packet.SampleRTT[i])
		encoding.WriteFloat32(buffer, &index, packet.SampleJitter[i])
		encoding.WriteFloat32(buffer, &index, packet.SamplePacketLoss[i])
	}

	encoding.WriteUint64(buffer, &index, packet.SessionCount)
	encoding.WriteBool(buffer, &index, packet.ShuttingDown)
	encoding.WriteString(buffer, &index, packet.RelayVersion, MaxRelayVersionStringLength)

	encoding.WriteUint8(buffer, &index, packet.CPU)

	encoding.WriteUint64(buffer, &index, packet.EnvelopeUpKbps)
	encoding.WriteUint64(buffer, &index, packet.EnvelopeDownKbps)
	encoding.WriteUint64(buffer, &index, packet.BandwidthSentKbps)
	encoding.WriteUint64(buffer, &index, packet.BandwidthRecvKbps)

	return buffer[:index]
}

func (packet *RelayUpdateRequestPacket) Read(buffer []byte) error {

	index := 0

	encoding.ReadUint32(buffer, &index, &packet.Version)

	if packet.Version != VersionNumberRelayUpdateRequest {
		return errors.New("invalid relay update request packet version")
	}

	var address string
	if !encoding.ReadString(buffer, &index, &address, MaxRelayAddressLength) {
		return errors.New("could not read relay address")
	}

	if udp, err := net.ResolveUDPAddr("udp", address); udp != nil && err == nil {
		packet.Address = *udp
	} else {
		return fmt.Errorf("could not resolve udp address '%s': %v", address, err)
	}

	if !encoding.ReadBytes(buffer, &index, &packet.Token, RelayTokenSize) {
		return errors.New("could not read relay token")
	}

	if !encoding.ReadUint32(buffer, &index, &packet.NumSamples) {
		return errors.New("could not read num samples")
	}

	if packet.NumSamples < 0 || packet.NumSamples > MaxRelays {
		return errors.New("invalid num samples")
	}

	for i := 0; i < int(packet.NumSamples); i++ {

		if !encoding.ReadUint64(buffer, &index, &packet.SampleRelayId[i]) {
			return errors.New("could not read sample relay id")
		}
		
		if !encoding.ReadFloat32(buffer, &index, &packet.SampleRTT[i]) {
			return errors.New("could not read sample rtt")			
		}
		
		if !encoding.ReadFloat32(buffer, &index, &packet.SampleJitter[i]) {
			return errors.New("could not read sample jitter")
		}

		if !encoding.ReadFloat32(buffer, &index, &packet.SamplePacketLoss[i]) {
			return errors.New("could not read sample packet loss")			
		}
	}

	if !encoding.ReadUint64(buffer, &index, &packet.SessionCount) {
		return errors.New("could not read session count")
	}

	if !encoding.ReadBool(buffer, &index, &packet.ShuttingDown) {
		return errors.New("could not read shutdown flag")
	}

	if !encoding.ReadString(buffer, &index, &packet.RelayVersion, MaxRelayVersionStringLength) {
		return errors.New("could not read relay version string")
	}

	if !encoding.ReadUint8(buffer, &index, &packet.CPU) {
		return errors.New("could not read cpu")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.EnvelopeUpKbps) {
		return errors.New("could not read envelope up kpbs")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.EnvelopeDownKbps) {
		return errors.New("could not read envelope down kpbs")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.BandwidthSentKbps) {
		return errors.New("could not read bandwidth sent kbps")
	}

	if !encoding.ReadUint64(buffer, &index, &packet.BandwidthRecvKbps) {
		return errors.New("could not read bandwidth received kbps")
	}

	return nil
}

func (packet *RelayUpdateRequestPacket) Peek(buffer []byte) error {

	index := 0

	encoding.ReadUint32(buffer, &index, &packet.Version)

	if packet.Version != VersionNumberRelayUpdateRequest {
		return errors.New("invalid relay update request packet version")
	}

	var address string
	if !encoding.ReadString(buffer, &index, &address, MaxRelayAddressLength) {
		return errors.New("could not read relay address")
	}

	if udp, err := net.ResolveUDPAddr("udp", address); udp != nil && err == nil {
		packet.Address = *udp
	} else {
		return fmt.Errorf("could not resolve udp address '%s': %v", address, err)
	}

	// todo: probably need to read in the token here too

	return nil
}

// --------------------------------------------------------------------------

type RelayUpdateResponsePacket struct {
	Version           uint32
	Timestamp         uint64
	NumRelays         uint32
	RelayId           [MaxRelays]uint64
	RelayAddress      [MaxRelays]string
	TargetVersion     string
	UpcomingMagic     []byte
	CurrentMagic      []byte
	PreviousMagic     []byte
}

func (packet *RelayUpdateResponsePacket) Write(buffer []byte) []byte {

	index := 0

	encoding.WriteUint32(buffer, &index, packet.Version)
	encoding.WriteUint64(buffer, &index, uint64(packet.Timestamp))
	encoding.WriteUint32(buffer, &index, uint32(packet.NumRelays))
	
	for i := 0; i < int(packet.NumRelays); i++ {
		encoding.WriteUint64(buffer, &index, packet.RelayId[i])
		encoding.WriteString(buffer, &index, packet.RelayAddress[i], MaxRelayAddressLength)
	}
	
	encoding.WriteString(buffer, &index, packet.TargetVersion, MaxRelayVersionStringLength)

	encoding.WriteBytes(buffer, &index, packet.UpcomingMagic, 8)
	encoding.WriteBytes(buffer, &index, packet.CurrentMagic, 8)
	encoding.WriteBytes(buffer, &index, packet.PreviousMagic, 8)

	// todo: remove this. we don't need internal addresses array anymore (as far as I can tell...)
	encoding.WriteUint32(buffer, &index, 0)

	return buffer[:index]
}

func (packet *RelayUpdateResponsePacket) Read(buffer []byte) error {
	
	index := 0

	if !encoding.ReadUint32(buffer, &index, &packet.Version) {
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

	if packet.NumRelays < 0 || packet.NumRelays > MaxRelays {
		return errors.New("invalid num relays")
	}

	for i := 0; i < int(packet.NumRelays); i++ {

		if !encoding.ReadUint64(buffer, &index, &packet.RelayId[i]) {
			return errors.New("could not read relay id")
		}

		if !encoding.ReadString(buffer, &index, &packet.RelayAddress[i], MaxRelayAddressLength) {
			return errors.New("could not read relay address")
		}
	}

	if !encoding.ReadString(buffer, &index, &packet.TargetVersion, MaxRelayVersionStringLength) {
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

	// todo: remove internal addresses. as far as I can tell they're no longer needed

	var numInternalAddresses uint32
	if !encoding.ReadUint32(buffer, &index, &numInternalAddresses) {
		return errors.New("could not read num internal addresses")
	}

	if numInternalAddresses > MaxRelays {
		return errors.New("invalid num internal addresses")
	}

	for i := 0; i < int(numInternalAddresses); i++ {
		var dummy string
		if !encoding.ReadString(buffer, &index, &dummy, MaxRelayAddressLength) {
			return errors.New("could not read relay internal address")
		}
	}

	return nil
}
