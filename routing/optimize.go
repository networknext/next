package routing

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"runtime"
	"sort"
	"sync"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	// CostMatrixVersion ...
	// IMPORTANT: Bump this version whenever you change the binary format
	CostMatrixVersion = 3

	// RouteMatrixVersion ...
	// IMPORTANT: Increment this when you change the binary format
	RouteMatrixVersion = 3

	// MaxRelays ...
	MaxRelays = 5

	// MaxRoutesPerRelayPair ...
	MaxRoutesPerRelayPair = 8

	/* Duplicated in package: transport */

	// MaxRelayAddressLength ...
	MaxRelayAddressLength = 256
)

func readIDOld(data []byte, index *int, storage *uint64, errmsg string) error {
	var tmp uint32
	if !encoding.ReadUint32(data, index, &tmp) {
		return errors.New(errmsg + " - ver < 3")
	}
	*storage = uint64(tmp)
	return nil
}

func readIDNew(data []byte, index *int, storage *uint64, errmsg string) error {
	if !encoding.ReadUint64(data, index, storage) {
		return errors.New(errmsg + " - v3")
	}
	return nil
}

func readBytesOld(data []byte, index *int, storage *[]byte, length uint32, errmsg string) error {
	var bytesRead int
	*storage, bytesRead = encoding.ReadBytesOld(data[*index:])
	*index += bytesRead
	return nil
}

func readBytesNew(data []byte, index *int, storage *[]byte, length uint32, errmsg string) error {
	if !encoding.ReadBytes(data, index, storage, length) {
		return errors.New(errmsg + " - v3")
	}
	return nil
}

// CostMatrix ...
type CostMatrix struct {
	mu sync.Mutex

	RelayIds         []uint64
	RelayNames       []string
	RelayAddresses   [][]byte
	RelayPublicKeys  [][]byte
	DatacenterIds    []uint64
	DatacenterNames  []string
	DatacenterRelays map[uint64][]uint64
	RTT              []int32
}

// ReadFrom implements the io.ReadFrom interface
func (m *CostMatrix) ReadFom(r io.Reader) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}

	if err := m.UnmarshalBinary(data); err != nil {
		return 0, err
	}

	return int64(len(data)), nil
}

// WriteTo implements the io.WriteTo interface
func (m *CostMatrix) WriteTo(w io.Writer) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *CostMatrix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	m.WriteTo(w)
}

/* Binary data outline for CostMatrix v2: "->" means seqential elements in memory and not another section
 * Version number { uint32 }
 * Number of relays { uint32 }
 * Relay IDs { [NumberOfRelays]uint64 }
 * Relay Names { [NumberOfRelays]string }
 * Number of Datacenters { uint32 }
 * Datacenter ID { [NumberOfDatacenters]uint64 } -> Datacenter Name { [NumberOfDatacenters]string }
 * Relay Addresses { [NumberOfRelays][MaxRelayAddressLength]byte }
 * Relay Public Keys { [NumberOfRelays][crypto.KeySize]byte }
 * Number of Datacenters { uint32 }
 * Datacenter ID { uint64 } -> Number of Relays in Datacenter { uint32 } -> Relay IDs in Datacenter { [NumberOfRelaysInDatacenter]uint64 }
 * RTT Info { []uint32 }
 */

// UnmarshalBinary ...
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

	m.RelayIds = make([]uint64, numRelays)

	for i := 0; i < int(numRelays); i++ {
		if err := idReadFunc(data, &index, &m.RelayIds[i], "[CostMatrix] invalid read at relay ids"); err != nil {
			return err
		}
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

		m.DatacenterIds = make([]uint64, datacenterCount)
		m.DatacenterNames = make([]string, datacenterCount)

		for i := 0; i < int(datacenterCount); i++ {
			if err := idReadFunc(data, &index, &m.DatacenterIds[i], "[CostMatrix] invalid read at datacenter ids"); err != nil {
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

	return nil
}

// MarshalBinary ...
func (m CostMatrix) MarshalBinary() ([]byte, error) {
	index := 0
	data := make([]byte, m.Size())

	encoding.WriteUint32(data, &index, CostMatrixVersion)

	numRelays := len(m.RelayIds)

	encoding.WriteUint32(data, &index, uint32(numRelays))

	for _, id := range m.RelayIds {
		encoding.WriteUint64(data, &index, id)
	}

	for _, name := range m.RelayNames {
		encoding.WriteString(data, &index, name, uint32(len(name)))
	}

	numDatacenters := len(m.DatacenterIds)

	if numDatacenters != len(m.DatacenterNames) {
		return nil, fmt.Errorf("Length of Datacenter IDs not equal to length of Datacenter Names: %d != %d", numDatacenters, len(m.DatacenterNames))
	}

	encoding.WriteUint32(data, &index, uint32(numDatacenters))

	for i := 0; i < numDatacenters; i++ {
		encoding.WriteUint64(data, &index, m.DatacenterIds[i])
		encoding.WriteString(data, &index, m.DatacenterNames[i], uint32(len(m.DatacenterNames[i])))
	}

	for i := range m.RelayAddresses {
		tmp := make([]byte, MaxRelayAddressLength)
		copy(tmp, m.RelayAddresses[i])
		encoding.WriteBytes(data, &index, tmp, MaxRelayAddressLength)
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

	return data, nil
}

// Optimize will fill up a *RouteMatrix with the optimized routes based on cost.
func (m *CostMatrix) Optimize(routes *RouteMatrix, thresholdRTT int32) error {
	m.mu.Lock()
	routes.mu.Lock()
	defer func() {
		m.mu.Unlock()
		routes.mu.Unlock()
	}()

	numRelays := len(m.RelayIds)

	entryCount := TriMatrixLength(numRelays)

	routes.RelayIds = m.RelayIds
	routes.RelayNames = m.RelayNames
	routes.RelayAddresses = m.RelayAddresses
	routes.RelayPublicKeys = m.RelayPublicKeys
	routes.DatacenterIds = m.DatacenterIds
	routes.DatacenterNames = m.DatacenterNames
	routes.DatacenterRelays = m.DatacenterRelays
	routes.Entries = make([]RouteMatrixEntry, entryCount)

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
							ikRtt := rtt[ikIndex]
							kjRtt := rtt[kjIndex]
							if ikRtt < 0 || kjRtt < 0 {
								continue
							}
							working[numRoutes].relay = uint64(k)
							working[numRoutes].rtt = ikRtt + kjRtt
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
							ikRtt := rtt[ikIndex]
							kjRtt := rtt[kjIndex]
							if ikRtt < 0 || kjRtt < 0 {
								continue
							}
							indirectRTT := ikRtt + kjRtt
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
	var length uint64
	numRelays := uint64(len(m.RelayIds))
	numDatacenters := uint64(len(m.DatacenterIds))

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

	for _, v := range m.DatacenterRelays {
		// datacenter id + number of relays for that datacenter + allocation for all of those relay ids
		length += uint64(8 + 4 + 8*len(v))
	}

	// length so far + number of rtt entries
	return length + uint64(4*len(m.RTT))
}

// RouteMatrixEntry ...
type RouteMatrixEntry struct {
	DirectRTT      int32
	NumRoutes      int32
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint64
}

// RouteMatrix ...
type RouteMatrix struct {
	mu sync.Mutex

	RelayIds         []uint64
	RelayNames       []string
	RelayAddresses   [][]byte
	RelayPublicKeys  [][]byte
	DatacenterRelays map[uint64][]uint64
	DatacenterIds    []uint64
	DatacenterNames  []string
	Entries          []RouteMatrixEntry
}

func (m *RouteMatrix) Route() (Route, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return Route{
		Type: RouteTypeDirect,
	}, nil
}

// ReadFrom implements the io.ReadFrom interface
func (m *RouteMatrix) ReadFom(r io.Reader) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}

	if err := m.UnmarshalBinary(data); err != nil {
		return 0, err
	}

	return int64(len(data)), nil
}

// WriteTo implements the io.WriteTo interface
func (m *RouteMatrix) WriteTo(w io.Writer) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *RouteMatrix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	m.WriteTo(w)
}

/* Binary data outline for RouteMatrix v2: "->" means seqential elements in memory and not another section, "(...)" mean that section sequentially repeats for however many
 * Version number { uint32 }
 * Number of relays { uint32 }
 * Relay IDs { [NumberOfRelays]uint64 }
 * Relay Names { [NumberOfRelays]string }
 * Number of Datacenters { uint32 }
 * Datacenter ID { [NumberOfDatacenters]uint64 } -> Datacenter Name { [NumberOfDatacenters]string }
 * Relay Addresses { [NumberOfRelays][MaxRelayAddressLength]byte }
 * Relay Public Keys { [NumberOfRelays][crypto.KeySize]byte }
 * Number of Datacenters { uint32 }
 * Datacenter ID { uint64 } -> Number of Relays in Datacenter { uint32 } -> Relay IDs in Datacenter { [NumberOfRelaysInDatacenter]uint64 }
 * RTT Info { []uint32 }
 * Entries { []RouteMatrixEntry } (
 * 	Direct RTT { uint32 }
 *	Number of routes { uint32 }
 *	Route RTT { [8]uint32 }
 *	Number of relays in the route { [8]uint32 }
 *	Relay IDs in each route { [8][5]uint64 }
 * )
 */

// UnmarshalBinary ...
func (m *RouteMatrix) UnmarshalBinary(data []byte) error {
	index := 0

	var version uint32
	if !encoding.ReadUint32(data, &index, &version) {
		return errors.New("[RouteMatrix] invalid read at version number")
	}

	if version > RouteMatrixVersion {
		return fmt.Errorf("unknown route matrix version: %d", version)
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
		return errors.New("[RouteMatrix] invalid read at number of relays")
	}

	m.RelayIds = make([]uint64, numRelays)

	for i := 0; i < int(numRelays); i++ {
		if err := idReadFunc(data, &index, &m.RelayIds[i], "[RouteMatrix] invalid read at relay ids"); err != nil {
			return err
		}
	}

	m.RelayNames = make([]string, numRelays)
	if version >= 1 {
		for i := range m.RelayNames {
			if !encoding.ReadString(data, &index, &m.RelayNames[i], math.MaxInt32) {
				return errors.New("[RouteMatrix] invalid read at relay names")
			}
		}
	}

	if version >= 2 {
		var datacenterCount uint32
		if !encoding.ReadUint32(data, &index, &datacenterCount) {
			return errors.New("[RouteMatrix] invalid read at datacenter count")
		}

		m.DatacenterIds = make([]uint64, datacenterCount)
		m.DatacenterNames = make([]string, datacenterCount)
		for i := 0; i < int(datacenterCount); i++ {
			if err := idReadFunc(data, &index, &m.DatacenterIds[i], "[CostMatrix] invalid read at datacenter ids"); err != nil {
				return err
			}

			if !encoding.ReadString(data, &index, &m.DatacenterNames[i], math.MaxInt32) {
				return errors.New("[RouteMatrix] invalid read at datacenter names")
			}
		}
	}

	m.RelayAddresses = make([][]byte, numRelays)
	for i := range m.RelayAddresses {
		if err := bytesReadFunc(data, &index, &m.RelayAddresses[i], MaxRelayAddressLength, "[RouteMatrix] invalid read at relay addresses"); err != nil {
			return err
		}
	}

	m.RelayPublicKeys = make([][]byte, numRelays)
	for i := range m.RelayPublicKeys {
		if err := bytesReadFunc(data, &index, &m.RelayPublicKeys[i], crypto.KeySize, "[RouteMatrix] invalid read at relay public keys"); err != nil {
			return err
		}
	}

	var numDatacenters uint32
	if !encoding.ReadUint32(data, &index, &numDatacenters) {
		return errors.New("[RouteMatrix] invalid read at number of datacenters (second time)")
	}

	m.DatacenterRelays = make(map[uint64][]uint64)

	for i := 0; i < int(numDatacenters); i++ {
		var datacenterID uint64

		if err := idReadFunc(data, &index, &datacenterID, "[RouteMatrix] invalid read at datacenter id"); err != nil {
			return err
		}

		var numRelaysInDatacenter uint32
		if !encoding.ReadUint32(data, &index, &numRelaysInDatacenter) {
			return errors.New("[RouteMatrix] invalid read at number of relays in datacenter")
		}

		m.DatacenterRelays[datacenterID] = make([]uint64, numRelaysInDatacenter)

		for j := 0; j < int(numRelaysInDatacenter); j++ {
			if err := idReadFunc(data, &index, &m.DatacenterRelays[datacenterID][j], "[RouteMatrix] invalid read at relay ids for datacenter"); err != nil {
				return err
			}
		}
	}

	entryCount := TriMatrixLength(int(numRelays))
	m.Entries = make([]RouteMatrixEntry, entryCount)

	for i := range m.Entries {
		entry := &m.Entries[i]
		var directRtt uint32
		if !encoding.ReadUint32(data, &index, &directRtt) {
			return errors.New("[RouteMatrix] invalid read at direct rtt")
		}
		entry.DirectRTT = int32(directRtt)

		var numRoutes uint32
		if !encoding.ReadUint32(data, &index, &numRoutes) {
			return errors.New("[RouteMatrix] invalid read at number of routes")
		}
		entry.NumRoutes = int32(numRoutes)

		for j := 0; j < int(entry.NumRoutes); j++ {
			var routeRtt uint32
			if !encoding.ReadUint32(data, &index, &routeRtt) {
				return errors.New("[RouteMatrix] invalid read at route rtt")
			}
			entry.RouteRTT[j] = int32(routeRtt)

			var routeNumRelays uint32
			if !encoding.ReadUint32(data, &index, &routeNumRelays) {
				return errors.New("[RouteMatrix] invalid read at number of relays in route")
			}
			entry.RouteNumRelays[j] = int32(routeNumRelays)

			if version >= 3 {
				for k := 0; k < int(entry.RouteNumRelays[j]); k++ {
					if !encoding.ReadUint64(data, &index, &entry.RouteRelays[j][k]) {
						return errors.New("[RouteMatrix] invalid read at relays in route - v3")
					}
				}
			} else {
				for k := 0; k < int(entry.RouteNumRelays[j]); k++ {
					var tmp uint32
					if !encoding.ReadUint32(data, &index, &tmp) {
						return errors.New("[RouteMatrix] invalid read at relays in route - ver < 3")
					}
					entry.RouteRelays[j][k] = uint64(tmp)
				}
			}

		}
	}

	return nil
}

// MarshalBinary ...
func (m RouteMatrix) MarshalBinary() ([]byte, error) {
	data := make([]byte, m.Size())
	index := 0

	encoding.WriteUint32(data, &index, RouteMatrixVersion)

	numRelays := len(m.RelayIds)

	encoding.WriteUint32(data, &index, uint32(numRelays))

	for _, id := range m.RelayIds {
		encoding.WriteUint64(data, &index, id)
	}

	for _, name := range m.RelayNames {
		encoding.WriteString(data, &index, name, uint32(len(name)))
	}

	numDatacenters := len(m.DatacenterIds)

	if numDatacenters != len(m.DatacenterNames) {
		return nil, fmt.Errorf("Length of Datacenter IDs not equal to length of Datacenter Names: %d != %d", numDatacenters, len(m.DatacenterNames))
	}

	encoding.WriteUint32(data, &index, uint32(numDatacenters))

	for i := 0; i < numDatacenters; i++ {
		encoding.WriteUint64(data, &index, m.DatacenterIds[i])
		encoding.WriteString(data, &index, m.DatacenterNames[i], uint32(len(m.DatacenterNames[i])))
	}

	for _, addr := range m.RelayAddresses {
		address := addr

		if len(addr) != MaxRelayAddressLength {
			address = make([]byte, MaxRelayAddressLength)
			copy(address, addr)
		}

		encoding.WriteBytes(data, &index, address, MaxRelayAddressLength)
	}

	for _, pk := range m.RelayPublicKeys {
		encoding.WriteBytes(data, &index, pk, crypto.KeySize)
	}

	numDatacenters = len(m.DatacenterRelays)

	encoding.WriteUint32(data, &index, uint32(numDatacenters))

	for k, v := range m.DatacenterRelays {

		encoding.WriteUint64(data, &index, k)

		numRelaysInDatacenter := len(v)

		encoding.WriteUint32(data, &index, uint32(numRelaysInDatacenter))

		for _, ids := range v {
			encoding.WriteUint64(data, &index, ids)
		}
	}

	for i := 0; i < len(m.Entries); i++ {
		entry := &m.Entries[i]

		encoding.WriteUint32(data, &index, uint32(entry.DirectRTT))

		encoding.WriteUint32(data, &index, uint32(entry.NumRoutes))

		for j := 0; j < int(entry.NumRoutes); j++ {

			encoding.WriteUint32(data, &index, uint32(entry.RouteRTT[j]))

			encoding.WriteUint32(data, &index, uint32(entry.RouteNumRelays[j]))

			for k := 0; k < int(entry.RouteNumRelays[j]); k++ {
				encoding.WriteUint64(data, &index, entry.RouteRelays[j][k])
			}
		}
	}

	return data, nil
}

func (m *RouteMatrix) Size() uint64 {
	var length uint64
	numRelays := uint64(len(m.RelayIds))
	numDatacenters := uint64(len(m.DatacenterIds))
	// same as CostMatrix's
	length = 4 + 4 + 8*numRelays

	for _, name := range m.RelayNames {
		// same as CostMatrix's
		length += uint64(4 + len(name))
	}

	// same as CostMatrix's
	length += 8 + 8*numDatacenters

	for _, name := range m.DatacenterNames {
		// same as CostMatrix's
		length += uint64(4 + len(name))
	}

	// same as CostMatrix's
	length += numRelays*uint64(MaxRelayAddressLength+crypto.KeySize) + 4

	// same as CostMatrix's
	for _, v := range m.DatacenterRelays {
		length += uint64(8 + 4 + 8*len(v))
	}

	for _, entry := range m.Entries {
		// DirectRTT + NumRoutes + allocation for RouteRTTs + allocation for RouteNumRelays
		length += uint64(4 + 4 + 4*len(entry.RouteRTT) + 4*len(entry.RouteNumRelays))

		for _, relays := range entry.RouteRelays {
			// allocation for relay ids
			length += uint64(8 * len(relays))
		}
	}

	return length
}

// RouteManager ...
type RouteManager struct {
	NumRoutes      int
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteHash      [MaxRoutesPerRelayPair]uint64
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint64
}

// fnv64
func RouteHash(relays ...uint64) uint64 {
	// http://www.isthe.com/chongo/tech/comp/fnv/
	const fnv64OffsetBasis = uint64(0xCBF29CE484222325)

	hash := uint64(0)
	for i := range relays {
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 56) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 48) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 40) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 32) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 24) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 16) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= (relays[i] >> 8) & 0xFF
		hash *= fnv64OffsetBasis
		hash ^= relays[i] & 0xFF
	}

	return hash
}

// AddRoute ...
func (manager *RouteManager) AddRoute(rtt int32, relays ...uint64) {
	if rtt < 0 {
		return
	}

	if manager.NumRoutes == 0 {

		// no routes yet. add the route

		manager.NumRoutes = 1
		manager.RouteRTT[0] = rtt
		manager.RouteHash[0] = RouteHash(relays...)
		manager.RouteNumRelays[0] = int32(len(relays))
		for i := range relays {
			manager.RouteRelays[0][i] = relays[i]
		}

	} else if manager.NumRoutes < MaxRoutesPerRelayPair {

		// not at max routes yet. insert according RTT sort order

		hash := RouteHash(relays...)
		for i := 0; i < manager.NumRoutes; i++ {
			if hash == manager.RouteHash[i] {
				return
			}
		}

		if rtt >= manager.RouteRTT[manager.NumRoutes-1] {

			// RTT is greater than existing entries. append.

			manager.RouteRTT[manager.NumRoutes] = rtt
			manager.RouteHash[manager.NumRoutes] = hash
			manager.RouteNumRelays[manager.NumRoutes] = int32(len(relays))
			for i := range relays {
				manager.RouteRelays[manager.NumRoutes][i] = relays[i]
			}
			manager.NumRoutes++

		} else {

			// RTT is lower than at least one entry. insert.

			insertIndex := manager.NumRoutes - 1
			for {
				if insertIndex == 0 || rtt > manager.RouteRTT[insertIndex-1] {
					break
				}
				insertIndex--
			}
			manager.NumRoutes++
			for i := manager.NumRoutes - 1; i > insertIndex; i-- {
				manager.RouteRTT[i] = manager.RouteRTT[i-1]
				manager.RouteHash[i] = manager.RouteHash[i-1]
				manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
				for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
					manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
				}
			}
			manager.RouteRTT[insertIndex] = rtt
			manager.RouteHash[insertIndex] = hash
			manager.RouteNumRelays[insertIndex] = int32(len(relays))
			for i := range relays {
				manager.RouteRelays[insertIndex][i] = relays[i]
			}

		}

	} else {

		// route set is full. only insert if lower RTT than at least one current route.

		if rtt >= manager.RouteRTT[manager.NumRoutes-1] {
			return
		}

		hash := RouteHash(relays...)
		for i := 0; i < manager.NumRoutes; i++ {
			if hash == manager.RouteHash[i] {
				return
			}
		}

		insertIndex := manager.NumRoutes - 1
		for {
			if insertIndex == 0 || rtt > manager.RouteRTT[insertIndex-1] {
				break
			}
			insertIndex--
		}

		for i := manager.NumRoutes - 1; i > insertIndex; i-- {
			manager.RouteRTT[i] = manager.RouteRTT[i-1]
			manager.RouteHash[i] = manager.RouteHash[i-1]
			manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
			for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
				manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
			}
		}

		manager.RouteRTT[insertIndex] = rtt
		manager.RouteHash[insertIndex] = hash
		manager.RouteNumRelays[insertIndex] = int32(len(relays))

		for i := range relays {
			manager.RouteRelays[insertIndex][i] = relays[i]
		}

	}
}
