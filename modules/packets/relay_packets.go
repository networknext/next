package packets

import (
	"net"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
)

const (
	VersionNumberRelayUpdateRequest  = 5
	VersionNumberRelayUpdateResponse = 1
	MaxRelayVersionStringLength      = 32
	MaxRelayAddressLength            = 256
)

// --------------------------------------------------------------------------

type RelayUpdateRequestPacket struct {
	Version           uint32
	Address           net.UDPAddr
	Token             []byte
	NumSamples        uint32
	SampleRelayId     []uint64
	SampleRTT         []float32
	SampleJitter      []float32
	SamplePacketLoss  []float32
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
	encoding.WriteBytes(buffer, &index, packet.Token, crypto.Box_KeySize)

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

/*
	var numRelays uint32

	var addr string
	if !encoding.ReadString(buff, &index, &addr, routing.MaxRelayAddressLength) {
		return errors.New("could not read relay address")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not convert address '%s' with reason: %v", addr, err)
	}

	if !encoding.ReadBytes(buff, &index, &r.Token, crypto_old.KeySize) {
		return errors.New("could not read relay token")
	}

	if !encoding.ReadUint32(buff, &index, &numRelays) {
		return errors.New("could not read num relays")
	}

	r.PingStats = make([]routing.RelayStatsPing, numRelays)

	for i := 0; i < int(numRelays); i++ {

		stats := &r.PingStats[i]

		// todo: these could be much more efficient as byte values [0,255]
		encoding.ReadUint64(buff, &index, &stats.RelayID)
		encoding.ReadFloat32(buff, &index, &stats.RTT)
		encoding.ReadFloat32(buff, &index, &stats.Jitter)
		encoding.ReadFloat32(buff, &index, &stats.PacketLoss)
	}

	if !encoding.ReadUint64(buff, &index, &r.SessionCount) {
		return errors.New("could not read session count")
	}

	if !encoding.ReadBool(buff, &index, &r.ShuttingDown) {
		return errors.New("could not read shutdown flag")
	}

	if !encoding.ReadString(buff, &index, &r.RelayVersion, MaxVersionStringLength) {
		return errors.New("could not read relay version string")
	}

	if !encoding.ReadUint8(buff, &index, &r.CPU) {
		return errors.New("could not read cpu")
	}

	if !encoding.ReadUint64(buff, &index, &r.EnvelopeUpKbps) {
		return errors.New("could not read envelope up kpbs")
	}

	if !encoding.ReadUint64(buff, &index, &r.EnvelopeDownKbps) {
		return errors.New("could not read envelope down kpbs")
	}

	if !encoding.ReadUint64(buff, &index, &r.BandwidthSentKbps) {
		return errors.New("could not read bandwidth sent kbps")
	}

	if !encoding.ReadUint64(buff, &index, &r.BandwidthRecvKbps) {
		return errors.New("could not read bandwidth received kbps")
	}
*/
	return nil
}

// --------------------------------------------------------------------------

type RelayUpdateResponsePacket struct {
	Version           uint32
	Timestamp         int64
	NumRelays         int32
	RelayId           uint64
	RelayAddress      string
	TargetVersion     string
	UpcomingMagic     []byte
	CurrentMagic      []byte
	PreviousMagic     []byte
	InternalAddresses []string
}

func (packet *RelayUpdateResponsePacket) Write(buffer []byte) []byte {
	// todo
	return nil
}

func (packet *RelayUpdateResponsePacket) Read(buffer []byte) error {
	// todo
	return nil
}

// ---------------------------------------------------------------

/*
func (r RelayUpdateResponse) MarshalBinary() ([]byte, error) {
	index := 0
	responseData := make([]byte, r.Size())
	encoding.WriteUint32(responseData, &index, r.Version)
	encoding.WriteUint64(responseData, &index, uint64(r.Timestamp))
	encoding.WriteUint32(responseData, &index, uint32(len(r.RelaysToPing)))
	for i := range r.RelaysToPing {
		encoding.WriteUint64(responseData, &index, r.RelaysToPing[i].ID)
		encoding.WriteString(responseData, &index, r.RelaysToPing[i].Address, routing.MaxRelayAddressLength)
	}
	encoding.WriteString(responseData, &index, r.TargetVersion, MaxVersionStringLength)

	if r.Version >= 1 {
		encoding.WriteBytes(responseData, &index, r.UpcomingMagic, 8)
		encoding.WriteBytes(responseData, &index, r.CurrentMagic, 8)
		encoding.WriteBytes(responseData, &index, r.PreviousMagic, 8)

		encoding.WriteUint32(responseData, &index, uint32(len(r.InternalAddrs)))
		for i := range r.InternalAddrs {
			encoding.WriteString(responseData, &index, r.InternalAddrs[i], routing.MaxRelayAddressLength)
		}
	}

	return responseData[:index], nil
}

func (r *RelayUpdateResponse) UnmarshalBinary(buff []byte) error {
	index := 0

	if !encoding.ReadUint32(buff, &index, &r.Version) {
		return errors.New("could not read version")
	}
	if r.Version > VersionNumberUpdateResponse {
		return errors.New("invalid version number")
	}

	var timestamp uint64
	if !encoding.ReadUint64(buff, &index, &timestamp) {
		return errors.New("could not read timestamp")
	}

	r.Timestamp = int64(timestamp)

	var numRelaysToPing uint32
	if !encoding.ReadUint32(buff, &index, &numRelaysToPing) {
		return errors.New("could not read num relays to ping")
	}

	r.RelaysToPing = make([]routing.RelayPingData, int32(numRelaysToPing))
	for i := 0; i < int(numRelaysToPing); i++ {
		stats := &r.RelaysToPing[i]

		// todo: these could be much more efficient as byte values [0,255]
		if !encoding.ReadUint64(buff, &index, &stats.ID) {
			return errors.New("could not read relay ID")
		}
		if !encoding.ReadString(buff, &index, &stats.Address, routing.MaxRelayAddressLength) {
			return errors.New("could not read relay address")
		}
	}

	if !encoding.ReadString(buff, &index, &r.TargetVersion, MaxVersionStringLength) {
		return errors.New("could not read target version")
	}

	if r.Version >= 1 {
		if !encoding.ReadBytes(buff, &index, &r.UpcomingMagic, 8) {
			return errors.New("could not read upcoming magic")
		}

		if !encoding.ReadBytes(buff, &index, &r.CurrentMagic, 8) {
			return errors.New("could not read current magic")
		}

		if !encoding.ReadBytes(buff, &index, &r.PreviousMagic, 8) {
			return errors.New("could not read previous magic")
		}

		var numInternalAddrs uint32
		if !encoding.ReadUint32(buff, &index, &numInternalAddrs) {
			return errors.New("could not read num internal addrs")
		}

		r.InternalAddrs = make([]string, int32(numInternalAddrs))
		for i := 0; i < int(numInternalAddrs); i++ {
			if !encoding.ReadString(buff, &index, &r.InternalAddrs[i], routing.MaxRelayAddressLength) {
				return errors.New("could not read relay internal address")
			}
		}
	}

	return nil
}
*/
