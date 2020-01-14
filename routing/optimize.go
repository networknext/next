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
	// IMPORTANT: Bump this version whenever you change the binary format
	CostMatrixVersion = 2

	MaxRelayAddressLength = 256

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

func (m CostMatrix) MarshalBinary() ([]byte, error) {
	var index int

	// todo: update this to the new way of reading/writing binary as per-backend.go

	binary.LittleEndian.PutUint32(buffer[index:], CostMatrixVersion)
	index += 4

	numRelays := len(m.RelayIds)
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelays))
	index += 4

	for i := range m.RelayIds {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(m.RelayIds[i]))
		index += 4
	}

	for i := range m.RelayNames {
		index += WriteString(buffer[index:], m.RelayNames[i])
	}

	if len(m.DatacenterIds) != len(m.DatacenterNames) {
		panic("datacenter ids length does not match datacenter names length")
	}

	binary.LittleEndian.PutUint32(buffer[index:], uint32(len(m.DatacenterIds)))
	index += 4

	for i := 0; i < len(m.DatacenterIds); i++ {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(m.DatacenterIds[i]))
		index += 4
		index += WriteString(buffer[index:], m.DatacenterNames[i])
	}

	for i := range m.RelayAddresses {
		index += WriteBytes(buffer[index:], m.RelayAddresses[i])
	}

	for i := range m.RelayPublicKeys {
		index += WriteBytes(buffer[index:], m.RelayPublicKeys[i])
	}

	numDatacenters := int32(len(m.DatacenterRelays))
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numDatacenters))
	index += 4

	for k, v := range m.DatacenterRelays {

		binary.LittleEndian.PutUint32(buffer[index:], uint32(k))
		index += 4

		numRelaysInDatacenter := len(v)
		binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelaysInDatacenter))
		index += 4

		for i := 0; i < numRelaysInDatacenter; i++ {
			binary.LittleEndian.PutUint32(buffer[index:], uint32(v[i]))
			index += 4
		}
	}

	for i := range m.RTT {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(m.RTT[i]))
		index += 4
	}

	return buffer[:index]
}

// Optimize will fill up a *RouteMatrix with the optimized routes based on cost.
func (m *CostMatrix) Optimize(routes *RouteMatrix, thresholdRTT int32) error {
	numRelays := len(costMatrix.RelayIds)

	entryCount := TriMatrixLength(numRelays)

	result := &RouteMatrix{}
	result.RelayIds = costMatrix.RelayIds
	result.RelayNames = costMatrix.RelayNames
	result.RelayAddresses = costMatrix.RelayAddresses
	result.RelayPublicKeys = costMatrix.RelayPublicKeys
	result.DatacenterIds = costMatrix.DatacenterIds
	result.DatacenterNames = costMatrix.DatacenterNames
	result.DatacenterRelays = costMatrix.DatacenterRelays
	result.Entries = make([]RouteMatrixEntry, entryCount)

	type Indirect struct {
		relay int32
		rtt   int32
	}

	rtt := costMatrix.RTT

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

					ij_index := TriMatrixIndex(i, j)

					numRoutes := 0
					rtt_direct := rtt[ij_index]

					if rtt_direct < 0 {

						// no direct route exists between i,j. subdivide valid routes so we don't miss indirect paths.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ik_index := TriMatrixIndex(i, k)
							kj_index := TriMatrixIndex(k, j)
							ik_rtt := rtt[ik_index]
							kj_rtt := rtt[kj_index]
							if ik_rtt < 0 || kj_rtt < 0 {
								continue
							}
							working[numRoutes].relay = int32(k)
							working[numRoutes].rtt = ik_rtt + kj_rtt
							numRoutes++
						}

					} else {

						// direct route exists between i,j. subdivide only when a significant rtt reduction occurs.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ik_index := TriMatrixIndex(i, k)
							kj_index := TriMatrixIndex(k, j)
							ik_rtt := rtt[ik_index]
							kj_rtt := rtt[kj_index]
							if ik_rtt < 0 || kj_rtt < 0 {
								continue
							}
							indirectRTT := ik_rtt + kj_rtt
							if indirectRTT > rtt_direct-thresholdRTT {
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

					ij_index := TriMatrixIndex(i, j)

					if indirect[i][j] == nil {

						if rtt[ij_index] >= 0 {

							// only direct route from i -> j exists, and it is suitable

							result.Entries[ij_index].DirectRTT = rtt[ij_index]
							result.Entries[ij_index].NumRoutes = 1
							result.Entries[ij_index].RouteRTT[0] = rtt[ij_index]
							result.Entries[ij_index].RouteNumRelays[0] = 2
							result.Entries[ij_index].RouteRelays[0][0] = uint32(i)
							result.Entries[ij_index].RouteRelays[0][1] = uint32(j)

						}

					} else {

						// subdivide routes from i -> j as follows: i -> (x) -> (y) -> (z) -> j, where the subdivision improves significantly on RTT

						routeManager := NewRouteManager()

						for k := range indirect[i][j] {

							routeManager.AddRoute(rtt[ij_index], uint32(i), uint32(j))

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
								ix_index := TriMatrixIndex(i, int(x.relay))
								xy_index := TriMatrixIndex(int(x.relay), int(y.relay))
								yj_index := TriMatrixIndex(int(y.relay), j)

								routeManager.AddRoute(rtt[ix_index]+rtt[xy_index]+rtt[yj_index],
									uint32(i), uint32(x.relay), uint32(y.relay), uint32(j))
							}

							if z != nil {
								iy_index := TriMatrixIndex(i, int(y.relay))
								yz_index := TriMatrixIndex(int(y.relay), int(z.relay))
								zj_index := TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[iy_index]+rtt[yz_index]+rtt[zj_index],
									uint32(i), uint32(y.relay), uint32(z.relay), uint32(j))
							}

							if x != nil && z != nil {
								ix_index := TriMatrixIndex(i, int(x.relay))
								xy_index := TriMatrixIndex(int(x.relay), int(y.relay))
								yz_index := TriMatrixIndex(int(y.relay), int(z.relay))
								zj_index := TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[ix_index]+rtt[xy_index]+rtt[yz_index]+rtt[zj_index],
									uint32(i), uint32(x.relay), uint32(y.relay), uint32(z.relay), uint32(j))
							}

							numRoutes := routeManager.NumRoutes

							result.Entries[ij_index].DirectRTT = rtt[ij_index]
							result.Entries[ij_index].NumRoutes = int32(numRoutes)

							for u := 0; u < numRoutes; u++ {
								result.Entries[ij_index].RouteRTT[u] = routeManager.RouteRTT[u]
								result.Entries[ij_index].RouteNumRelays[u] = routeManager.RouteNumRelays[u]
								numRelays := int(result.Entries[ij_index].RouteNumRelays[u])
								for v := 0; v < numRelays; v++ {
									result.Entries[ij_index].RouteRelays[u][v] = routeManager.RouteRelays[u][v]
								}
							}
						}
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	return result
}

type RouteMatrix struct{}

func (m *RouteMatrix) UnmarshalBinary(data []byte) error {
	return nil
}

func (m RouteMatrix) MarshalBinary() ([]byte, error) {
	return nil, nil
}
