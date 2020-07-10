package routing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"sort"
	"strconv"
	"sync"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	// IMPORTANT: Increment this when you change the binary format
	RouteMatrixVersion = 6
)

// todo: ryan, there's absolutely no reason to keep compatibility past the last route matrix version in production
// currently that is version 5. this means we can drop anything before route matrix version 5.
// the versioning code and tests are super complex :(

type RouteMatrixEntry struct {
	DirectRTT      int32
	NumRoutes      int32
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint64
}

type RouteMatrix struct {
	RelayIndices map[uint64]int

	RelayIDs              []uint64
	RelayNames            []string
	RelayAddresses        [][]byte
	RelayLatitude         []float64
	RelayLongitude        []float64
	RelayPublicKeys       [][]byte
	DatacenterRelays      map[uint64][]uint64
	DatacenterIDs         []uint64
	DatacenterNames       []string
	Entries               []RouteMatrixEntry
	RelaySellers          []Seller
	RelaySessionCounts    []uint32
	RelayMaxSessionCounts []uint32

	responseBuffer     []byte
	reponseBufferMutex sync.RWMutex

	analysisBuffer      []byte
	analysisBufferMutex sync.RWMutex

	relayAddressCache []*net.UDPAddr
}

func Truncate(value float64) float64 {
	return float64(int64(value))
}

func HaversineDistance(lat1 float64, long1 float64, lat2 float64, long2 float64) float64 {
	lat1 *= math.Pi / 180
	lat2 *= math.Pi / 180
	long1 *= math.Pi / 180
	long2 *= math.Pi / 180
	delta_lat := lat2 - lat1
	delta_long := long2 - long1
	lat_sine := math.Sin(delta_lat / 2)
	long_sine := math.Sin(delta_long / 2)
	a := lat_sine*lat_sine + math.Cos(lat1)*math.Cos(lat2)*long_sine*long_sine
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	r := 6371.0
	d := r * c
	return d // kilometers
}

func (m *RouteMatrix) GetNearRelays(latitude float64, longitude float64, maxNearRelays int) ([]Relay, error) {
	type NearRelayData struct {
		id       uint64
		distance int
	}

	nearRelayData := make([]NearRelayData, len(m.RelayIDs))

	// IMPORTANT: Truncate the lat/long values to nearest integer.
	// This fixes numerical instabilities that can happen in the haversine function
	// when two relays are really close together, they can get sorted differently in
	// subsequent passes otherwise.

	lat1 := Truncate(latitude)
	long1 := Truncate(longitude)

	for i, relayID := range m.RelayIDs {
		nearRelayData[i].id = relayID
		lat2 := m.RelayLatitude[i]
		long2 := m.RelayLongitude[i]
		nearRelayData[i].distance = int(HaversineDistance(lat1, long1, lat2, long2))
	}

	// IMPORTANT: Sort near relays by distance using a *stable sort*
	// This is necessary to ensure that relays are always sorted in the same order,
	// even when some relays have the same integer distance from the client. Without this
	// the set of near relays passed down to the SDK can be different from one slice to the next!

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].distance < nearRelayData[j].distance })

	if len(nearRelayData) > maxNearRelays {
		nearRelayData = nearRelayData[:maxNearRelays]
	}

	// Now that the relays are sorted by distance, construct the final near relays slice and return it
	nearRelays := make([]Relay, len(nearRelayData))
	var err error
	for i, nearRelayData := range nearRelayData {
		nearRelays[i], err = m.ResolveRelay(nearRelayData.id)
		if err != nil {
			return nil, fmt.Errorf("could not resolve relay ID %d: %v", nearRelayData.id, err)
		}
	}

	return nearRelays, nil
}

func (m *RouteMatrix) ResolveRelay(id uint64) (Relay, error) {
	relayIndex, ok := m.RelayIndices[id]
	if !ok {
		return Relay{}, fmt.Errorf("relay %d not in matrix", id)
	}

	if relayIndex >= len(m.RelayIDs) ||
		relayIndex >= len(m.RelayNames) ||
		relayIndex >= len(m.relayAddressCache) ||
		relayIndex >= len(m.RelayPublicKeys) ||
		relayIndex >= len(m.RelaySellers) ||
		relayIndex >= len(m.RelaySessionCounts) ||
		relayIndex >= len(m.RelayMaxSessionCounts) {
		return Relay{}, fmt.Errorf("relay %d has an invalid index %d", id, relayIndex)
	}

	return Relay{
		ID:        m.RelayIDs[relayIndex],
		Name:      m.RelayNames[relayIndex],
		Addr:      *m.relayAddressCache[relayIndex],
		PublicKey: m.RelayPublicKeys[relayIndex],
		Seller:    m.RelaySellers[relayIndex],
		TrafficStats: RelayTrafficStats{
			SessionCount: uint64(m.RelaySessionCounts[relayIndex]),
		},
		MaxSessions: m.RelayMaxSessionCounts[relayIndex],
	}, nil
}

// GetDatacenterRelays will return the set of Relays in the provided Datacenter
func (m *RouteMatrix) GetDatacenterRelays(d Datacenter) []Relay {
	relayIDs, ok := m.DatacenterRelays[d.ID]
	if !ok {
		return nil
	}

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

// GetRoutes returns acceptable routes between the set of near relays and destination relays.
// maxAcceptableRoutes is the maximum number of acceptable routes that can be returned.
// Returns the acceptable routes, the number of routes in the slice, and an error, if one exists.
func (m *RouteMatrix) GetRoutes(near []Relay, dest []Relay, maxAcceptableRoutes uint64) ([]Route, int, error) {
	acceptableRoutes := make([]Route, maxAcceptableRoutes)
	var acceptableRoutesLength uint64

	// Keep a parallel slice of route hashes for quick route comparisons
	acceptableRoutesHashes := make([]uint64, maxAcceptableRoutes)

	for i := range acceptableRoutes {
		acceptableRoutes[i] = Route{
			Stats: Stats{
				RTT:        InvalidRouteValue,
				Jitter:     InvalidRouteValue,
				PacketLoss: InvalidRouteValue,
			},
		}

		acceptableRoutesHashes[i] = acceptableRoutes[i].Hash64()
	}

	// For all (near, dest) relay pairs, check each route to see if it is acceptable
	for _, nearRelay := range near {
		for _, destRelay := range dest {
			entryIndex, reverse := m.GetEntryIndex(&nearRelay, &destRelay)

			entry := &m.Entries[entryIndex]

			for i := 0; i < int(entry.NumRoutes); i++ {
				routeRTT := entry.RouteRTT[i]

				if acceptableRoutesLength == 0 {
					// no routes added yet, add the route

					routeRelays := make([]Relay, entry.RouteNumRelays[i])
					var err error

					numRelays := int(entry.RouteNumRelays[i])
					for j := 0; j < numRelays; j++ {
						relayIndex := entry.RouteRelays[i][j]
						relayID := m.RelayIDs[relayIndex]

						if !reverse {
							routeRelays[j], err = m.ResolveRelay(relayID)
						} else {
							routeRelays[numRelays-1-j], err = m.ResolveRelay(relayID)
						}

						if err != nil {
							return nil, 0, err
						}
					}

					acceptableRoutes[acceptableRoutesLength] = Route{
						Relays: routeRelays,
						Stats: Stats{
							RTT: float64(int32(math.Ceil(near[i].ClientStats.RTT)) + routeRTT),
						},
					}

					acceptableRoutesLength++

				} else if acceptableRoutesLength < maxAcceptableRoutes {
					// not at max routes yet, insert according RTT sort order

				} else {
					// route set is full, only insert if lower RTT than at least one current route.
				}
			}

		}
	}
}

// Returns the index in the route matrix representing the route between the near Relay and dest Relay and whether or not to reverse them
func (m *RouteMatrix) GetEntryIndex(near *Relay, dest *Relay) (int, bool) {
	destidx, ok := m.RelayIndices[dest.ID]
	if !ok {
		return -1, false
	}

	nearidx, ok := m.RelayIndices[near.ID]
	if !ok {
		return -1, false
	}

	return TriMatrixIndex(nearidx, destidx), destidx > nearidx
}

// FillRoutes populates the given route buffer.
// It takes the entryIndex and reverse data and fills the given route buffer, incrementing the routeIndex after
// each route it adds.
func (m *RouteMatrix) FillRoutes(routes []Route, routeIndex *int, nearCost int, entryIndex int, reverse bool) error {
	var err error

	entry := m.Entries[entryIndex]

	for i := 0; i < int(entry.NumRoutes); i++ {
		numRelays := int(entry.RouteNumRelays[i])

		routeRelays := make([]Relay, numRelays)

		for j := 0; j < numRelays; j++ {
			relayIndex := entry.RouteRelays[i][j]

			id := m.RelayIDs[relayIndex]

			if !reverse {
				routeRelays[j], err = m.ResolveRelay(id)
			} else {
				routeRelays[numRelays-1-j], err = m.ResolveRelay(id)
			}

			if err != nil {
				return err
			}
		}

		route := Route{
			Relays: routeRelays,
			Stats: Stats{
				RTT: float64(nearCost + int(m.Entries[entryIndex].RouteRTT[i])),
			},
		}

		if *routeIndex >= len(routes) {
			continue
		}

		routes[*routeIndex] = route
		*routeIndex++
	}

	return nil
}

// implements the io.ReadFrom interface
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

// implements the io.WriteTo interface
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

	m.RelayIndices = make(map[uint64]int)
	m.RelayIDs = make([]uint64, numRelays)

	for i := 0; i < int(numRelays); i++ {
		var tmp uint64
		if err := idReadFunc(data, &index, &tmp, "[RouteMatrix] invalid read at relay ids"); err != nil {
			return err
		}
		m.RelayIndices[tmp] = i
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

	m.RelayLatitude = make([]float64, numRelays)
	m.RelayLongitude = make([]float64, numRelays)

	if version >= 6 {

		for i := range m.RelayLatitude {
			if !encoding.ReadFloat64(data, &index, &m.RelayLatitude[i]) {
				return errors.New("[RouteMatrix] invalid read at relay latitude")
			}
		}

		for i := range m.RelayLongitude {
			if !encoding.ReadFloat64(data, &index, &m.RelayLongitude[i]) {
				return errors.New("[RouteMatrix] invalid read at relay longitude")
			}
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

	m.UpdateRelayAddressCache()

	return nil
}

func (m *RouteMatrix) UpdateRelayAddressCache() error {
	if len(m.relayAddressCache) == 0 && len(m.RelayIDs) > 0 {
		m.relayAddressCache = make([]*net.UDPAddr, len(m.RelayIDs))
		for i := range m.RelayIDs {
			// This trim is necessary because RelayAddresses has a fixed size of MaxRelayAddressLength which causes extra 0 bytes to be parsed if we don't trim
			host, port, err := net.SplitHostPort(string(bytes.Trim(m.RelayAddresses[i], string([]byte{0x00}))))
			if err != nil {
				return err
			}

			iport, err := strconv.Atoi(port)
			if err != nil {
				return err
			}

			m.relayAddressCache[i] = &net.UDPAddr{
				IP:   net.ParseIP(host),
				Port: int(iport),
			}
		}
	}
	return nil
}

func (m *RouteMatrix) MarshalBinary() ([]byte, error) {
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
	var length uint64
	numRelays := uint64(len(m.RelayIDs))
	numDatacenters := uint64(len(m.DatacenterIDs))
	length = 4 + 4 + 8*numRelays

	for _, name := range m.RelayNames {
		length += uint64(4 + len(name))
	}

	length += 8 + 8*numDatacenters

	for _, name := range m.DatacenterNames {
		length += uint64(4 + len(name))
	}

	length += numRelays*uint64(MaxRelayAddressLength+crypto.KeySize) + 4

	length += numRelays * 8 * 2

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
	var b bytes.Buffer
	for _, routeEntry := range m.Entries {
		for routeidx := int32(0); routeidx < routeEntry.NumRoutes; routeidx++ {
			b.WriteString(fmt.Sprintf("RTT(%d) ", routeEntry.RouteRTT[routeidx]))

			for relayidx := int32(0); relayidx < routeEntry.RouteNumRelays[routeidx]; relayidx++ {
				relay, err := m.ResolveRelay(m.RelayIDs[routeEntry.RouteRelays[routeidx][relayidx]])
				if err != nil {
					fmt.Println(err)
				}

				// display ip addr locally
				// display actual name in dev/prod
				name := relay.Addr.String()
				if len(relay.Name) != 0 {
					name = relay.Name
				}
				b.WriteString(name)
				b.WriteString(" ")
			}
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	writer.Write(b.Bytes())
}

func (m *RouteMatrix) WriteAnalysisTo(writer io.Writer) {
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

	fmt.Fprintf(writer, "%s Improvement:\n\n", "RTT")
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
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have only one route\n", float64(relayPairsWithOneRoute)/float64(numRelayPairs)*100)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have no route\n", float64(relayPairsWithNoRoutes)/float64(numRelayPairs)*100)
}

func (m *RouteMatrix) GetResponseData() []byte {
	m.reponseBufferMutex.RLock()
	data := m.responseBuffer
	m.reponseBufferMutex.RUnlock()
	return data
}

func (m *RouteMatrix) WriteResponseData() error {
	var buffer bytes.Buffer
	if _, err := m.WriteTo(&buffer); err != nil {
		return err
	}
	m.reponseBufferMutex.Lock()
	m.responseBuffer = buffer.Bytes()
	m.reponseBufferMutex.Unlock()
	return nil
}

func (m *RouteMatrix) GetAnalysisData() []byte {
	m.analysisBufferMutex.RLock()
	data := m.analysisBuffer
	m.analysisBufferMutex.RUnlock()
	return data
}

func (m *RouteMatrix) WriteAnalysisData() {
	var buffer bytes.Buffer
	m.WriteAnalysisTo(&buffer)

	m.analysisBufferMutex.Lock()
	m.analysisBuffer = buffer.Bytes()
	m.analysisBufferMutex.Unlock()
}
