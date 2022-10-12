package encoding

import (
	"encoding/binary"
	"math"
	"net"
)

const (
	IPAddressNone = 0
	IPAddressIPv4 = 1
	IPAddressIPv6 = 2
	AddressSize   = 19
)

func WriteBool(data []byte, index *int, value bool) {
	if value {
		data[*index] = byte(1)
	} else {
		data[*index] = byte(0)
	}

	*index += 1
}

func WriteUint8(data []byte, index *int, value uint8) {
	data[*index] = byte(value)
	*index += 1
}

func WriteUint16(data []byte, index *int, value uint16) {
	binary.LittleEndian.PutUint16(data[*index:], value)
	*index += 2
}

func WriteUint32(data []byte, index *int, value uint32) {
	binary.LittleEndian.PutUint32(data[*index:], value)
	*index += 4
}

func WriteUint64(data []byte, index *int, value uint64) {
	binary.LittleEndian.PutUint64(data[*index:], value)
	*index += 8
}

func WriteInt(data []byte, index *int, value int) {
	binary.LittleEndian.PutUint64(data[*index:], uint64(value))
	*index += 8
}

func WriteFloat32(data []byte, index *int, value float32) {
	uintValue := math.Float32bits(value)
	WriteUint32(data, index, uintValue)
}

func WriteFloat64(data []byte, index *int, value float64) {
	uintValue := math.Float64bits(value)
	WriteUint64(data, index, uintValue)
}

func WriteString(data []byte, index *int, value string, maxStringLength uint32) {
	stringLength := uint32(len(value))
	if stringLength > maxStringLength {
		panic("string is too long!\n")
	}
	binary.LittleEndian.PutUint32(data[*index:], stringLength)
	*index += 4
	for i := 0; i < int(stringLength); i++ {
		data[*index] = value[i]
		*index++
	}
}

func WriteBytes(data []byte, index *int, value []byte, numBytes int) {
	for i := 0; i < numBytes; i++ {
		data[*index] = value[i]
		*index++
	}
}

func WriteAddress(data []byte, index *int, address *net.UDPAddr) {
	if address == nil {
		data[*index] = IPAddressNone
		*index += 1
		return
	}
	ipv4 := address.IP.To4()
	port := address.Port
	if ipv4 != nil {
		data[*index] = IPAddressIPv4
		data[*index+1] = ipv4[0]
		data[*index+2] = ipv4[1]
		data[*index+3] = ipv4[2]
		data[*index+4] = ipv4[3]
		data[*index+5] = (byte)(port & 0xFF)
		data[*index+6] = (byte)(port >> 8)
		*index += 7
	} else {
		data[0] = IPAddressIPv6
		copy(data[*index+1:], address.IP)
		data[*index+17] = (byte)(port & 0xFF)
		data[*index+18] = (byte)(port >> 8)
		*index += 19
	}
}

func ReadBool(data []byte, index *int, value *bool) bool {

	if *index+1 > len(data) {
		return false
	}

	if data[*index] > 0 {
		*value = true
	} else {
		*value = false
	}

	*index += 1
	return true
}

func ReadUint8(data []byte, index *int, value *uint8) bool {
	if *index+1 > len(data) {
		return false
	}
	*value = data[*index]
	*index += 1
	return true
}

func ReadUint16(data []byte, index *int, value *uint16) bool {
	if *index+2 > len(data) {
		return false
	}
	*value = binary.LittleEndian.Uint16(data[*index:])
	*index += 2
	return true
}

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

func ReadInt(data []byte, index *int, value *int) bool {
	if *index+8 > len(data) {
		return false
	}
	*value = int(binary.LittleEndian.Uint64(data[*index:]))
	*index += 8
	return true
}

func ReadFloat32(data []byte, index *int, value *float32) bool {
	var intValue uint32
	if !ReadUint32(data, index, &intValue) {
		return false
	}
	*value = math.Float32frombits(intValue)
	return true
}

func ReadFloat64(data []byte, index *int, value *float64) bool {
	var uintValue uint64
	if !ReadUint64(data, index, &uintValue) {
		return false
	}
	*value = math.Float64frombits(uintValue)
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

func ReadAddress(data []byte, index *int, address *net.UDPAddr) bool {
	addressType := data[*index]
	switch addressType {
	case IPAddressNone:
		*address = net.UDPAddr{}
		*index += 1
	case IPAddressIPv4:
		if *index+7 > len(data) {
			return false
		}
		*address = net.UDPAddr{IP: net.IPv4(data[*index+1], data[*index+2], data[*index+3], data[*index+4]), Port: ((int)(binary.LittleEndian.Uint16(data[*index+5:])))}
		*index += 7
		return true
	case IPAddressIPv6:
		if *index+19 > len(data) {
			return false
		}
		*address = net.UDPAddr{IP: data[*index+1:], Port: ((int)(binary.LittleEndian.Uint16(data[*index+17:])))}
		*index += 19
		return true
	}
	return false
}
