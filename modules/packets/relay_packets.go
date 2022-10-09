package packets

import (
	"net"

	// "github.com/networknext/backend/modules/encoding"
)

const (
	VersionNumberRelayUpdateRequest  = 5
	VersionNumberRelayUpdateResponse = 1
	MaxRelayVersionStringLength      = 32
)

// --------------------------------------------------------------------------

type RelayUpdateRequestPacket struct {
	Version           uint32
	Address           net.UDPAddr
	Token             []byte
	// todo: SoA
	// PingStats         []routing.RelayStatsPing
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
	// todo
	return nil
}

func (packet *RelayUpdateRequestPacket) Read(buffer []byte) error {
	// todo
	return nil
}

// --------------------------------------------------------------------------

type RelayUpdateResponsePacket struct {
	Version       uint32
	Timestamp     int64
	// todo: structure of arrays for relay ping data
	// RelaysToPing  []routing.RelayPingData
	TargetVersion string
	UpcomingMagic []byte
	CurrentMagic  []byte
	PreviousMagic []byte
	InternalAddrs []string
}

func (packet *RelayUpdateResponsePacket) Write(buffer []byte) []byte {
	// todo
	return nil
}

func (packet *RelayUpdateResponsePacket) Read(buffer []byte) error {
	// todo
	return nil
}

// --------------------------------------------------------------------------

/*
func (r *RelayUpdateRequest) unmarshalBinaryV5(buff []byte, index int) error {

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

	return nil
}

func (r *RelayUpdateRequest) marshalBinaryV5() ([]byte, error) {
	data := make([]byte, r.sizeV5())

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address.String(), routing.MaxRelayAddressLength)
	encoding.WriteBytes(data, &index, r.Token, crypto_old.KeySize)

	encoding.WriteUint32(data, &index, uint32(len(r.PingStats)))
	for i := 0; i < len(r.PingStats); i++ {
		stats := r.PingStats[i]

		encoding.WriteUint64(data, &index, stats.RelayID)
		encoding.WriteFloat32(data, &index, stats.RTT)
		encoding.WriteFloat32(data, &index, stats.Jitter)
		encoding.WriteFloat32(data, &index, stats.PacketLoss)
	}

	encoding.WriteUint64(data, &index, r.SessionCount)
	encoding.WriteBool(data, &index, r.ShuttingDown)
	encoding.WriteString(data, &index, r.RelayVersion, uint32(len(r.RelayVersion)))

	encoding.WriteUint8(data, &index, r.CPU)

	encoding.WriteUint64(data, &index, r.EnvelopeUpKbps)
	encoding.WriteUint64(data, &index, r.EnvelopeDownKbps)
	encoding.WriteUint64(data, &index, r.BandwidthSentKbps)
	encoding.WriteUint64(data, &index, r.BandwidthRecvKbps)

	return data[:index], nil
}
*/

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
