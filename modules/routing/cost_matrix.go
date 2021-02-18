package routing

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"

	"github.com/networknext/backend/modules/encoding"
)

const CostMatrixV1 = 1

type CostMatrix struct {
	Version            uint32
	CreatedAt          uint64
	RelayIDs           []uint64
	RelayAddresses     []net.UDPAddr
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIDs []uint64
	Costs              []int32

	cachedResponse       []byte
	cachedResponseBinary []byte
	cachedResponseMutex  sync.RWMutex
}

func (m *CostMatrix) Serialize(stream encoding.Stream) error {
	numRelays := uint32(len(m.RelayIDs))
	stream.SerializeUint32(&numRelays)

	if stream.IsReading() {
		m.RelayIDs = make([]uint64, numRelays)
		m.RelayAddresses = make([]net.UDPAddr, numRelays)
		m.RelayNames = make([]string, numRelays)
		m.RelayLatitudes = make([]float32, numRelays)
		m.RelayLongitudes = make([]float32, numRelays)
		m.RelayDatacenterIDs = make([]uint64, numRelays)
	}

	for i := uint32(0); i < numRelays; i++ {
		stream.SerializeUint64(&m.RelayIDs[i])
		stream.SerializeAddress(&m.RelayAddresses[i])
		stream.SerializeString(&m.RelayNames[i], MaxRelayNameLength)
		stream.SerializeFloat32(&m.RelayLatitudes[i])
		stream.SerializeFloat32(&m.RelayLongitudes[i])
		stream.SerializeUint64(&m.RelayDatacenterIDs[i])
	}

	costsLength := uint32(len(m.Costs))
	stream.SerializeUint32(&costsLength)
	if stream.IsReading() {
		m.Costs = make([]int32, costsLength)
	}

	for i := uint32(0); i < costsLength; i++ {
		stream.SerializeInteger(&m.Costs[i], -1, InvalidRouteValue)
	}

	return stream.Error()
}

func (m *CostMatrix) encodeBinaryV1(data []byte) error {
	index := 0
	encoding.WriteUint32(data, &index, CostMatrixV1)
	encoding.WriteUint64(data, &index, m.CreatedAt)

	numRelays := uint16(len(m.RelayIDs))
	encoding.WriteUint16(data, &index, numRelays)

	for i := uint16(0); i < numRelays; i++ {
		encoding.WriteUint64(data, &index, m.RelayIDs[i])
		addrBytes := make([]byte, 20)
		encoding.WriteAddress(addrBytes, &m.RelayAddresses[i])
		encoding.WriteUint8(data, &index, uint8(len(addrBytes)))
		encoding.WriteBytes(data, &index, addrBytes, len(addrBytes))
		encoding.WriteString(data, &index, m.RelayNames[i], MaxRelayNameLength)
		encoding.WriteFloat32(data, &index, m.RelayLatitudes[i])
		encoding.WriteFloat32(data, &index, m.RelayLongitudes[i])
		encoding.WriteUint64(data, &index, m.RelayDatacenterIDs[i])
	}

	costsLength := uint32(len(m.Costs))
	encoding.WriteUint32(data, &index, costsLength)

	for i := uint32(0); i < costsLength; i++ {
		if m.Costs[i] > InvalidRouteValue {
			return fmt.Errorf("cost greater than Invalid Route Value")
		}
		encoding.WriteInt32(data, &index, m.Costs[i])
	}

	return nil
}

func (m *CostMatrix) decodeBinaryV1(data []byte, index int) error {

	encoding.ReadUint64(data, &index, &m.CreatedAt)

	var numRelays uint16
	if !encoding.ReadUint16(data, &index, &numRelays) {
		return fmt.Errorf("unable to decode number of relays")
	}

	m.RelayIDs = make([]uint64, numRelays)
	m.RelayAddresses = make([]net.UDPAddr, numRelays)
	m.RelayNames = make([]string, numRelays)
	m.RelayLatitudes = make([]float32, numRelays)
	m.RelayLongitudes = make([]float32, numRelays)
	m.RelayDatacenterIDs = make([]uint64, numRelays)

	for i := uint16(0); i < numRelays; i++ {
		if !encoding.ReadUint64(data, &index, &m.RelayIDs[i]) {
			return fmt.Errorf("unable to decode relay ID")
		}
		var addrSize uint8
		if !encoding.ReadUint8(data, &index, &addrSize) {
			return fmt.Errorf("unable to decode relay address size")
		}

		addrBytes := make([]byte, addrSize)
		if !encoding.ReadBytes(data, &index, &addrBytes, uint32(addrSize)) {
			return fmt.Errorf("unable to retrieve relay address bytes from data")
		}
		m.RelayAddresses[i] = *encoding.ReadAddress(addrBytes)

		if !encoding.ReadString(data, &index, &m.RelayNames[i], MaxRelayNameLength) {
			return fmt.Errorf("unable to decode relay name")
		}

		if !encoding.ReadFloat32(data, &index, &m.RelayLatitudes[i]) {
			return fmt.Errorf("unable to decode relay latitude")
		}

		if !encoding.ReadFloat32(data, &index, &m.RelayLongitudes[i]) {
			return fmt.Errorf("unable to decode relay longitude")
		}

		if !encoding.ReadUint64(data, &index, &m.RelayDatacenterIDs[i]) {
			return fmt.Errorf("unable to decode relay datacenter ID")
		}
	}

	var costsLength uint32
	if !encoding.ReadUint32(data, &index, &costsLength) {
		return fmt.Errorf("unable to decode costs array length")
	}
	m.Costs = make([]int32, costsLength)

	for i := uint32(0); i < costsLength; i++ {
		if !encoding.ReadInt32(data, &index, &m.Costs[i]) {
			return fmt.Errorf("unable to decode costs")
		}
		if m.Costs[i] > int32(InvalidRouteValue) {
			return fmt.Errorf("cost value greater than Invalid Route Value")
		}
	}
	return nil
}

func (m *CostMatrix) ReadFromBinary(data []byte) error {
	index := 0
	encoding.ReadUint32(data, &index, &m.Version)

	switch m.Version {
	case CostMatrixV1:
		return m.decodeBinaryV1(data, index)
	default:
		return fmt.Errorf("unsuported Cost Matrix version: %v", m.Version)
	}
}

func (m *CostMatrix) WriteToBinary(data []byte) error {
	switch m.Version {
	case CostMatrixV1:
		return m.encodeBinaryV1(data)
	default:
		return fmt.Errorf("unsuported Cost Matrix version: %v", m.Version)
	}
}

func (m *CostMatrix) ReadFrom(reader io.Reader) (int64, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, err
	}

	readStream := encoding.CreateReadStream(data)
	err = m.Serialize(readStream)
	return int64(readStream.GetBytesProcessed()), err
}

func (m *CostMatrix) GetResponseData() []byte {
	m.cachedResponseMutex.RLock()
	response := m.cachedResponse
	m.cachedResponseMutex.RUnlock()
	return response
}

func (m *CostMatrix) GetResponseDataBinary() []byte {
	m.cachedResponseMutex.RLock()
	response := m.cachedResponseBinary
	m.cachedResponseMutex.RUnlock()
	return response
}

func (m *CostMatrix) WriteResponseData(bufferSize int) error {
	ws, err := encoding.CreateWriteStream(bufferSize)
	if err != nil {
		return fmt.Errorf("failed to create write stream in cost matrix WriteResponseData(): %v", err)
	}

	if err := m.Serialize(ws); err != nil {
		return fmt.Errorf("failed to serialize cost matrix in WriteResponseData(): %v", err)
	}

	ws.Flush()

	m.cachedResponseMutex.Lock()
	m.cachedResponse = ws.GetData()[:ws.GetBytesProcessed()]
	m.cachedResponseMutex.Unlock()
	return nil
}

func (m *CostMatrix) WriteResponseDataBinary(bufferSize int) error {
	data := make([]byte, bufferSize)
	if err := m.WriteToBinary(data); err != nil {
		return fmt.Errorf("failed to encode Cost Matrix in WriteResponseData(): %v", err)
	}

	m.cachedResponseMutex.Lock()
	m.cachedResponseBinary = data
	m.cachedResponseMutex.Unlock()
	return nil
}
