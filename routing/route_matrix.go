package routing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	// RouteMatrixVersion ...
	// IMPORTANT: Increment this when you change the binary format
	RouteMatrixVersion = 5
)

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
	mu sync.RWMutex

	RelayIndicies map[uint64]int

	RelayIDs              []uint64
	RelayNames            []string
	RelayAddresses        [][]byte
	RelayPublicKeys       [][]byte
	DatacenterRelays      map[uint64][]uint64
	DatacenterIDs         []uint64
	DatacenterNames       []string
	Entries               []RouteMatrixEntry
	RelaySellers          []Seller
	RelaySessionCounts    []uint32
	RelayMaxSessionCounts []uint32
}

func (m *RouteMatrix) ResolveRelay(id uint64) (Relay, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	relayIndex, ok := m.RelayIndicies[id]
	if !ok {
		return Relay{}, fmt.Errorf("relay %d not in matrix", id)
	}

	if relayIndex >= len(m.RelayAddresses) ||
		relayIndex >= len(m.RelayPublicKeys) ||
		relayIndex >= len(m.RelaySellers) ||
		relayIndex >= len(m.RelaySessionCounts) ||
		relayIndex >= len(m.RelayMaxSessionCounts) {
		return Relay{}, fmt.Errorf("relay %d has an invalid index %d", id, relayIndex)
	}

	host, port, err := net.SplitHostPort(string(bytes.Trim(m.RelayAddresses[relayIndex], string([]byte{0x00}))))
	if err != nil {
		return Relay{}, err
	}

	iport, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		return Relay{}, err
	}

	return Relay{
		ID: m.RelayIDs[relayIndex],
		Addr: net.UDPAddr{
			IP:   net.ParseIP(host),
			Port: int(iport),
		},
		PublicKey: m.RelayPublicKeys[relayIndex],
		Seller:    m.RelaySellers[relayIndex],
		TrafficStats: RelayTrafficStats{
			SessionCount: uint64(m.RelaySessionCounts[relayIndex]),
		},
		MaxSessions: m.RelayMaxSessionCounts[relayIndex],
	}, nil
}

// RelaysIn will return the set of Relays in the provided Datacenter
func (m *RouteMatrix) RelaysIn(d Datacenter) []Relay {
	m.mu.RLock()

	relayIDs, ok := m.DatacenterRelays[d.ID]
	if !ok {
		m.mu.RUnlock()
		return nil
	}

	m.mu.RUnlock()

	var err error
	relayLength := len(relayIDs)

	if relayLength <= 0 {
		return nil
	}

	relays := make([]Relay, relayLength)
	for i := 0; i < relayLength; i++ {
		relays[i], err = m.ResolveRelay(relayIDs[i])
		if err != nil {
			continue
		}
	}

	return relays
}

// Routes will return a set of routes for each from and to Relay based on the given selectors.
// The selectors are chained together in order, so the selected routes from the first selector will be passed
// as the argument to the second selector. If at any point a selector fails to select a new slice of routes,
// the chain breaks.
func (m *RouteMatrix) Routes(from []Relay, to []Relay, routeSelectors ...SelectorFunc) ([]Route, error) {
	m.mu.RLock()

	type RelayPairResult struct {
		fromtoidx int  // The index in the route matrix entry
		reverse   bool // Whether or not to reverse the relays to stay on the same side of the diagnol in the triangular matrix
	}

	relayPairLength := len(from) * len(to)
	relayPairResults := make([]RelayPairResult, relayPairLength)

	// Do a "first pass" to determine the size of the Route buffer
	var routeTotal int
	for i, fromrelay := range from {
		for j, torelay := range to {
			fromtoidx, reverse := m.getFromToRelayIndex(fromrelay, torelay)

			// Add a bad pair result so that the second pass will skip over it.
			// This way we don't have to append only good results to a new list, which is more expensive.
			if fromtoidx < 0 || fromtoidx >= len(m.Entries) {
				relayPairResults[i+j*len(from)] = RelayPairResult{-1, false}
				continue
			}

			relayPairResults[i+j*len(from)] = RelayPairResult{fromtoidx, reverse}
			routeTotal += int(m.Entries[fromtoidx].NumRoutes)
		}
	}

	m.mu.RUnlock()

	// Now that we have the route total, make the Route buffer and fill it
	var routeIndex int
	routes := make([]Route, routeTotal)
	for i := 0; i < relayPairLength; i++ {
		if relayPairResults[i].fromtoidx >= 0 {
			m.fillRoutes(routes, &routeIndex, relayPairResults[i].fromtoidx, relayPairResults[i].reverse)
		}
	}

	// No routes found
	if len(routes) == 0 {
		return nil, errors.New("no routes in route matrix")
	}

	// Apply the selectors in order
	for _, selector := range routeSelectors {
		routes = selector(routes)

		if len(routes) == 0 {
			break
		}
	}

	return routes, nil
}

// Returns the index in the route matrix representing the route between the from Relay and to Relay and whether or not to reverse them
func (m *RouteMatrix) getFromToRelayIndex(from Relay, to Relay) (int, bool) {
	toidx, ok := m.RelayIndicies[to.ID]
	if !ok {
		return -1, false
	}

	fromidx, ok := m.RelayIndicies[from.ID]
	if !ok {
		return -1, false
	}

	return TriMatrixIndex(fromidx, toidx), toidx > fromidx
}

// fillRoutes is just the internal function to populate the given route buffer.
// It takes the fromtoidx and reverse data and fills the given route buffer, incrementing the routeIndex after
// each route it adds.
func (m *RouteMatrix) fillRoutes(routes []Route, routeIndex *int, fromtoidx int, reverse bool) error {
	var err error

	m.mu.RLock()
	entry := m.Entries[fromtoidx]
	m.mu.RUnlock()

	for i := 0; i < int(entry.NumRoutes); i++ {
		numRelays := int(entry.RouteNumRelays[i])

		routeRelays := make([]Relay, numRelays)

		for j := 0; j < numRelays; j++ {
			relayIndex := entry.RouteRelays[i][j]

			m.mu.RLock()
			id := m.RelayIDs[relayIndex]
			m.mu.RUnlock()

			if !reverse {
				routeRelays[j], err = m.ResolveRelay(id)
			} else {
				routeRelays[numRelays-1-j], err = m.ResolveRelay(id)
			}

			if err != nil {
				return err
			}
		}

		m.mu.RLock()
		route := Route{
			Relays: routeRelays,
			Stats: Stats{
				RTT: float64(m.Entries[fromtoidx].RouteRTT[i]),
			},
		}
		m.mu.RUnlock()

		if *routeIndex >= len(routes) {
			continue
		}

		routes[*routeIndex] = route
		*routeIndex++
	}

	return nil
}

// ReadFrom implements the io.ReadFrom interface
func (m *RouteMatrix) ReadFrom(r io.Reader) (int64, error) {
	if r == nil {
		return 0, errors.New("reader is nil")
	}

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
	_, err := m.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

/* Binary data outline for RouteMatrix v5: "->" means seqential elements in memory and not another section, "(...)" mean that section sequentially repeats for however many
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
 * Entries { []RouteMatrixEntry } (
 * 	Direct RTT { uint32 }
 *	Number of routes { uint32 }
 *	Route RTT { [8]uint32 }
 *	Number of relays in the route { [8]uint32 }
 *	Relay IDs in each route { [8][5]uint64 }
 * )
 * Sellers { [NumberOfRelays]Seller } (
 *	ID { string }
 *	Name { string }
 * 	IngressPriceCents { uint64 }
 *	EgressPriceCents { uint64 }
 * )
 * Relay Session Counts { [NumberOfRelays]uint32 }
 * Relay Max Session Counts { [NumberOfRelays]uint32 }
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

	m.mu.Lock()
	defer m.mu.Unlock()

	m.RelayIndicies = make(map[uint64]int)
	m.RelayIDs = make([]uint64, numRelays)

	for i := 0; i < int(numRelays); i++ {
		var tmp uint64
		if err := idReadFunc(data, &index, &tmp, "[RouteMatrix] invalid read at relay ids"); err != nil {
			return err
		}
		m.RelayIndicies[tmp] = i
		m.RelayIDs[i] = tmp
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

		m.DatacenterIDs = make([]uint64, datacenterCount)
		m.DatacenterNames = make([]string, datacenterCount)
		for i := 0; i < int(datacenterCount); i++ {
			if err := idReadFunc(data, &index, &m.DatacenterIDs[i], "[RouteMatrix] invalid read at datacenter ids"); err != nil {
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
		var directRTT uint32
		if !encoding.ReadUint32(data, &index, &directRTT) {
			return errors.New("[RouteMatrix] invalid read at direct rtt")
		}
		entry.DirectRTT = int32(directRTT)

		var numRoutes uint32
		if !encoding.ReadUint32(data, &index, &numRoutes) {
			return errors.New("[RouteMatrix] invalid read at number of routes")
		}
		entry.NumRoutes = int32(numRoutes)

		for j := 0; j < int(entry.NumRoutes); j++ {
			var routeRTT uint32
			if !encoding.ReadUint32(data, &index, &routeRTT) {
				return errors.New("[RouteMatrix] invalid read at route rtt")
			}
			entry.RouteRTT[j] = int32(routeRTT)

			var routeNumRelays uint32
			if !encoding.ReadUint32(data, &index, &routeNumRelays) {
				return errors.New("[RouteMatrix] invalid read at number of relays in route")
			}
			entry.RouteNumRelays[j] = int32(routeNumRelays)

			for k := 0; k < int(entry.RouteNumRelays[j]); k++ {
				if err := idReadFunc(data, &index, &entry.RouteRelays[j][k], "[RouteMatrix] invalid read at relays in route"); err != nil {
					return err
				}
			}
		}
	}

	m.RelaySellers = make([]Seller, numRelays)
	if version >= 4 {
		for i := range m.RelaySellers {
			if !encoding.ReadString(data, &index, &m.RelaySellers[i].ID, math.MaxInt32) {
				return errors.New("[RouteMatrix] invalid read on relay seller ID")
			}
			if !encoding.ReadString(data, &index, &m.RelaySellers[i].Name, math.MaxInt32) {
				return errors.New("[RouteMatrix] invalid read on relay seller name")
			}
			if !encoding.ReadUint64(data, &index, &m.RelaySellers[i].IngressPriceCents) {
				return errors.New("[RouteMatrix] invalid read on relay seller ingress price")
			}
			if !encoding.ReadUint64(data, &index, &m.RelaySellers[i].EgressPriceCents) {
				return errors.New("[RouteMatrix] invalid read on relay seller egress price")
			}
		}
	}

	m.RelaySessionCounts = make([]uint32, numRelays)
	if version >= 5 {
		for i := range m.RelaySessionCounts {
			if !encoding.ReadUint32(data, &index, &m.RelaySessionCounts[i]) {
				return errors.New("[RouteMatrix] invalid read on relay session count")
			}
		}
	}

	m.RelayMaxSessionCounts = make([]uint32, numRelays)
	if version >= 5 {
		for i := range m.RelayMaxSessionCounts {
			if !encoding.ReadUint32(data, &index, &m.RelayMaxSessionCounts[i]) {
				return errors.New("[RouteMatrix] invalid read on relay max session count")
			}
		}
	}

	return nil
}

// MarshalBinary ...
func (m *RouteMatrix) MarshalBinary() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := make([]byte, m.Size())
	index := 0

	encoding.WriteUint32(data, &index, RouteMatrixVersion)

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

	for _, seller := range m.RelaySellers {
		encoding.WriteString(data, &index, seller.ID, uint32(len(seller.ID)))
		encoding.WriteString(data, &index, seller.Name, uint32(len(seller.Name)))
		encoding.WriteUint64(data, &index, seller.IngressPriceCents)
		encoding.WriteUint64(data, &index, seller.EgressPriceCents)
	}

	for i := range m.RelaySessionCounts {
		encoding.WriteUint32(data, &index, m.RelaySessionCounts[i])
	}

	for i := range m.RelayMaxSessionCounts {
		encoding.WriteUint32(data, &index, m.RelayMaxSessionCounts[i])
	}

	return data, nil
}

func (m *RouteMatrix) Size() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var length uint64
	numRelays := uint64(len(m.RelayIDs))
	numDatacenters := uint64(len(m.DatacenterIDs))
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
		length += uint64(4 + 4 + 4 + 4 + 4*len(entry.RouteRTT) + 4*len(entry.RouteNumRelays))

		for _, relays := range entry.RouteRelays {
			// allocation for relay ids
			length += uint64(8 * len(relays))
		}
	}

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

func (m *RouteMatrix) WriteRoutesTo(writer io.Writer) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var b bytes.Buffer
	for _, routeEntry := range m.Entries {
		for routeidx := int32(0); routeidx < routeEntry.NumRoutes; routeidx++ {
			b.WriteString(fmt.Sprintf("RTT(%d) ", routeEntry.RouteRTT[routeidx]))

			for relayidx := int32(0); relayidx < routeEntry.RouteNumRelays[routeidx]; relayidx++ {
				relay, err := m.ResolveRelay(m.RelayIDs[routeEntry.RouteRelays[routeidx][relayidx]])
				if err != nil {
					fmt.Println(err)
				}
				b.WriteString(relay.Addr.String())
				b.WriteString(" ")
			}
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	writer.Write(b.Bytes())
}

func (m *RouteMatrix) WriteAnalysisTo(writer io.Writer) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	src := m.RelayIDs
	dest := m.RelayIDs

	numRelayPairs := 0.0
	numValidRelayPairs := 0.0

	numValidRelayPairsWithoutImprovement := 0.0

	buckets := make([]int, 11)

	for i := range src {
		for j := range dest {
			if j < i {
				numRelayPairs++
				abFlatIndex := TriMatrixIndex(i, j)
				if len(m.Entries[abFlatIndex].RouteRTT) > 0 {
					numValidRelayPairs++
					improvement := m.Entries[abFlatIndex].DirectRTT - m.Entries[abFlatIndex].RouteRTT[0]
					if improvement > 0.0 {
						if improvement <= 5 {
							buckets[0]++
						} else if improvement <= 10 {
							buckets[1]++
						} else if improvement <= 15 {
							buckets[2]++
						} else if improvement <= 20 {
							buckets[3]++
						} else if improvement <= 25 {
							buckets[4]++
						} else if improvement <= 30 {
							buckets[5]++
						} else if improvement <= 35 {
							buckets[6]++
						} else if improvement <= 40 {
							buckets[7]++
						} else if improvement <= 45 {
							buckets[8]++
						} else if improvement <= 50 {
							buckets[9]++
						} else {
							buckets[10]++
						}
					} else {
						numValidRelayPairsWithoutImprovement++
					}
				}
			}
		}
	}

	fmt.Fprintf(writer, "\n%s Improvement:\n\n", "RTT")
	fmt.Fprintf(writer, "    None: %d (%.2f%%)\n", int(numValidRelayPairsWithoutImprovement), numValidRelayPairsWithoutImprovement/numValidRelayPairs*100.0)

	for i := range buckets {
		if i != len(buckets)-1 {
			fmt.Fprintf(writer, "    %d-%d%s: %d (%.2f%%)\n", i*5, (i+1)*5, "ms", buckets[i], float64(buckets[i])/numValidRelayPairs*100.0)
		} else {
			fmt.Fprintf(writer, "    %d%s+: %d (%.2f%%)\n", i*5, "ms", buckets[i], float64(buckets[i])/numValidRelayPairs*100.0)
		}
	}

	totalRoutes := uint64(0)
	maxRouteLength := int32(0)
	maxRoutesPerRelayPair := int32(0)
	relayPairsWithNoRoutes := 0
	relayPairsWithOneRoute := 0
	averageRouteLength := 0.0

	for i := range src {
		for j := range dest {
			if j < i {
				ijFlatIndex := TriMatrixIndex(i, j)
				n := m.Entries[ijFlatIndex].NumRoutes
				if n > maxRoutesPerRelayPair {
					maxRoutesPerRelayPair = n
				}
				totalRoutes += uint64(n)
				if n == 0 {
					relayPairsWithNoRoutes++
				}
				if n == 1 {
					relayPairsWithOneRoute++
				}
				for k := 0; k < int(m.Entries[ijFlatIndex].NumRoutes); k++ {
					numRelays := m.Entries[ijFlatIndex].RouteNumRelays[k]
					averageRouteLength += float64(numRelays)
					if numRelays > maxRouteLength {
						maxRouteLength = numRelays
					}
				}
			}
		}
	}

	averageNumRoutes := float64(totalRoutes) / float64(numRelayPairs)
	averageRouteLength = averageRouteLength / float64(totalRoutes)

	fmt.Fprintf(writer, "\n%s Summary:\n\n", "Route")
	fmt.Fprintf(writer, "    %.1f routes per relay pair on average (%d max)\n", averageNumRoutes, maxRoutesPerRelayPair)
	fmt.Fprintf(writer, "    %.1f relays per route on average (%d max)\n", averageRouteLength, maxRouteLength)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have no route\n", float64(relayPairsWithNoRoutes)/float64(numRelayPairs)*100)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have only one route\n", float64(relayPairsWithOneRoute)/float64(numRelayPairs)*100)
	fmt.Fprintf(writer, "\n")
}
