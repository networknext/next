package transport

import (
	"errors"
	"fmt"
	"math"
	"net"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
)

const (
	VersionNumberUpdateRequest  = 2
	VersionNumberUpdateResponse = 0
)

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
	case 2:
		return r.unmarshalBinaryV2(buff, index)
	default:
		return fmt.Errorf("invalid packet, unknown version: %d", r.Version)
	}
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
	case 2:
		return r.marshalBinaryV2()
	default:
		return nil, fmt.Errorf("invalid update request version: %d", r.Version)
	}
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
