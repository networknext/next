package transport

import (
	"errors"
	"fmt"
	"math"
	"net"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/routing"
)

const (
	VersionNumberInitRequest    = 1
	VersionNumberInitResponse   = 0
	VersionNumberUpdateRequest  = 2
	VersionNumberUpdateResponse = 0

	PacketSizeRelayInitResponse = 4 + 8 + crypto.KeySize
)

// RelayInitRequest is the struct that describes the packets comming into the relay_init endpoint
type RelayInitRequest struct {
	Magic          uint32
	Version        uint32
	Nonce          []byte
	Address        net.UDPAddr
	EncryptedToken []byte
	RelayVersion   string
}

// UnmarshalBinary decodes binary data into a RelayInitRequest struct

func (r *RelayInitRequest) UnmarshalBinary(buf []byte) error {
	index := 0
	if !encoding.ReadUint32(buf, &index, &r.Magic) {
		return errors.New("invalid packet, unable to read magic number")
	}

	if !encoding.ReadUint32(buf, &index, &r.Version) {
		return errors.New("invalid packet, unable to read packet version")
	}

	switch r.Version {
	case 0:
		return r.unmarshalBinaryV0(buf, index)
	case 1:
		return r.unmarshalBinaryV1(buf, index)
	default:
		return fmt.Errorf("invalid packet, unknown version: %d", r.Version)
	}

	return nil
}

func (r *RelayInitRequest) unmarshalBinaryV0(buf []byte, index int) error {
	var addr string
	if !encoding.ReadBytes(buf, &index, &r.Nonce, crypto.NonceSize) {
		return errors.New("invalid packet, unable to read nonce")
	}

	if !encoding.ReadString(buf, &index, &addr, routing.MaxRelayAddressLength) {
		return errors.New("invalid packet, unable to read relay address")
	}

	if !encoding.ReadBytes(buf, &index, &r.EncryptedToken, routing.EncryptedRelayTokenSize) {
		return errors.New("invalid packet, unable to read encrypted token")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not resolve init packet with address '%s' with reason: %v", addr, err)
	}

	return nil
}

func (r *RelayInitRequest) unmarshalBinaryV1(buf []byte, index int) error {
	var addr string
	if !encoding.ReadBytes(buf, &index, &r.Nonce, crypto.NonceSize) {
		return errors.New("invalid packet, unable to read nonce")
	}

	if !encoding.ReadString(buf, &index, &addr, routing.MaxRelayAddressLength) {
		return errors.New("invalid packet, unable to read relay address")
	}

	if !encoding.ReadBytes(buf, &index, &r.EncryptedToken, routing.EncryptedRelayTokenSize) {
		return errors.New("invalid packet, unable to read encrypted token")
	}

	if !encoding.ReadString(buf, &index, &r.RelayVersion, math.MaxUint32) {
		return errors.New("invalid packet, unable to read relay version")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not resolve init packet with address '%s' with reason: %v", addr, err)
	}

	return nil
}

// MarshalBinary ...
func (r RelayInitRequest) MarshalBinary() ([]byte, error) {
	switch r.Version {
	case 0:
		return r.marshalBinaryV0()
	case 1:
		return r.marshalBinaryV1()
	default:
		return nil, fmt.Errorf("invalid init request version: %d", r.Version)
	}
}

func (r *RelayInitRequest) marshalBinaryV0() ([]byte, error) {
	data := make([]byte, 4+4+crypto.NonceSize+4+len(r.Address.String())+routing.EncryptedRelayTokenSize)
	index := 0
	encoding.WriteUint32(data, &index, r.Magic)
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteBytes(data, &index, r.Nonce, crypto.NonceSize)
	encoding.WriteString(data, &index, r.Address.String(), uint32(len(r.Address.String())))
	encoding.WriteBytes(data, &index, r.EncryptedToken, routing.EncryptedRelayTokenSize)

	return data, nil
}

func (r *RelayInitRequest) marshalBinaryV1() ([]byte, error) {
	data := make([]byte, 4+4+crypto.NonceSize+4+len(r.Address.String())+routing.EncryptedRelayTokenSize+4+len(r.RelayVersion))
	index := 0
	encoding.WriteUint32(data, &index, r.Magic)
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteBytes(data, &index, r.Nonce, crypto.NonceSize)
	encoding.WriteString(data, &index, r.Address.String(), uint32(len(r.Address.String())))
	encoding.WriteBytes(data, &index, r.EncryptedToken, routing.EncryptedRelayTokenSize)
	encoding.WriteString(data, &index, r.RelayVersion, uint32(len(r.RelayVersion)))

	return data, nil
}

// RelayInitResponse ...
type RelayInitResponse struct {
	Version   uint32
	Timestamp uint64
	PublicKey []byte
}

func (r RelayInitResponse) MarshalBinary() ([]byte, error) {
	index := 0
	responseData := make([]byte, PacketSizeRelayInitResponse)

	encoding.WriteUint32(responseData, &index, VersionNumberInitResponse)
	encoding.WriteUint64(responseData, &index, r.Timestamp)
	encoding.WriteBytes(responseData, &index, r.PublicKey, crypto.KeySize)

	return responseData, nil
}

func (r *RelayInitResponse) UnmarshalBinary(buf []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(buf, &index, &version) {
		return errors.New("failed to unmarshal relay init response version")
	}

	var timestamp uint64
	if !encoding.ReadUint64(buf, &index, &timestamp) {
		return errors.New("failed to unmarshal relay init response timestamp")
	}

	var publicKey []byte
	if !encoding.ReadBytes(buf, &index, &publicKey, crypto.KeySize) {
		return errors.New("failed to unmarshal relay init response public key")
	}

	r.Version = version
	r.Timestamp = timestamp
	r.PublicKey = publicKey

	return nil
}

type RelayUpdateRequest struct {
	Version      uint32
	RelayVersion string
	Address      net.UDPAddr
	Token        []byte

	PingStats    []routing.RelayStatsPing
	TrafficStats routing.TrafficStats

	ShuttingDown bool

	CPUUsage float64
	MemUsage float64
}

func (r *RelayUpdateRequest) UnmarshalBinary(buff []byte) error {
	index := 0
	if !encoding.ReadUint32(buff, &index, &r.Version) {
		return errors.New("invalid packet, could not read packet version")
	}

	switch r.Version {
	case 0:
		return r.unmarshalBinaryV0(buff, index)
	case 1:
		return r.unmarshalBinaryV1(buff, index)
	case 2:
		return r.unmarshalBinaryV2(buff, index)
	default:
		return fmt.Errorf("invalid packet, unknown version: %d", r.Version)
	}
}

func (r *RelayUpdateRequest) unmarshalBinaryV0(buff []byte, index int) error {
	var numRelays uint32

	var addr string
	if !(encoding.ReadString(buff, &index, &addr, routing.MaxRelayAddressLength) &&
		encoding.ReadBytes(buff, &index, &r.Token, crypto.KeySize) &&
		encoding.ReadUint32(buff, &index, &numRelays)) {
		return errors.New("invalid packet")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not resolve init packet with address '%s' with reason: %v", addr, err)
	}

	r.PingStats = make([]routing.RelayStatsPing, numRelays)
	for i := 0; i < int(numRelays); i++ {
		stats := &r.PingStats[i]

		if !(encoding.ReadUint64(buff, &index, &stats.RelayID) &&
			encoding.ReadFloat32(buff, &index, &stats.RTT) &&
			encoding.ReadFloat32(buff, &index, &stats.Jitter) &&
			encoding.ReadFloat32(buff, &index, &stats.PacketLoss)) {
			return errors.New("invalid packet, could not read a ping stat")
		}
	}

	if err := r.TrafficStats.ReadFrom(buff, &index, 0); err != nil {
		return err
	}

	if !encoding.ReadBool(buff, &index, &r.ShuttingDown) {
		return errors.New("invalid packet, could not read shutdown flag")
	}

	if !encoding.ReadFloat64(buff, &index, &r.CPUUsage) {
		return errors.New("invalid packet, could not read cpu usage")
	}

	if !encoding.ReadFloat64(buff, &index, &r.MemUsage) {
		return errors.New("invalid packet, could not read memory usage")
	}

	if !encoding.ReadString(buff, &index, &r.RelayVersion, math.MaxUint32) {
		return errors.New("invalid packet, could not read relay version")
	}

	return nil
}

func (r *RelayUpdateRequest) unmarshalBinaryV1(buff []byte, index int) error {
	var numRelays uint32

	var addr string
	if !(encoding.ReadString(buff, &index, &addr, routing.MaxRelayAddressLength) &&
		encoding.ReadBytes(buff, &index, &r.Token, crypto.KeySize) &&
		encoding.ReadUint32(buff, &index, &numRelays)) {
		return errors.New("invalid packet")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not resolve init packet with address '%s' with reason: %v", addr, err)
	}

	r.PingStats = make([]routing.RelayStatsPing, numRelays)
	for i := 0; i < int(numRelays); i++ {
		stats := &r.PingStats[i]

		if !(encoding.ReadUint64(buff, &index, &stats.RelayID) &&
			encoding.ReadFloat32(buff, &index, &stats.RTT) &&
			encoding.ReadFloat32(buff, &index, &stats.Jitter) &&
			encoding.ReadFloat32(buff, &index, &stats.PacketLoss)) {
			return errors.New("invalid packet, could not read a ping stat")
		}
	}

	if err := r.TrafficStats.ReadFrom(buff, &index, 1); err != nil {
		return err
	}

	r.TrafficStats.BytesReceived = r.TrafficStats.AllRx()

	r.TrafficStats.BytesSent = r.TrafficStats.AllTx()

	if !encoding.ReadBool(buff, &index, &r.ShuttingDown) {
		return errors.New("invalid packet, could not read shutdown flag")
	}

	if !encoding.ReadFloat64(buff, &index, &r.CPUUsage) {
		return errors.New("invalid packet, could not read cpu usage")
	}

	if !encoding.ReadFloat64(buff, &index, &r.MemUsage) {
		return errors.New("invalid packet, could not read memory usage")
	}

	if !encoding.ReadString(buff, &index, &r.RelayVersion, math.MaxUint32) {
		return errors.New("invalid packet, could not read relay version")
	}

	return nil
}

func (r *RelayUpdateRequest) unmarshalBinaryV2(buff []byte, index int) error {
	var numRelays uint32

	var addr string
	if !(encoding.ReadString(buff, &index, &addr, routing.MaxRelayAddressLength) &&
		encoding.ReadBytes(buff, &index, &r.Token, crypto.KeySize) &&
		encoding.ReadUint32(buff, &index, &numRelays)) {
		return errors.New("invalid packet")
	}

	if udp, err := net.ResolveUDPAddr("udp", addr); udp != nil && err == nil {
		r.Address = *udp
	} else {
		return fmt.Errorf("could not resolve init packet with address '%s' with reason: %v", addr, err)
	}

	r.PingStats = make([]routing.RelayStatsPing, numRelays)
	for i := 0; i < int(numRelays); i++ {
		stats := &r.PingStats[i]

		if !(encoding.ReadUint64(buff, &index, &stats.RelayID) &&
			encoding.ReadFloat32(buff, &index, &stats.RTT) &&
			encoding.ReadFloat32(buff, &index, &stats.Jitter) &&
			encoding.ReadFloat32(buff, &index, &stats.PacketLoss)) {
			return errors.New("invalid packet, could not read a ping stat")
		}
	}

	if err := r.TrafficStats.ReadFrom(buff, &index, 2); err != nil {
		return err
	}

	r.TrafficStats.BytesReceived = r.TrafficStats.AllRx()

	r.TrafficStats.BytesSent = r.TrafficStats.AllTx()

	if !encoding.ReadBool(buff, &index, &r.ShuttingDown) {
		return errors.New("invalid packet, could not read shutdown flag")
	}

	if !encoding.ReadFloat64(buff, &index, &r.CPUUsage) {
		return errors.New("invalid packet, could not read cpu usage")
	}

	if !encoding.ReadFloat64(buff, &index, &r.MemUsage) {
		return errors.New("invalid packet, could not read memory usage")
	}

	return nil
}

// MarshalBinary ...
func (r RelayUpdateRequest) MarshalBinary() ([]byte, error) {
	switch r.Version {
	case 0:
		return r.marshalBinaryV0()
	case 1:
		return r.marshalBinaryV1()
	case 2:
		return r.marshalBinaryV2()
	default:
		return nil, fmt.Errorf("invalid update request version: %d", r.Version)
	}
}

func (r RelayUpdateRequest) marshalBinaryV0() ([]byte, error) {
	data := make([]byte, r.sizeV0())

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address.String(), math.MaxInt32)
	encoding.WriteBytes(data, &index, r.Token, crypto.KeySize)
	encoding.WriteUint32(data, &index, uint32(len(r.PingStats)))

	for i := 0; i < len(r.PingStats); i++ {
		stats := &r.PingStats[i]

		encoding.WriteUint64(data, &index, stats.RelayID)
		encoding.WriteUint32(data, &index, math.Float32bits(stats.RTT))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.Jitter))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.PacketLoss))
	}

	r.TrafficStats.WriteTo(data, &index, 0)
	encoding.WriteBool(data, &index, r.ShuttingDown)
	encoding.WriteFloat64(data, &index, r.CPUUsage)
	encoding.WriteFloat64(data, &index, r.MemUsage)
	encoding.WriteString(data, &index, r.RelayVersion, uint32(len(r.RelayVersion)))

	return data[:index], nil
}

func (r RelayUpdateRequest) marshalBinaryV1() ([]byte, error) {
	data := make([]byte, r.sizeV1())

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address.String(), math.MaxInt32)
	encoding.WriteBytes(data, &index, r.Token, crypto.KeySize)
	encoding.WriteUint32(data, &index, uint32(len(r.PingStats)))

	for i := 0; i < len(r.PingStats); i++ {
		stats := &r.PingStats[i]

		encoding.WriteUint64(data, &index, stats.RelayID)
		encoding.WriteUint32(data, &index, math.Float32bits(stats.RTT))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.Jitter))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.PacketLoss))
	}

	r.TrafficStats.WriteTo(data, &index, 1)
	encoding.WriteBool(data, &index, r.ShuttingDown)
	encoding.WriteFloat64(data, &index, r.CPUUsage)
	encoding.WriteFloat64(data, &index, r.MemUsage)
	encoding.WriteString(data, &index, r.RelayVersion, uint32(len(r.RelayVersion)))

	return data[:index], nil
}

func (r *RelayUpdateRequest) marshalBinaryV2() ([]byte, error) {
	data := make([]byte, r.sizeV2())

	index := 0
	encoding.WriteUint32(data, &index, r.Version)
	encoding.WriteString(data, &index, r.Address.String(), math.MaxInt32)
	encoding.WriteBytes(data, &index, r.Token, crypto.KeySize)
	encoding.WriteUint32(data, &index, uint32(len(r.PingStats)))

	for i := 0; i < len(r.PingStats); i++ {
		stats := &r.PingStats[i]

		encoding.WriteUint64(data, &index, stats.RelayID)
		encoding.WriteUint32(data, &index, math.Float32bits(stats.RTT))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.Jitter))
		encoding.WriteUint32(data, &index, math.Float32bits(stats.PacketLoss))
	}

	r.TrafficStats.WriteTo(data, &index, 2)
	encoding.WriteBool(data, &index, r.ShuttingDown)
	encoding.WriteFloat64(data, &index, r.CPUUsage)
	encoding.WriteFloat64(data, &index, r.MemUsage)

	return data[:index], nil
}

func (r *RelayUpdateRequest) sizeV0() uint {
	return uint(4 + // version
		4 + // address length
		len(r.Address.String()) + // address
		crypto.KeySize + // public key
		4 + // number of ping stats
		20*len(r.PingStats) + // ping stats list
		8 + // session count
		8 + // bytes sent
		8 + // bytes received
		1 + // shutdown flag
		8 + // cpu usage
		8 + // memory usage
		4 + // length of relay version
		len(r.RelayVersion)) // relay version string
}

// Changes from v0
// Removed bytes sent/received
// Added individual stats
func (r *RelayUpdateRequest) sizeV1() uint {
	return uint(4 + // version
		4 + // address length
		len(r.Address.String()) + // address
		crypto.KeySize + // public key
		4 + // number of ping stats
		20*len(r.PingStats) + // ping stats list
		8 + // session count
		8 + // outbound ping tx
		8 + // route request rx
		8 + // route request tx
		8 + // route response rx
		8 + // route response tx
		8 + // client to server rx
		8 + // client to server tx
		8 + // server to client rx
		8 + // server to client tx
		8 + // inbound ping rx
		8 + // inbound ping tx
		8 + // pong rx
		8 + // session ping rx
		8 + // session ping tx
		8 + // session pong rx
		8 + // session pong tx
		8 + // continue request rx
		8 + // continue request tx
		8 + // continue response rx
		8 + // continue response tx
		8 + // near ping rx
		8 + // near ping tx
		8 + // unknown Rx
		1 + // shutdown flag
		8 + // cpu usage
		8 + // memory usage
		4 + // length of relay version
		len(r.RelayVersion)) // relay version string
}

// Changes from v1:
// Added envelope up/down
// Removed version string
func (r *RelayUpdateRequest) sizeV2() uint {
	return uint(4 + // version
		4 + // address length
		len(r.Address.String()) + // address
		crypto.KeySize + // public key
		4 + // number of ping stats
		20*len(r.PingStats) + // ping stats list
		8 + // session count
		8 + // envelope up
		8 + // envelope down
		8 + // outbound ping tx
		8 + // route request rx
		8 + // route request tx
		8 + // route response rx
		8 + // route response tx
		8 + // client to server rx
		8 + // client to server tx
		8 + // server to client rx
		8 + // server to client tx
		8 + // inbound ping rx
		8 + // inbound ping tx
		8 + // pong rx
		8 + // session ping rx
		8 + // session ping tx
		8 + // session pong rx
		8 + // session pong tx
		8 + // continue request rx
		8 + // continue request tx
		8 + // continue response rx
		8 + // continue response tx
		8 + // near ping rx
		8 + // near ping tx
		8 + // unknown Rx
		1 + // shutdown flag
		8 + // cpu usage
		8) // memory usage
}

type RelayUpdateResponse struct {
	Timestamp    int64
	RelaysToPing []routing.RelayPingData
}

func (r *RelayUpdateResponse) UnmarshalBinary(buff []byte) error {
	index := 0
	var version uint32
	if !encoding.ReadUint32(buff, &index, &version) {
		return errors.New("failed to unmarshal relay update response version")
	}

	var timestamp uint64
	if !encoding.ReadUint64(buff, &index, &timestamp) {
		return errors.New("failed to unmarshal relay update response timestamp")
	}
	r.Timestamp = int64(timestamp)

	var numRelaysToPing uint32
	if !encoding.ReadUint32(buff, &index, &numRelaysToPing) {
		return errors.New("failed to unmarshal relay update response number of relays to ping")
	}

	for i := 0; uint32(i) < numRelaysToPing; i++ {
		var id uint64
		if !encoding.ReadUint64(buff, &index, &id) {
			return errors.New("failed to unmarshal relay update response relay id")
		}

		var addr string
		if !encoding.ReadString(buff, &index, &addr, routing.MaxRelayAddressLength) {
			return errors.New("failed to unmarshal relay update response relay address")
		}

		r.RelaysToPing = append(r.RelaysToPing, routing.RelayPingData{
			ID:      id,
			Address: addr,
		})
	}

	return nil
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

	return responseData[:index], nil
}

func (r *RelayUpdateResponse) size() int {
	return 4 + 8 + 4 + len(r.RelaysToPing)*(8+routing.MaxRelayAddressLength)
}
