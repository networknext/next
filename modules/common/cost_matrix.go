package common

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/encoding"
)

const (
	CostMatrixVersion_Min   = 1 // the minimum version we can read
	CostMatrixVersion_Max   = 2 // the maximum version we can read
	CostMatrixVersion_Write = 2 // the version we write
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
	Costs              []uint8
	RelayPrice         []uint8
}

func (m *CostMatrix) GetMaxSize() int {
	// IMPORTANT: This must be an upper bound *and* a multiple of 4
	numRelays := len(m.RelayIds)
	size := 256 + numRelays*(8+19+constants.MaxRelayNameLength+4+4+8+1) + core.TriMatrixLength(numRelays) + numRelays + 4
	size += 4
	size -= size % 4
	return size
}

func (m *CostMatrix) Serialize(stream encoding.Stream) error {

	if stream.IsWriting() && (m.Version < CostMatrixVersion_Min || m.Version > CostMatrixVersion_Max) {
		panic(fmt.Errorf("invalid cost matrix version: %d", m.Version))
	}

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
		stream.SerializeString(&m.RelayNames[i], constants.MaxRelayNameLength)
		stream.SerializeFloat32(&m.RelayLatitudes[i])
		stream.SerializeFloat32(&m.RelayLongitudes[i])
		stream.SerializeUint64(&m.RelayDatacenterIds[i])
	}

	if stream.IsReading() {
		costSize := core.TriMatrixLength(int(numRelays))
		m.Costs = make([]uint8, costSize)
		m.RelayPrice = make([]uint8, numRelays)
	}

	if len(m.Costs) > 0 {
		stream.SerializeBytes(m.Costs)
	}

	if m.Version >= 2 && len(m.RelayPrice) > 0 {
		stream.SerializeBytes(m.RelayPrice)
	}

	if stream.IsReading() {
		m.DestRelays = make([]bool, numRelays)
	}
	for i := range m.DestRelays {
		stream.SerializeBool(&m.DestRelays[i])
	}

	return stream.Error()
}

func (m *CostMatrix) Write() ([]byte, error) {
	buffer := make([]byte, m.GetMaxSize())
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

func GenerateRandomCostMatrix(numRelays int) CostMatrix {

	if numRelays > constants.MaxRelays {
		numRelays = constants.MaxRelays
	}

	costMatrix := CostMatrix{
		Version: CostMatrixVersion_Write,
	}

	costMatrix.RelayIds = make([]uint64, numRelays)
	costMatrix.RelayAddresses = make([]net.UDPAddr, numRelays)
	costMatrix.RelayNames = make([]string, numRelays)
	costMatrix.RelayLatitudes = make([]float32, numRelays)
	costMatrix.RelayLongitudes = make([]float32, numRelays)
	costMatrix.RelayDatacenterIds = make([]uint64, numRelays)
	costMatrix.DestRelays = make([]bool, numRelays)

	for i := 0; i < numRelays; i++ {
		costMatrix.RelayIds[i] = rand.Uint64()
		costMatrix.RelayAddresses[i] = RandomAddress()
		costMatrix.RelayNames[i] = RandomString(constants.MaxRelayNameLength)
		costMatrix.RelayLatitudes[i] = rand.Float32()
		costMatrix.RelayLongitudes[i] = rand.Float32()
		costMatrix.RelayDatacenterIds[i] = rand.Uint64()
		costMatrix.DestRelays[i] = RandomBool()
	}

	costSize := core.TriMatrixLength(numRelays)
	costMatrix.Costs = make([]uint8, costSize)
	for i := 0; i < costSize; i++ {
		costMatrix.Costs[i] = uint8(RandomInt(0, 255))
	}

	costMatrix.RelayPrice = make([]uint8, numRelays)
	for i := 0; i < numRelays; i++ {
		costMatrix.RelayPrice[i] = uint8(RandomInt(0,255))
	}

	return costMatrix
}
