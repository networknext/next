package encoding

import (
	"encoding/binary"
	"math"
)

func ReadUint32(data []byte, index *int, value *uint32) bool {
	if *index+4 > len(data) {
		return false
	}
	*value = binary.LittleEndian.Uint32(data[*index:])
	*index += 4
	return true
}

func ReadUint64(data []byte, index *int, value *uint64) bool {
	if *index+8 > len(data) {
		return false
	}
	*value = binary.LittleEndian.Uint64(data[*index:])
	*index += 8
	return true
}

func ReadFloat32(data []byte, index *int, value *float32) bool {
	var int_value uint32
	if !ReadUint32(data, index, &int_value) {
		return false
	}
	*value = math.Float32frombits(int_value)
	return true
}

func ReadString(data []byte, index *int, value *string, maxStringLength uint32) bool {
	var stringLength uint32
	if !ReadUint32(data, index, &stringLength) {
		return false
	}
	if stringLength > maxStringLength {
		return false
	}
	if *index+int(stringLength) > len(data) {
		return false
	}
	stringData := make([]byte, stringLength)
	for i := uint32(0); i < stringLength; i++ {
		stringData[i] = data[*index]
		*index++
	}
	*value = string(stringData)
	return true
}

// used for CostMatrix & RouteMatrix unmarshaling. needed for when version < 3
func ReadStringOld(buffer []byte) (string, int) {
	stringLength := binary.LittleEndian.Uint32(buffer)
	stringData := make([]byte, stringLength)
	copy(stringData, buffer[4:4+stringLength])
	return string(stringData), int(4 + stringLength)
}

func ReadBytes(data []byte, index *int, value *[]byte, bytes uint32) bool {
	if *index+int(bytes) > len(data) {
		return false
	}
	*value = make([]byte, bytes)
	for i := uint32(0); i < bytes; i++ {
		(*value)[i] = data[*index]
		*index++
	}
	return true
}

// used for CostMatrix & RouteMatrix unmarshaling. needed for when version < 3
func ReadBytesOld(buffer []byte) ([]byte, int) {
	length := binary.LittleEndian.Uint32(buffer)
	data := make([]byte, length)
	copy(data, buffer[4:4+length])
	return data, int(4 + length)
}
