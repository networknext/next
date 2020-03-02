package core

import (
	"encoding/binary"
	"fmt"
)

type CostMatrix struct {
	RelayIDs         []RelayID
	RelayNames       []string
	RelayAddresses   [][]byte
	RelayPublicKeys  [][]byte
	DatacenterIDs    []DatacenterID
	DatacenterNames  []string
	DatacenterRelays map[DatacenterID][]RelayID
	RTT              []int32
}

// IMPORTANT: Bump this version whenever you change the binary format
const CostMatrixVersion = 2

func WriteCostMatrix(buffer []byte, costMatrix *CostMatrix) []byte {

	var index int

	// todo: update this to the new way of reading/writing binary as per-backend.go

	binary.LittleEndian.PutUint32(buffer[index:], CostMatrixVersion)
	index += 4

	numRelays := len(costMatrix.RelayIDs)
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelays))
	index += 4

	for i := range costMatrix.RelayIDs {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(costMatrix.RelayIDs[i]))
		index += 4
	}

	for i := range costMatrix.RelayNames {
		index += WriteString(buffer[index:], costMatrix.RelayNames[i])
	}

	if len(costMatrix.DatacenterIDs) != len(costMatrix.DatacenterNames) {
		panic("datacenter ids length does not match datacenter names length")
	}

	binary.LittleEndian.PutUint32(buffer[index:], uint32(len(costMatrix.DatacenterIDs)))
	index += 4

	for i := 0; i < len(costMatrix.DatacenterIDs); i++ {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(costMatrix.DatacenterIDs[i]))
		index += 4
		index += WriteString(buffer[index:], costMatrix.DatacenterNames[i])
	}

	for i := range costMatrix.RelayAddresses {
		index += WriteBytes(buffer[index:], costMatrix.RelayAddresses[i])
	}

	for i := range costMatrix.RelayPublicKeys {
		index += WriteBytes(buffer[index:], costMatrix.RelayPublicKeys[i])
	}

	numDatacenters := int32(len(costMatrix.DatacenterRelays))
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numDatacenters))
	index += 4

	for k, v := range costMatrix.DatacenterRelays {

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

	for i := range costMatrix.RTT {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(costMatrix.RTT[i]))
		index += 4
	}

	return buffer[:index]
}

func ReadCostMatrix(buffer []byte) (*CostMatrix, error) {

	var index int

	var costMatrix CostMatrix

	// todo: update to new way to read/write binary as per backend.go

	version := binary.LittleEndian.Uint32(buffer[index:])
	index += 4

	if version > CostMatrixVersion {
		return nil, fmt.Errorf("unknown cost matrix version %d", version)
	}

	numRelays := int32(binary.LittleEndian.Uint32(buffer[index:]))
	index += 4

	costMatrix.RelayIDs = make([]RelayID, numRelays)
	for i := 0; i < int(numRelays); i++ {
		costMatrix.RelayIDs[i] = RelayID(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4
	}

	var bytes_read int

	costMatrix.RelayNames = make([]string, numRelays)
	if version >= 1 {
		for i := range costMatrix.RelayNames {
			costMatrix.RelayNames[i], bytes_read = ReadString(buffer[index:])
			index += bytes_read
		}
	}

	if version >= 2 {
		datacenterCount := binary.LittleEndian.Uint32(buffer[index:])
		index += 4

		costMatrix.DatacenterIDs = make([]DatacenterID, datacenterCount)
		costMatrix.DatacenterNames = make([]string, datacenterCount)
		for i := 0; i < int(datacenterCount); i++ {
			costMatrix.DatacenterIDs[i] = DatacenterID(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4
			costMatrix.DatacenterNames[i], bytes_read = ReadString(buffer[index:])
			index += bytes_read
		}
	}

	costMatrix.RelayAddresses = make([][]byte, numRelays)
	for i := range costMatrix.RelayAddresses {
		costMatrix.RelayAddresses[i], bytes_read = ReadBytes(buffer[index:])
		index += bytes_read
	}

	costMatrix.RelayPublicKeys = make([][]byte, numRelays)
	for i := range costMatrix.RelayPublicKeys {
		costMatrix.RelayPublicKeys[i], bytes_read = ReadBytes(buffer[index:])
		index += bytes_read
	}

	numDatacenters := int32(binary.LittleEndian.Uint32(buffer[index:]))
	index += 4

	costMatrix.DatacenterRelays = make(map[DatacenterID][]RelayID)

	for i := 0; i < int(numDatacenters); i++ {

		datacenterID := DatacenterID(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		numRelaysInDatacenter := int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		costMatrix.DatacenterRelays[datacenterID] = make([]RelayID, numRelaysInDatacenter)

		for j := 0; j < int(numRelaysInDatacenter); j++ {
			costMatrix.DatacenterRelays[datacenterID][j] = RelayID(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4
		}
	}

	entryCount := TriMatrixLength(int(numRelays))
	costMatrix.RTT = make([]int32, entryCount)
	for i := range costMatrix.RTT {
		costMatrix.RTT[i] = int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4
	}

	return &costMatrix, nil
}
