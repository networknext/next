package routing

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"

	"github.com/networknext/backend/encoding"
)

type CostMatrix struct {
	RelayIDs           []uint64
	RelayAddresses     []net.UDPAddr
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIDs []uint64
	Costs              []int32

	cachedResponse      []byte
	cachedResponseMutex sync.RWMutex
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
