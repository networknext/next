package core

import (
	"fmt"
	"encoding/binary"
)

type RouteMatrix struct {
	RelayIds         []RelayId
	RelayNames       []string
	RelayAddresses   [][]byte
	RelayPublicKeys  [][]byte
	DatacenterRelays map[DatacenterId][]RelayId
	DatacenterIds    []DatacenterId
	DatacenterNames  []string
	Entries          []RouteMatrixEntry
}

type RouteMatrixEntry struct {
	DirectRTT      int32
	NumRoutes      int32
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint32
}

// IMPORTANT: Increment this when you change the binary format
const RouteMatrixVersion = 2

func WriteRouteMatrix(buffer []byte, routeMatrix *RouteMatrix) []byte {

	var index int

	// todo: update to new way to read/write binary as per backend.go

	binary.LittleEndian.PutUint32(buffer[index:], RouteMatrixVersion)
	index += 4

	numRelays := len(routeMatrix.RelayIds)
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelays))
	index += 4

	for i := range routeMatrix.RelayIds {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.RelayIds[i]))
		index += 4
	}

	for i := range routeMatrix.RelayNames {
		index += WriteString(buffer[index:], routeMatrix.RelayNames[i])
	}

	if len(routeMatrix.DatacenterIds) != len(routeMatrix.DatacenterNames) {
		panic("datacenter ids length does not match datacenter names length")
	}

	binary.LittleEndian.PutUint32(buffer[index:], uint32(len(routeMatrix.DatacenterIds)))
	index += 4

	for i := 0; i < len(routeMatrix.DatacenterIds); i++ {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.DatacenterIds[i]))
		index += 4
		index += WriteString(buffer[index:], routeMatrix.DatacenterNames[i])
	}

	for i := range routeMatrix.RelayAddresses {
		index += WriteBytes(buffer[index:], routeMatrix.RelayAddresses[i])
	}

	for i := range routeMatrix.RelayPublicKeys {
		index += WriteBytes(buffer[index:], routeMatrix.RelayPublicKeys[i])
	}

	numDatacenters := int32(len(routeMatrix.DatacenterRelays))
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numDatacenters))
	index += 4

	for k, v := range routeMatrix.DatacenterRelays {

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

	for i := range routeMatrix.Entries {

		binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].DirectRTT))
		index += 4

		binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].NumRoutes))
		index += 4

		for j := 0; j < int(routeMatrix.Entries[i].NumRoutes); j++ {

			binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].RouteRTT[j]))
			index += 4

			binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].RouteNumRelays[j]))
			index += 4

			for k := 0; k < int(routeMatrix.Entries[i].RouteNumRelays[j]); k++ {
				binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].RouteRelays[j][k]))
				index += 4
			}
		}
	}

	return buffer[:index]
}

func ReadRouteMatrix(buffer []byte) (*RouteMatrix, error) {

	var index int

	var routeMatrix RouteMatrix

	// todo: update to new and better way to read/write binary

	version := binary.LittleEndian.Uint32(buffer[index:])
	index += 4

	if version > RouteMatrixVersion {
		return nil, fmt.Errorf("unknown route matrix version: %d", version)
	}

	var numRelays int32
	numRelays = int32(binary.LittleEndian.Uint32(buffer[index:]))
	index += 4

	routeMatrix.RelayIds = make([]RelayId, numRelays)
	for i := 0; i < int(numRelays); i++ {
		routeMatrix.RelayIds[i] = RelayId(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4
	}

	var bytes_read int

	routeMatrix.RelayNames = make([]string, numRelays)
	if version >= 1 {
		for i := range routeMatrix.RelayNames {
			routeMatrix.RelayNames[i], bytes_read = ReadString(buffer[index:])
			index += bytes_read
		}
	}

	if version >= 2 {
		datacenterCount := binary.LittleEndian.Uint32(buffer[index:])
		index += 4

		routeMatrix.DatacenterIds = make([]DatacenterId, datacenterCount)
		routeMatrix.DatacenterNames = make([]string, datacenterCount)
		for i := 0; i < int(datacenterCount); i++ {
			routeMatrix.DatacenterIds[i] = DatacenterId(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4
			routeMatrix.DatacenterNames[i], bytes_read = ReadString(buffer[index:])
			index += bytes_read
		}
	}

	routeMatrix.RelayAddresses = make([][]byte, numRelays)
	for i := range routeMatrix.RelayAddresses {
		routeMatrix.RelayAddresses[i], bytes_read = ReadBytes(buffer[index:])
		index += bytes_read
	}

	routeMatrix.RelayPublicKeys = make([][]byte, numRelays)
	for i := range routeMatrix.RelayPublicKeys {
		routeMatrix.RelayPublicKeys[i], bytes_read = ReadBytes(buffer[index:])
		index += bytes_read
	}

	numDatacenters := int32(binary.LittleEndian.Uint32(buffer[index:]))
	index += 4

	routeMatrix.DatacenterRelays = make(map[DatacenterId][]RelayId)

	for i := 0; i < int(numDatacenters); i++ {

		datacenterId := DatacenterId(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		numRelaysInDatacenter := int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		routeMatrix.DatacenterRelays[datacenterId] = make([]RelayId, numRelaysInDatacenter)

		for j := 0; j < int(numRelaysInDatacenter); j++ {
			routeMatrix.DatacenterRelays[datacenterId][j] = RelayId(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4
		}
	}

	entryCount := TriMatrixLength(int(numRelays))

	routeMatrix.Entries = make([]RouteMatrixEntry, entryCount)

	for i := range routeMatrix.Entries {

		routeMatrix.Entries[i].DirectRTT = int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		routeMatrix.Entries[i].NumRoutes = int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		for j := 0; j < int(routeMatrix.Entries[i].NumRoutes); j++ {

			routeMatrix.Entries[i].RouteRTT[j] = int32(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4

			routeMatrix.Entries[i].RouteNumRelays[j] = int32(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4

			for k := 0; k < int(routeMatrix.Entries[i].RouteNumRelays[j]); k++ {
				routeMatrix.Entries[i].RouteRelays[j][k] = binary.LittleEndian.Uint32(buffer[index:])
				index += 4
			}
		}
	}

	return &routeMatrix, nil
}
