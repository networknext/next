package routing

import (
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/encoding"
)

const (
	// CostMatrixVersion ...
	// IMPORTANT: Bump this version whenever you change the binary format
	CostMatrixVersion = 2

	// RouteMatrixVersion ...
	// IMPORTANT: Increment this when you change the binary format
	RouteMatrixVersion = 2

	// MaxRelays ...
	MaxRelays = 5

	// MaxRoutesPerRelayPair ...
	MaxRoutesPerRelayPair = 8

	/* Duplicated in package: transport */

	// MaxRelayAddressLength ...
	MaxRelayAddressLength = 256

	// LengthOfRelayToken ...
	LengthOfRelayToken = 32
)

type CostMatrix struct {
	RelayIds         []uint64
	RelayNames       []string
	RelayAddresses   [][]byte
	RelayPublicKeys  [][]byte
	DatacenterIds    []uint64
	DatacenterNames  []string
	DatacenterRelays map[uint64][]uint64
	RTT              []int32
}

/* Binary Data Outline for v2: "->" means seqential elements in memory and not another section
 * Version number { uint32 }
 * Number of relays { uint32 }
 * Relay IDs { [NumberOfRelays]uint64 }
 * Relay Names { [NumberOfRelays]string }
 * Number of Datacenters { uint64 }
 * Datacenter ID { [NumberOfDatacenters]uint64 } -> Datacenter Name { [NumberOfDatacenters]string }
 * Relay Addresses { [NumberOfRelays][MaxRelayAddressLength]byte }
 * Relay Public Keys { [NumberOfRelays][LengthOfRelayToken]byte }
 * Number of Datacenters { uint32 }
 * Datacenter ID { uint64 } -> Number of Relays in Datacenter { uint32 } -> Relay IDs in Datacenter { [NumberOfRelaysInDatacenter]uint64 }
 */

// UnmarshalBinary ...
func (m *CostMatrix) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	//version = binary.LittleEndian.Uint32(data[index:])
	//index += 4
	encoding.ReadUint32(data, &index, &version)

	if version > CostMatrixVersion {
		return fmt.Errorf("unknown cost matrix version %d", version)
	}

	numRelays := int32(binary.LittleEndian.Uint32(data[index:]))
	index += 4

	m.RelayIds = make([]uint64, numRelays)
	for i := 0; i < int(numRelays); i++ {
		m.RelayIds[i] = binary.LittleEndian.Uint64(data[index:])
		index += 4
	}

	m.RelayNames = make([]string, numRelays)
	if version >= 1 {
		for i := range m.RelayNames {
			//m.RelayNames[i], bytes_read = ReadString(data[index:])
			//index += bytes_read
			encoding.ReadString(data, &index, &m.RelayNames[i], math.MaxInt32)
		}
	}

	if version >= 2 {
		datacenterCount := binary.LittleEndian.Uint64(data[index:])
		index += 4

		m.DatacenterIds = make([]uint64, datacenterCount)
		m.DatacenterNames = make([]string, datacenterCount)
		for i := 0; i < int(datacenterCount); i++ {
			//m.DatacenterIds[i] = binary.LittleEndian.Uint64(data[index:])
			//index += 4
			encoding.ReadUint64(data, &index, &m.DatacenterIds[i])
			//m.DatacenterNames[i], bytes_read = ReadString(data[index:])
			//index += bytes_read
			encoding.ReadString(data, &index, &m.DatacenterNames[i], math.MaxInt32)
		}
	}

	m.RelayAddresses = make([][]byte, numRelays)
	for i := range m.RelayAddresses {
		//m.RelayAddresses[i], bytes_read = ReadBytes(data[index:])
		//index += bytes_read
		encoding.ReadBytes(data, &index, &m.RelayAddresses[i], MaxRelayAddressLength)
	}

	m.RelayPublicKeys = make([][]byte, numRelays)
	for i := range m.RelayPublicKeys {
		//m.RelayPublicKeys[i], bytes_read = ReadBytes(data[index:])
		//index += bytes_read
		encoding.ReadBytes(data, &index, &m.RelayPublicKeys[i], LengthOfRelayToken)
	}

	numDatacenters := int32(binary.LittleEndian.Uint32(data[index:]))
	index += 4

	m.DatacenterRelays = make(map[uint64][]uint64)

	for i := 0; i < int(numDatacenters); i++ {

		datacenterID := binary.LittleEndian.Uint64(data[index:])
		index += 4

		numRelaysInDatacenter := int32(binary.LittleEndian.Uint32(data[index:]))
		index += 4

		m.DatacenterRelays[datacenterID] = make([]uint64, numRelaysInDatacenter)

		for j := 0; j < int(numRelaysInDatacenter); j++ {
			m.DatacenterRelays[datacenterID][j] = binary.LittleEndian.Uint64(data[index:])
			index += 4
		}
	}

	entryCount := core.TriMatrixLength(int(numRelays))
	m.RTT = make([]int32, entryCount)
	for i := range m.RTT {
		m.RTT[i] = int32(binary.LittleEndian.Uint32(data[index:]))
		index += 4
	}

	return nil
}

// MarshalBinary ...
func (m CostMatrix) MarshalBinary() ([]byte, error) {
	var index int
	data := make([]byte, 0)

	// todo: update this to the new way of reading/writing binary as per-backend.go

	//binary.LittleEndian.PutUint32(buffer[index:], CostMatrixVersion)
	//index += 4

	index = 0
	buff := make([]byte, 4)
	encoding.WriteUint32(buff, &index, CostMatrixVersion)
	data = append(data, buff...)

	//binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelays))
	//index += 4

	numRelays := len(m.RelayIds)
	index = 0
	buff = make([]byte, 4)
	encoding.WriteUint32(buff, &index, uint32(numRelays))
	data = append(data, buff...)

	for _, id := range m.RelayIds {
		//binary.LittleEndian.PutUint32(buffer[index:], uint32(m.RelayIds[i]))
		//index += 4
		index = 0
		buff = make([]byte, 8)
		encoding.WriteUint64(buff, &index, id)
		data = append(data, buff...)
	}

	for _, name := range m.RelayNames {
		//index += WriteString(buffer[index:], m.RelayNames[i])
		index = 0
		buff = make([]byte, len(name)+4)
		encoding.WriteString(buff, &index, name, uint32(len(name)))
		data = append(data, buff...)
	}

	numDatacenters := len(m.DatacenterIds)

	if numDatacenters != len(m.DatacenterNames) {
		return nil, fmt.Errorf("Length of Datacenter IDs not equal to length of Datacenter Names: %d != %d", numDatacenters, len(m.DatacenterIds))
	}

	//binary.LittleEndian.PutUint32(buffer[index:], uint32(len(m.DatacenterIds)))
	//index += 4

	index = 0
	buff = make([]byte, 4)
	encoding.WriteUint32(buff, &index, uint32(numDatacenters))
	data = append(data, buff...)

	for i := 0; i < numDatacenters; i++ {
		//binary.LittleEndian.PutUint32(buffer[index:], uint32(m.DatacenterIds[i]))
		//index += 4
		buff = make([]byte, 8)
		index = 0
		encoding.WriteUint64(buff, &index, m.DatacenterIds[i])
		data = append(data, buff...)

		//index += WriteString(buffer[index:], m.DatacenterNames[i])
		buff = make([]byte, len(m.DatacenterNames[i])+4)
		index = 0
		encoding.WriteString(buff, &index, m.DatacenterNames[i], uint32(len(m.DatacenterNames[i])))
		data = append(data, buff...)
	}

	for i := range m.RelayAddresses {
		//index += WriteBytes(buffer[index:], m.RelayAddresses[i])
		buff = make([]byte, MaxRelayAddressLength)
		index = 0
		encoding.WriteBytes(buff, &index, m.RelayAddresses[i], MaxRelayAddressLength)
		data = append(data, buff...)
	}

	for i := range m.RelayPublicKeys {
		//index += WriteBytes(buffer[index:], m.RelayPublicKeys[i])
		buff = make([]byte, LengthOfRelayToken)
		index = 0
		encoding.WriteBytes(buff, &index, m.RelayPublicKeys[i], LengthOfRelayToken)
		data = append(data, buff...)
	}

	//binary.LittleEndian.PutUint32(buffer[index:], uint32(numDatacenters))
	//index += 4

	numDatacenters = len(m.DatacenterRelays)
	buff = make([]byte, 4)
	index = 0
	encoding.WriteUint32(buff, &index, uint32(numDatacenters))
	data = append(data, buff...)

	for k, v := range m.DatacenterRelays {

		//binary.LittleEndian.PutUint32(buffer[index:], uint32(k))
		//index += 4
		buff = make([]byte, 8)
		index = 0
		encoding.WriteUint64(buff, &index, k)
		data = append(data, buff...)

		//binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelaysInDatacenter))
		//index += 4

		numRelaysInDatacenter := len(v)

		buff = make([]byte, 4)
		index = 0
		encoding.WriteUint32(buff, &index, uint32(numRelaysInDatacenter))
		data = append(data, buff...)

		for i := 0; i < numRelaysInDatacenter; i++ {
			//binary.LittleEndian.PutUint32(buffer[index:], uint32(v[i]))
			//index += 4

			buff = make([]byte, 8)
			index = 0
			encoding.WriteUint64(buff, &index, v[i])
			data = append(data, buff...)
		}
	}

	for i := range m.RTT {
		//binary.LittleEndian.PutUint32(buffer[index:], uint32(m.RTT[i]))
		//index += 4

		buff = make([]byte, 32)
		index = 0
		encoding.WriteUint32(buff, &index, uint32(m.RTT[i]))
		data = append(data, buff...)
	}

	return data, nil
}

// Optimize will fill up a *RouteMatrix with the optimized routes based on cost.
func (m *CostMatrix) Optimize(routes *RouteMatrix, thresholdRTT int32) error {
	numRelays := len(m.RelayIds)

	entryCount := core.TriMatrixLength(numRelays)

	routes.RelayIds = m.RelayIds
	routes.RelayNames = m.RelayNames
	routes.RelayAddresses = m.RelayAddresses
	routes.RelayPublicKeys = m.RelayPublicKeys
	routes.DatacenterIds = m.DatacenterIds
	routes.DatacenterNames = m.DatacenterNames
	routes.DatacenterRelays = m.DatacenterRelays
	routes.Entries = make([]RouteMatrixEntry, entryCount)

	type Indirect struct {
		relay int32
		rtt   int32
	}

	rtt := m.RTT

	indirect := make([][][]Indirect, numRelays)

	// phase 1: build a matrix of indirect routes from relays i -> j that have lower rtt than direct, eg. i -> (x) -> j, where x is every other relay

	numCPUs := runtime.NumCPU()

	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	var wg sync.WaitGroup

	wg.Add(numSegments)

	for segment := 0; segment < numSegments; segment++ {

		startIndex := segment * numRelays / numSegments
		endIndex := (segment+1)*numRelays/numSegments - 1
		if segment == numSegments-1 {
			endIndex = numRelays - 1
		}

		go func(startIndex int, endIndex int) {

			defer wg.Done()

			working := make([]Indirect, numRelays)

			for i := startIndex; i <= endIndex; i++ {

				indirect[i] = make([][]Indirect, numRelays)

				for j := 0; j < numRelays; j++ {

					// can't route to self
					if i == j {
						continue
					}

					ijIndex := core.TriMatrixIndex(i, j)

					numRoutes := 0
					rttDirect := rtt[ijIndex]

					if rttDirect < 0 {

						// no direct route exists between i,j. subdivide valid routes so we don't miss indirect paths.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ikIndex := core.TriMatrixIndex(i, k)
							kjIndex := core.TriMatrixIndex(k, j)
							ikRtt := rtt[ikIndex]
							kjRtt := rtt[kjIndex]
							if ikRtt < 0 || kjRtt < 0 {
								continue
							}
							working[numRoutes].relay = int32(k)
							working[numRoutes].rtt = ikRtt + kjRtt
							numRoutes++
						}

					} else {

						// direct route exists between i,j. subdivide only when a significant rtt reduction occurs.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ikIndex := core.TriMatrixIndex(i, k)
							kjIndex := core.TriMatrixIndex(k, j)
							ikRtt := rtt[ikIndex]
							kjRtt := rtt[kjIndex]
							if ikRtt < 0 || kjRtt < 0 {
								continue
							}
							indirectRTT := ikRtt + kjRtt
							if indirectRTT > rttDirect-thresholdRTT {
								continue
							}
							working[numRoutes].relay = int32(k)
							working[numRoutes].rtt = indirectRTT
							numRoutes++
						}

					}

					if numRoutes > 0 {
						indirect[i][j] = make([]Indirect, numRoutes)
						copy(indirect[i][j], working)
						sort.Slice(indirect[i][j], func(a, b int) bool { return indirect[i][j][a].rtt < indirect[i][j][b].rtt })
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	// phase 2: use the indirect matrix to subdivide a route up to 5 hops

	wg.Add(numSegments)

	for segment := 0; segment < numSegments; segment++ {

		startIndex := segment * numRelays / numSegments
		endIndex := (segment+1)*numRelays/numSegments - 1
		if segment == numSegments-1 {
			endIndex = numRelays - 1
		}

		go func(startIndex int, endIndex int) {

			defer wg.Done()

			for i := startIndex; i <= endIndex; i++ {

				for j := 0; j < i; j++ {

					ijIndex := core.TriMatrixIndex(i, j)

					if indirect[i][j] == nil {

						if rtt[ijIndex] >= 0 {

							// only direct route from i -> j exists, and it is suitable

							routes.Entries[ijIndex].DirectRTT = rtt[ijIndex]
							routes.Entries[ijIndex].NumRoutes = 1
							routes.Entries[ijIndex].RouteRTT[0] = rtt[ijIndex]
							routes.Entries[ijIndex].RouteNumRelays[0] = 2
							routes.Entries[ijIndex].RouteRelays[0][0] = uint32(i)
							routes.Entries[ijIndex].RouteRelays[0][1] = uint32(j)

						}

					} else {

						// subdivide routes from i -> j as follows: i -> (x) -> (y) -> (z) -> j, where the subdivision improves significantly on RTT

						routeManager := core.NewRouteManager()

						for k := range indirect[i][j] {

							routeManager.AddRoute(rtt[ijIndex], uint32(i), uint32(j))

							y := indirect[i][j][k]

							routeManager.AddRoute(y.rtt, uint32(i), uint32(y.relay), uint32(j))

							var x *Indirect
							if indirect[i][y.relay] != nil {
								x = &indirect[i][y.relay][0]
							}

							var z *Indirect
							if indirect[j][y.relay] != nil {
								z = &indirect[j][y.relay][0]
							}

							if x != nil {
								ixIndex := core.TriMatrixIndex(i, int(x.relay))
								xyIndex := core.TriMatrixIndex(int(x.relay), int(y.relay))
								yjIndex := core.TriMatrixIndex(int(y.relay), j)

								routeManager.AddRoute(rtt[ixIndex]+rtt[xyIndex]+rtt[yjIndex],
									uint32(i), uint32(x.relay), uint32(y.relay), uint32(j))
							}

							if z != nil {
								iyIndex := core.TriMatrixIndex(i, int(y.relay))
								yzIndex := core.TriMatrixIndex(int(y.relay), int(z.relay))
								zjIndex := core.TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[iyIndex]+rtt[yzIndex]+rtt[zjIndex],
									uint32(i), uint32(y.relay), uint32(z.relay), uint32(j))
							}

							if x != nil && z != nil {
								ixIndex := core.TriMatrixIndex(i, int(x.relay))
								xyIndex := core.TriMatrixIndex(int(x.relay), int(y.relay))
								yzIndex := core.TriMatrixIndex(int(y.relay), int(z.relay))
								zjIndex := core.TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[ixIndex]+rtt[xyIndex]+rtt[yzIndex]+rtt[zjIndex],
									uint32(i), uint32(x.relay), uint32(y.relay), uint32(z.relay), uint32(j))
							}

							numRoutes := routeManager.NumRoutes

							routes.Entries[ijIndex].DirectRTT = rtt[ijIndex]
							routes.Entries[ijIndex].NumRoutes = int32(numRoutes)

							for u := 0; u < numRoutes; u++ {
								routes.Entries[ijIndex].RouteRTT[u] = routeManager.RouteRTT[u]
								routes.Entries[ijIndex].RouteNumRelays[u] = routeManager.RouteNumRelays[u]
								numRelays := int(routes.Entries[ijIndex].RouteNumRelays[u])
								for v := 0; v < numRelays; v++ {
									routes.Entries[ijIndex].RouteRelays[u][v] = routeManager.RouteRelays[u][v]
								}
							}
						}
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	return nil
}

// RouteMatrixEntry ...
type RouteMatrixEntry struct {
	DirectRTT      int32
	NumRoutes      int32
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint32
}

// RouteMatrix ...
type RouteMatrix struct {
	RelayIds         []uint64
	RelayNames       []string
	RelayAddresses   [][]byte
	RelayPublicKeys  [][]byte
	DatacenterRelays map[uint64][]uint64
	DatacenterIds    []uint64
	DatacenterNames  []string
	Entries          []RouteMatrixEntry
}

// UnmarshalBinary ...
func (m *RouteMatrix) UnmarshalBinary(data []byte) error {
	return nil
}

// MarshalBinary ...
func (m RouteMatrix) MarshalBinary() ([]byte, error) {
	return nil, nil
}
