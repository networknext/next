package common

import (
	"fmt"
	"net"

	"github.com/networknext/backend/modules/encoding"
)

const (
	CostMatrixVersion_Min = 2      // the minimum version we can read
	CostMatrixVersion_Max = 2      // the maximum version we can read
	CostMatrixVersion_Write = 2    // the version we write

	MaxRelayNameLength = 63

	InvalidRouteValue = 10000
)

type CostMatrix struct {
	Version            uint32
	RelayIds           []uint64
	RelayAddresses     []net.UDPAddr
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIds []uint64
	DestRelays         []bool
	Costs              []int32
}

func (m *CostMatrix) Serialize(stream encoding.Stream) error {

	stream.SerializeUint32(&m.Version)

	if stream.IsReading() && (m.Version < CostMatrixVersion_Min || m.Version > CostMatrixVersion_Max) {
		return fmt.Errorf("invalid cost matrix version: %d", m.Version)
	}

	numRelays := uint32(len(m.RelayIds))
	stream.SerializeUint32(&numRelays)

	if stream.IsReading() {
		m.RelayIds = make([]uint64, numRelays)
		m.RelayAddresses = make([]net.UDPAddr, numRelays)
		m.RelayNames = make([]string, numRelays)
		m.RelayLatitudes = make([]float32, numRelays)
		m.RelayLongitudes = make([]float32, numRelays)
		m.RelayDatacenterIds = make([]uint64, numRelays)
	}

	for i := uint32(0); i < numRelays; i++ {
		stream.SerializeUint64(&m.RelayIds[i])
		stream.SerializeAddress(&m.RelayAddresses[i])
		stream.SerializeString(&m.RelayNames[i], MaxRelayNameLength)
		stream.SerializeFloat32(&m.RelayLatitudes[i])
		stream.SerializeFloat32(&m.RelayLongitudes[i])
		stream.SerializeUint64(&m.RelayDatacenterIds[i])
	}

	costsLength := uint32(len(m.Costs))
	stream.SerializeUint32(&costsLength)
	if stream.IsReading() {
		m.Costs = make([]int32, costsLength)
	}

	for i := uint32(0); i < costsLength; i++ {
		stream.SerializeInteger(&m.Costs[i], -1, InvalidRouteValue)
	}

	if m.Version >= 2 {
		if stream.IsReading() {
			m.DestRelays = make([]bool, numRelays)
		}
		for i := range m.DestRelays {
			stream.SerializeBool(&m.DestRelays[i])
		}
	}

	return stream.Error()
}

// todo: we need read and write unit tests for the cost matrix

// todo: tests should include writing with the new codebase, and reading with the old codebase

func (m *CostMatrix) Write(bufferSize int) ([]byte, error) {
	// todo: do we really want to allocate this here?
	buffer := make([]byte, bufferSize)
	ws := encoding.CreateWriteStream(buffer)
	if err := m.Serialize(ws); err != nil {
		return nil, fmt.Errorf("failed to serialize cost matrix: %v", err)
	}
	ws.Flush()
	return buffer[:ws.GetBytesProcessed()], nil
}

func (m *CostMatrix) Read(buffer []byte) error {
	readStream := encoding.CreateReadStream(buffer)
	return m.Serialize(readStream)
}
