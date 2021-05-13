package transport

import (
	"errors"
	"fmt"
	"net"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
)

const (
	VersionNumberUpdateRequest  = 4
	VersionNumberUpdateResponse = 0
	MaxVersionStringLength      = 32
)

type RelayUpdateRequest struct {
	Version      uint32
	RelayVersion string
	Address      net.UDPAddr
	Token        []byte
	PingStats    []routing.RelayStatsPing
	SessionCount uint64
	ShuttingDown bool
	CPU          uint8
}

func (r *RelayUpdateRequest) UnmarshalBinary(buff []byte) error {
	index := 0
	if !encoding.ReadUint32(buff, &index, &r.Version) {
		return errors.New("invalid packet, could not read packet version")
	}

	switch r.Version {
	case 3:
		return r.unmarshalBinaryV3(buff, index)
	case 4:
		return r.unmarshalBinaryV4(buff, index)
	default:
		return fmt.Errorf("invalid packet, unknown version: %d", r.Version)
	}
}

func (r *RelayUpdateRequest) unmarshalBinaryV3(buff []byte, index int) error {
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

	if !encoding.ReadBytes(buff, &index, &r.Token, crypto.KeySize) {
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
		return errors.New("invalid packet, could not read session count")
	}

	if !encoding.ReadBool(buff, &index, &r.ShuttingDown) {
		return errors.New("invalid packet, could not read shutdown flag")
	}

	if !encoding.ReadString(buff, &index, &r.RelayVersion, MaxVersionStringLength) {
		return errors.New("invalid relay version string")
	}

	return nil
}

func (r *RelayUpdateRequest) unmarshalBinaryV4(buff []byte, index int) error {
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

	if !encoding.ReadBytes(buff, &index, &r.Token, crypto.KeySize) {
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
		return errors.New("invalid packet, could not read session count")
	}

	if !encoding.ReadBool(buff, &index, &r.ShuttingDown) {
		return errors.New("invalid packet, could not read shutdown flag")
	}

	if !encoding.ReadString(buff, &index, &r.RelayVersion, MaxVersionStringLength) {
		return errors.New("invalid relay version string")
	}

	if !encoding.ReadUint8(buff, &index, &r.CPU) {
		return errors.New("invalid packet, could not read cpu")
	}

	return nil
}

// Marshals a RelayUpdateRequest. Useful for tests and fake relays.
func (r RelayUpdateRequest) MarshalBinary() ([]byte, error) {
	switch r.Version {
	case 3:
		return r.marshalBinaryV3()
	case 4:
		return r.marshalBinaryV4()
	default:
		return nil, fmt.Errorf("invalid update request version: %d", r.Version)
	}
}

func (r *RelayUpdateRequest) marshalBinaryV3() ([]byte, error) {
	data := make([]byte, r.sizeV3())

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address.String(), routing.MaxRelayAddressLength)
	encoding.WriteBytes(data, &index, r.Token, crypto.KeySize)

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

	return data[:index], nil
}

func (r *RelayUpdateRequest) marshalBinaryV4() ([]byte, error) {
	data := make([]byte, r.sizeV4())

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address.String(), routing.MaxRelayAddressLength)
	encoding.WriteBytes(data, &index, r.Token, crypto.KeySize)

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

	return data[:index], nil
}

func (r *RelayUpdateRequest) sizeV3() int {
	return (4 + // version
		len(r.Address.String()) + // address
		crypto.KeySize + // public key
		4 + // number of ping stats
		len(r.PingStats)*(8+4+4+4) + // ping stats list
		8 + // session count
		1 + // shutting down
		MaxVersionStringLength) // relay version
}

func (r *RelayUpdateRequest) sizeV4() int {
	return (4 + // version
		len(r.Address.String()) + // address
		crypto.KeySize + // public key
		4 + // number of ping stats
		len(r.PingStats)*(8+4+4+4) + // ping stats list
		8 + // session count
		1 + // shutting down
		MaxVersionStringLength + // relay version
		1) // CPU
}

type RelayUpdateResponse struct {
	Timestamp     int64
	RelaysToPing  []routing.RelayPingData
	TargetVersion string
}

func (r RelayUpdateResponse) MarshalBinary() ([]byte, error) {
	index := 0
	responseData := make([]byte, r.size())
	encoding.WriteUint32(responseData, &index, VersionNumberUpdateResponse)
	encoding.WriteUint64(responseData, &index, uint64(r.Timestamp))
	encoding.WriteUint32(responseData, &index, uint32(len(r.RelaysToPing)))
	for i := range r.RelaysToPing {
		encoding.WriteUint64(responseData, &index, r.RelaysToPing[i].ID)
		encoding.WriteString(responseData, &index, r.RelaysToPing[i].Address, routing.MaxRelayAddressLength)
	}
	encoding.WriteString(responseData, &index, r.TargetVersion, MaxVersionStringLength)
	return responseData[:index], nil
}

func (r *RelayUpdateResponse) UnmarshalBinary(buff []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(buff, &index, &version) {
		return errors.New("could not read version")
	}
	if version > VersionNumberUpdateResponse {
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

	return nil
}

func (r RelayUpdateResponse) size() int {
	return (4 + // Version
		8 + // Timestamp
		4 + // Num relays to ping
		len(r.RelaysToPing)*(8+routing.MaxRelayAddressLength) + // Relays to ping
		MaxVersionStringLength) // Target Version
}
