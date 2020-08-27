package routing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	// IMPORTANT: Bump this version whenever you change the binary format
	CostMatrixVersion = 6
)

type CostMatrix struct {
	mu sync.RWMutex

	RelayIndices map[uint64]int

	RelayIDs              []uint64
	RelayNames            []string
	RelayAddresses        [][]byte
	RelayLatitude         []float64
	RelayLongitude        []float64
	RelayPublicKeys       [][]byte
	DatacenterIDs         []uint64
	DatacenterNames       []string
	DatacenterRelays      map[uint64][]uint64
	RTT                   []int32
	RelaySellers          []Seller
	RelaySessionCounts    []uint32
	RelayMaxSessionCounts []uint32

	responseBuffer     []byte
	reponseBufferMutex sync.RWMutex
}

// implements the io.ReadFrom interface
func (m *CostMatrix) ReadFrom(r io.Reader) (int64, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}

	if err := m.UnmarshalBinary(data); err != nil {
		return 0, err
	}

	return int64(len(data)), nil
}

// implements the io.WriteTo interface
func (m *CostMatrix) WriteTo(w io.Writer) (int64, error) {
	data, err := m.MarshalBinary()
	if err != nil {
		return 0, err
	}

	n, err := w.Write(data)
	if err != nil {
		return 0, err
	}

	return int64(n), nil
}

func (m *CostMatrix) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(data, &index, &version) {
		return errors.New("[CostMatrix] invalid read at version number")
	}

	if version > CostMatrixVersion {
		return fmt.Errorf("unknown cost matrix version %d", version)
	}

	var idReadFunc func([]byte, *int, *uint64, string) error
	var bytesReadFunc func([]byte, *int, *[]byte, uint32, string) error

	if version >= 3 {
		idReadFunc = readIDNew
		bytesReadFunc = readBytesNew
	} else {
		idReadFunc = readIDOld
		bytesReadFunc = readBytesOld
	}

	var numRelays uint32
	if !encoding.ReadUint32(data, &index, &numRelays) {
		return errors.New("[CostMatrix] invalid read at number of relays")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.RelayIndices = make(map[uint64]int)
	m.RelayIDs = make([]uint64, numRelays)

	for i := 0; i < int(numRelays); i++ {
		var tmp uint64
		if err := idReadFunc(data, &index, &tmp, "[CostMatrix] invalid read at relay ids"); err != nil {
			return err
		}
		m.RelayIndices[tmp] = i
		m.RelayIDs[i] = tmp
	}

	m.RelayNames = make([]string, numRelays)
	if version >= 1 {
		for i := range m.RelayNames {
			if !encoding.ReadString(data, &index, &m.RelayNames[i], math.MaxInt32) {
				return errors.New("[CostMatrix] invalid read at relay names")
			}
		}
	}

	if version >= 2 {
		var datacenterCount uint32
		if !encoding.ReadUint32(data, &index, &datacenterCount) {
			return errors.New("[CostMatrix] invalid read at datacenter count")
		}

		m.DatacenterIDs = make([]uint64, datacenterCount)
		m.DatacenterNames = make([]string, datacenterCount)

		for i := 0; i < int(datacenterCount); i++ {
			if err := idReadFunc(data, &index, &m.DatacenterIDs[i], "[CostMatrix] invalid read at datacenter ids"); err != nil {
				return err
			}

			if !encoding.ReadString(data, &index, &m.DatacenterNames[i], math.MaxInt32) {
				return errors.New("[CostMatrix] invalid read at datacenter names")
			}
		}
	}

	m.RelayAddresses = make([][]byte, numRelays)
	for i := range m.RelayAddresses {
		if err := bytesReadFunc(data, &index, &m.RelayAddresses[i], MaxRelayAddressLength, "[CostMatrix] invalid read at relay addresses"); err != nil {
			return err
		}
	}

	m.RelayLatitude = make([]float64, numRelays)
	m.RelayLongitude = make([]float64, numRelays)

	if version >= 6 {

		for i := range m.RelayLatitude {
			if !encoding.ReadFloat64(data, &index, &m.RelayLatitude[i]) {
				return errors.New("[CostMatrix] invalid read at relay latitude")
			}
		}

		for i := range m.RelayLongitude {
			if !encoding.ReadFloat64(data, &index, &m.RelayLongitude[i]) {
				return errors.New("[CostMatrix] invalid read at relay longitude")
			}
		}
	}

	m.RelayPublicKeys = make([][]byte, numRelays)
	for i := range m.RelayPublicKeys {
		if err := bytesReadFunc(data, &index, &m.RelayPublicKeys[i], crypto.KeySize, "[CostMatrix] invalid read at relay public keys"); err != nil {
			return err
		}
	}

	var numDatacenters uint32
	if !encoding.ReadUint32(data, &index, &numDatacenters) {
		return errors.New("[CostMatrix] invalid read at number of datacenters (second time)")
	}

	m.DatacenterRelays = make(map[uint64][]uint64)

	for i := 0; i < int(numDatacenters); i++ {
		var datacenterID uint64

		if err := idReadFunc(data, &index, &datacenterID, "[CostMatrix] invalid read at datacenter id"); err != nil {
			return err
		}

		var numRelaysInDatacenter uint32
		if !encoding.ReadUint32(data, &index, &numRelaysInDatacenter) {
			return errors.New("[CostMatrix] invalid read at number of relays in datacenter")
		}

		m.DatacenterRelays[datacenterID] = make([]uint64, numRelaysInDatacenter)

		for j := 0; j < int(numRelaysInDatacenter); j++ {
			if err := idReadFunc(data, &index, &m.DatacenterRelays[datacenterID][j], "[CostMatrix] invalid read at relay ids for datacenter"); err != nil {
				return err
			}
		}
	}

	entryCount := TriMatrixLength(int(numRelays))
	m.RTT = make([]int32, entryCount)

	for i := range m.RTT {
		var tmp uint32
		// read rtt entries
		if !encoding.ReadUint32(data, &index, &tmp) {
			return errors.New("[CostMatrix] invalid read at rtt")
		}
		m.RTT[i] = int32(tmp)
	}

	m.RelaySellers = make([]Seller, numRelays)
	if version >= 4 {
		for i := range m.RelaySellers {
			if !encoding.ReadString(data, &index, &m.RelaySellers[i].ID, math.MaxInt32) {
				return errors.New("[CostMatrix] invalid read on relay seller ID")
			}
			if !encoding.ReadString(data, &index, &m.RelaySellers[i].Name, math.MaxInt32) {
				return errors.New("[CostMatrix] invalid read on relay seller name")
			}

			var ingressNibblins uint64
			if !encoding.ReadUint64(data, &index, &ingressNibblins) {
				return errors.New("[CostMatrix] invalid read on relay seller ingress price")
			}
			m.RelaySellers[i].IngressPriceNibblinsPerGB = Nibblin(ingressNibblins)

			var egressNibblins uint64
			if !encoding.ReadUint64(data, &index, &egressNibblins) {
				return errors.New("[CostMatrix] invalid read on relay seller egress price")
			}
			m.RelaySellers[i].EgressPriceNibblinsPerGB = Nibblin(egressNibblins)
		}
	}

	m.RelaySessionCounts = make([]uint32, numRelays)
	if version >= 5 {
		for i := range m.RelaySessionCounts {
			if !encoding.ReadUint32(data, &index, &m.RelaySessionCounts[i]) {
				return errors.New("[CostMatrix] invalid read on relay session count")
			}
		}
	}

	m.RelayMaxSessionCounts = make([]uint32, numRelays)
	if version >= 5 {
		for i := range m.RelayMaxSessionCounts {
			if !encoding.ReadUint32(data, &index, &m.RelayMaxSessionCounts[i]) {
				return errors.New("[CostMatrix] invalid read on relay max session count")
			}
		}
	}

	return nil
}

func (m *CostMatrix) MarshalBinary() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	index := 0
	data := make([]byte, m.Size())

	encoding.WriteUint32(data, &index, CostMatrixVersion)

	numRelays := len(m.RelayIDs)

	if numRelays != len(m.RelayNames) {
		return nil, fmt.Errorf("length of Relay IDs not equal to length of Relay Names: %d != %d", numRelays, len(m.RelayNames))
	}

	encoding.WriteUint32(data, &index, uint32(numRelays))

	for _, id := range m.RelayIDs {
		encoding.WriteUint64(data, &index, id)
	}

	for _, name := range m.RelayNames {
		encoding.WriteString(data, &index, name, uint32(len(name)))
	}

	numDatacenters := len(m.DatacenterIDs)

	if numDatacenters != len(m.DatacenterNames) {
		return nil, fmt.Errorf("length of Datacenter IDs not equal to length of Datacenter Names: %d != %d", numDatacenters, len(m.DatacenterNames))
	}

	encoding.WriteUint32(data, &index, uint32(numDatacenters))

	for i := 0; i < numDatacenters; i++ {
		encoding.WriteUint64(data, &index, m.DatacenterIDs[i])
		encoding.WriteString(data, &index, m.DatacenterNames[i], uint32(len(m.DatacenterNames[i])))
	}

	for i := range m.RelayAddresses {
		tmp := make([]byte, MaxRelayAddressLength)
		copy(tmp, m.RelayAddresses[i])
		encoding.WriteBytes(data, &index, tmp, MaxRelayAddressLength)
	}

	if RouteMatrixVersion >= 6 {

		if len(m.RelayLatitude) != numRelays {
			return nil, fmt.Errorf("bad relay latitude array length")
		}

		for i := range m.RelayLatitude {
			encoding.WriteFloat64(data, &index, m.RelayLatitude[i])
		}

		if len(m.RelayLongitude) != numRelays {
			return nil, fmt.Errorf("bad relay longitude array length")
		}

		for i := range m.RelayLongitude {
			encoding.WriteFloat64(data, &index, m.RelayLongitude[i])
		}

	}

	for i := range m.RelayPublicKeys {
		tmp := make([]byte, MaxRelayAddressLength)
		copy(tmp, m.RelayPublicKeys[i])
		encoding.WriteBytes(data, &index, tmp, crypto.KeySize)
	}

	numDatacenters = len(m.DatacenterRelays)

	encoding.WriteUint32(data, &index, uint32(numDatacenters))

	for k, v := range m.DatacenterRelays {

		encoding.WriteUint64(data, &index, k)

		numRelaysInDatacenter := len(v)

		encoding.WriteUint32(data, &index, uint32(numRelaysInDatacenter))

		for i := 0; i < numRelaysInDatacenter; i++ {
			encoding.WriteUint64(data, &index, v[i])
		}
	}

	for i := range m.RTT {
		encoding.WriteUint32(data, &index, uint32(m.RTT[i]))
	}

	for _, seller := range m.RelaySellers {
		encoding.WriteString(data, &index, seller.ID, uint32(len(seller.ID)))
		encoding.WriteString(data, &index, seller.Name, uint32(len(seller.Name)))
		encoding.WriteUint64(data, &index, uint64(seller.IngressPriceNibblinsPerGB))
		encoding.WriteUint64(data, &index, uint64(seller.EgressPriceNibblinsPerGB))
	}

	for i := range m.RelaySessionCounts {
		encoding.WriteUint32(data, &index, m.RelaySessionCounts[i])
	}

	for i := range m.RelayMaxSessionCounts {
		encoding.WriteUint32(data, &index, m.RelayMaxSessionCounts[i])
	}

	return data, nil
}

func (m *CostMatrix) Optimize(routes *RouteMatrix, thresholdRTT int32) error {

	m.mu.RLock()
	defer func() {
		m.mu.RUnlock()
	}()

	numRelays := len(m.RelayIDs)

	entryCount := TriMatrixLength(numRelays)

	routes.RelayIndices = m.RelayIndices
	routes.RelayIDs = m.RelayIDs
	routes.RelayNames = m.RelayNames
	routes.RelayAddresses = m.RelayAddresses
	routes.RelayLatitude = m.RelayLatitude
	routes.RelayLongitude = m.RelayLongitude
	routes.RelayPublicKeys = m.RelayPublicKeys
	routes.DatacenterIDs = m.DatacenterIDs
	routes.DatacenterNames = m.DatacenterNames
	routes.DatacenterRelays = m.DatacenterRelays
	routes.Entries = make([]RouteMatrixEntry, entryCount)
	routes.RelaySellers = m.RelaySellers
	routes.RelaySessionCounts = m.RelaySessionCounts
	routes.RelayMaxSessionCounts = m.RelayMaxSessionCounts

	if err := routes.UpdateRelayAddressCache(); err != nil {
		return err
	}

	routes.UpdateRouteCache()

	type Indirect struct {
		relay uint64
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

					ijIndex := TriMatrixIndex(i, j)

					numRoutes := 0
					rttDirect := rtt[ijIndex]

					if rttDirect < 0 {

						// no direct route exists between i,j. subdivide valid routes so we don't miss indirect paths.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ikIndex := TriMatrixIndex(i, k)
							kjIndex := TriMatrixIndex(k, j)
							ikRTT := rtt[ikIndex]
							kjRTT := rtt[kjIndex]
							if ikRTT < 0 || kjRTT < 0 {
								continue
							}
							working[numRoutes].relay = uint64(k)
							working[numRoutes].rtt = ikRTT + kjRTT
							numRoutes++
						}

					} else {

						// direct route exists between i,j. subdivide only when a significant rtt reduction occurs.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ikIndex := TriMatrixIndex(i, k)
							kjIndex := TriMatrixIndex(k, j)
							ikRTT := rtt[ikIndex]
							kjRTT := rtt[kjIndex]
							if ikRTT < 0 || kjRTT < 0 {
								continue
							}
							indirectRTT := ikRTT + kjRTT
							if indirectRTT > rttDirect-thresholdRTT {
								continue
							}
							working[numRoutes].relay = uint64(k)
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

					ijIndex := TriMatrixIndex(i, j)

					if indirect[i][j] == nil {

						if rtt[ijIndex] >= 0 {

							// only direct route from i -> j exists, and it is suitable

							routes.Entries[ijIndex].DirectRTT = rtt[ijIndex]
							routes.Entries[ijIndex].NumRoutes = 1
							routes.Entries[ijIndex].RouteRTT[0] = rtt[ijIndex]
							routes.Entries[ijIndex].RouteNumRelays[0] = 2
							routes.Entries[ijIndex].RouteRelays[0][0] = uint64(i)
							routes.Entries[ijIndex].RouteRelays[0][1] = uint64(j)

						}

					} else {

						// subdivide routes from i -> j as follows: i -> (x) -> (y) -> (z) -> j, where the subdivision improves significantly on RTT

						var routeManager RouteManager

						for k := range indirect[i][j] {

							routeManager.AddRoute(rtt[ijIndex], uint64(i), uint64(j))

							y := indirect[i][j][k]

							routeManager.AddRoute(y.rtt, uint64(i), y.relay, uint64(j))

							var x *Indirect
							if indirect[i][y.relay] != nil {
								x = &indirect[i][y.relay][0]
							}

							var z *Indirect
							if indirect[j][y.relay] != nil {
								z = &indirect[j][y.relay][0]
							}

							if x != nil {
								ixIndex := TriMatrixIndex(i, int(x.relay))
								xyIndex := TriMatrixIndex(int(x.relay), int(y.relay))
								yjIndex := TriMatrixIndex(int(y.relay), j)

								routeManager.AddRoute(rtt[ixIndex]+rtt[xyIndex]+rtt[yjIndex],
									uint64(i), x.relay, y.relay, uint64(j))
							}

							if z != nil {
								iyIndex := TriMatrixIndex(i, int(y.relay))
								yzIndex := TriMatrixIndex(int(y.relay), int(z.relay))
								zjIndex := TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[iyIndex]+rtt[yzIndex]+rtt[zjIndex],
									uint64(i), y.relay, z.relay, uint64(j))
							}

							if x != nil && z != nil {
								ixIndex := TriMatrixIndex(i, int(x.relay))
								xyIndex := TriMatrixIndex(int(x.relay), int(y.relay))
								yzIndex := TriMatrixIndex(int(y.relay), int(z.relay))
								zjIndex := TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[ixIndex]+rtt[xyIndex]+rtt[yzIndex]+rtt[zjIndex],
									uint64(i), x.relay, y.relay, z.relay, uint64(j))
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

func (m *CostMatrix) Size() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var length uint64
	numRelays := uint64(len(m.RelayIDs))
	numDatacenters := uint64(len(m.DatacenterIDs))

	// uint32 version number + number of relays + allocation for all relay ids
	length = 4 + 4 + 8*numRelays

	for _, name := range m.RelayNames {
		// length of relay name + allocation for relay name
		length += uint64(4 + len(name))
	}

	// number of datacenters + allocation for datacenter ids
	length += 8 + 8*numDatacenters

	for _, name := range m.DatacenterNames {
		// length of datacenter name + allocation for datacenter name
		length += uint64(4 + len(name))
	}

	// allocation for relay addresses + allocation for relay public keys + the No. of datacenters, duplication?
	length += numRelays*uint64(MaxRelayAddressLength+crypto.KeySize) + 4

	// allocation for relay lat and longs
	length += numRelays * 8 * 2

	for _, v := range m.DatacenterRelays {
		// datacenter id + number of relays for that datacenter + allocation for all of those relay ids
		length += uint64(8 + 4 + 8*len(v))
	}

	// length so far + number of rtt entries
	length += uint64(4 * len(m.RTT))

	// Add length of relay sellers
	for _, seller := range m.RelaySellers {
		length += uint64(4 + len(seller.ID) + 4 + len(seller.Name) + 8 + 8)
	}

	// Add length of relay session counts
	length += uint64(len(m.RelaySessionCounts) * 4)

	// Add length of relay max session counts
	length += uint64(len(m.RelayMaxSessionCounts) * 4)

	return length
}

func (m *CostMatrix) GetResponseData() []byte {
	m.reponseBufferMutex.RLock()
	data := m.responseBuffer
	m.reponseBufferMutex.RUnlock()
	return data
}

func (m *CostMatrix) WriteResponseData() error {
	var buffer bytes.Buffer
	if _, err := m.WriteTo(&buffer); err != nil {
		return err
	}
	m.reponseBufferMutex.Lock()
	m.responseBuffer = buffer.Bytes()
	m.reponseBufferMutex.Unlock()
	return nil
}
