package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/chacha20poly1305"
	"hash/fnv"
	"net"
	"os"
	"strconv"
	"time"
)

const MaxPacketBytes = 1384

const RouteTokenBytes = 71
const EncryptedRouteTokenBytes = 111

const ContinueTokenBytes = 17
const EncryptedContinueTokenBytes = 57

const (
	ADDRESS_NONE = 0
	ADDRESS_IPV4 = 1
	ADDRESS_IPV6 = 2
)

func ParseAddress(input string) net.UDPAddr {
	address := net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err != nil {
		address.IP = net.ParseIP(input)
		address.Port = 0
		return address
	}
	address.IP = net.ParseIP(ip_string)
	address.Port, _ = strconv.Atoi(port_string)
	return address
}

func GetEnvAddress(name string, defaultValue string) net.UDPAddr {
	string, ok := os.LookupEnv(name)
	if !ok {
		return ParseAddress(defaultValue)
	}
	return ParseAddress(string)
}

func GetEnvInt(name string, defaultValue int) int {
	string, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := strconv.ParseInt(string, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("env string is not an integer: %s\n", name))
	}
	return int(value)
}

func GeneratePittle(output []byte, fromAddress []byte, toAddress []byte, packetLength int) {

	var packetLengthData [2]byte
	binary.LittleEndian.PutUint16(packetLengthData[:], uint16(packetLength))

	sum := uint16(0)

	for i := 0; i < 4; i++ {
		sum += uint16(fromAddress[i])
	}

	for i := 0; i < 4; i++ {
		sum += uint16(toAddress[i])
	}

	sum += uint16(packetLengthData[0])
	sum += uint16(packetLengthData[1])

	var sumData [2]byte
	binary.LittleEndian.PutUint16(sumData[:], sum)

	output[0] = 1 | (sumData[0] ^ sumData[1] ^ 193)
	output[1] = 1 | ((255 - output[0]) ^ 113)
}

func GenerateChonkle(output []byte, magic []byte, fromAddressData []byte, toAddressData []byte, packetLength int) {

	var packetLengthData [2]byte
	binary.LittleEndian.PutUint16(packetLengthData[:], uint16(packetLength))

	hash := fnv.New64a()
	hash.Write(magic)
	hash.Write(fromAddressData)
	hash.Write(toAddressData)
	hash.Write(packetLengthData[:])
	hashValue := hash.Sum64()

	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], uint64(hashValue))

	output[0] = ((data[6] & 0xC0) >> 6) + 42
	output[1] = (data[3] & 0x1F) + 200
	output[2] = ((data[2] & 0xFC) >> 2) + 5
	output[3] = data[0]
	output[4] = (data[2] & 0x03) + 78
	output[5] = (data[4] & 0x7F) + 96
	output[6] = ((data[1] & 0xFC) >> 2) + 100

	if (data[7] & 1) == 0 {
		output[7] = 79
	} else {
		output[7] = 7
	}
	if (data[4] & 0x80) == 0 {
		output[8] = 37
	} else {
		output[8] = 83
	}

	output[9] = (data[5] & 0x07) + 124
	output[10] = ((data[1] & 0xE0) >> 5) + 175
	output[11] = (data[6] & 0x3F) + 33

	value := (data[1] & 0x03)
	if value == 0 {
		output[12] = 97
	} else if value == 1 {
		output[12] = 5
	} else if value == 2 {
		output[12] = 43
	} else {
		output[12] = 13
	}

	output[13] = ((data[5] & 0xF8) >> 3) + 210
	output[14] = ((data[7] & 0xFE) >> 1) + 17
}

func GetAddressData(address net.UDPAddr) []byte {
	return address.IP.To4()
}

func BasicPacketFilter(data []byte, packetLength int) bool {

	if packetLength < 18 {
		return false
	}

	if data[0] < 0x01 || data[0] > 0x0E {
		return false
	}

	if data[2] != (1 | ((255 - data[1]) ^ 113)) {
		return false
	}

	if data[3] < 0x2A || data[3] > 0x2D {
		return false
	}

	if data[4] < 0xC8 || data[4] > 0xE7 {
		return false
	}

	if data[5] < 0x05 || data[5] > 0x44 {
		return false
	}

	if data[7] < 0x4E || data[7] > 0x51 {
		return false
	}

	if data[8] < 0x60 || data[8] > 0xDF {
		return false
	}

	if data[9] < 0x64 || data[9] > 0xE3 {
		return false
	}

	if data[10] != 0x07 && data[10] != 0x4F {
		return false
	}

	if data[11] != 0x25 && data[11] != 0x53 {
		return false
	}

	if data[12] < 0x7C || data[12] > 0x83 {
		return false
	}

	if data[13] < 0xAF || data[13] > 0xB6 {
		return false
	}

	if data[14] < 0x21 || data[14] > 0x60 {
		return false
	}

	if data[15] != 0x61 && data[15] != 0x05 && data[15] != 0x2B && data[15] != 0x0D {
		return false
	}

	if data[16] < 0xD2 || data[16] > 0xF1 {
		return false
	}

	if data[17] < 0x11 || data[17] > 0x90 {
		return false
	}

	return true
}

func AdvancedPacketFilter(data []byte, magic []byte, fromAddress []byte, toAddress []byte, packetLength int) bool {
	if packetLength < 18 {
		return false
	}
	var a [2]byte
	var b [15]byte
	GeneratePittle(a[:], fromAddress, toAddress, packetLength)
	GenerateChonkle(b[:], magic, fromAddress, toAddress, packetLength)
	if bytes.Compare(a[0:2], data[1:3]) != 0 {
		return false
	}
	if bytes.Compare(b[0:15], data[3:18]) != 0 {
		return false
	}
	return true
}

func GeneratePingToken(expireTimestamp uint64, from net.UDPAddr, to net.UDPAddr, pingKey []byte) []byte {
	data := make([]byte, 32+20)
	index := 0
	copy(data[index:], pingKey)
	index += 32
	binary.LittleEndian.PutUint64(data[index:], expireTimestamp)
	index += 8
	copy(data[index:], from.IP.To4())
	index += 4
	copy(data[index:], to.IP.To4())
	index += 4
	binary.BigEndian.PutUint16(data[index:], uint16(from.Port))
	index += 2
	binary.BigEndian.PutUint16(data[index:], uint16(to.Port))
	result := sha256.Sum256(data)
	return result[:]
}

func GenerateHeaderTag(packetType uint8, packetSequence uint64, sessionId uint64, sessionVersion uint8, sessionPrivateKey []byte) []byte {
	data := make([]byte, 32+1+8+8+1)
	index := 0
	copy(data[index:], sessionPrivateKey)
	index += 32
	data[index] = packetType
	index += 1
	binary.LittleEndian.PutUint64(data[index:], packetSequence)
	index += 8
	binary.LittleEndian.PutUint64(data[index:], sessionId)
	index += 8
	data[index] = sessionVersion
	result := sha256.Sum256(data)
	return result[0:8]
}

type RouteTokenData struct {
	SessionPrivateKey [32]byte
	ExpireTimestamp   uint64
	SessionId         uint64
	SessionVersion    uint8
	EnvelopeKbpsUp    uint32
	EnvelopeKbpsDown  uint32
	NextAddress       net.UDPAddr
	PrevAddress       net.UDPAddr
	NextInternal      uint8
	PrevInternal      uint8
}

func GenerateRouteToken(secretKey []byte, data RouteTokenData) []byte {

	routeToken := make([]byte, RouteTokenBytes)

	index := 0

	copy(routeToken[index:], data.SessionPrivateKey[:])
	index += 32

	binary.LittleEndian.PutUint64(routeToken[index:], data.ExpireTimestamp)
	index += 8

	binary.LittleEndian.PutUint64(routeToken[index:], data.SessionId)
	index += 8

	binary.LittleEndian.PutUint32(routeToken[index:], data.EnvelopeKbpsUp)
	index += 4

	binary.LittleEndian.PutUint32(routeToken[index:], data.EnvelopeKbpsDown)
	index += 4

	copy(routeToken[index:], GetAddressData(data.NextAddress))
	index += 4

	copy(routeToken[index:], GetAddressData(data.PrevAddress))
	index += 4

	binary.BigEndian.PutUint16(routeToken[index:], uint16(data.NextAddress.Port))
	index += 2

	binary.BigEndian.PutUint16(routeToken[index:], uint16(data.PrevAddress.Port))
	index += 2

	routeToken[index] = data.SessionVersion
	index += 1

	routeToken[index] = data.NextInternal
	index += 1

	routeToken[index] = data.PrevInternal
	index += 1

	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(routeToken)+aead.Overhead())
	if _, err := rand.Read(nonce[:aead.NonceSize()]); err != nil {
		panic(err)
	}

	encryptedRouteToken := aead.Seal(nonce, nonce, routeToken, nil)

	return encryptedRouteToken
}

type ContinueTokenData struct {
	ExpireTimestamp uint64
	SessionId       uint64
	SessionVersion  uint8
}

func GenerateContinueToken(secretKey []byte, data ContinueTokenData) []byte {

	continueToken := make([]byte, ContinueTokenBytes)

	index := 0

	binary.LittleEndian.PutUint64(continueToken[index:], data.ExpireTimestamp)
	index += 8

	binary.LittleEndian.PutUint64(continueToken[index:], data.SessionId)
	index += 8

	continueToken[index] = data.SessionVersion
	index += 1

	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(continueToken)+aead.Overhead())
	if _, err := rand.Read(nonce[:aead.NonceSize()]); err != nil {
		panic(err)
	}

	encryptedContinueToken := aead.Seal(nonce, nonce, continueToken, nil)

	return encryptedContinueToken
}

// ----------------------------------------------------------------------------------------------------

func CreateClientToServerPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte, sequenceNumber uint64) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	payload := make([]byte, 1000)

	packetType := uint8(3) // client to server packet

	binary.LittleEndian.PutUint64(payload[0:8], sequenceNumber)

	sessionPrivateKey := make([]byte, 32)
	for i := range sessionPrivateKey {
		sessionPrivateKey[i] = 100 + uint8(i)
	}

	sessionId := uint64(0x12345)
	binary.LittleEndian.PutUint64(payload[8:16], sessionId)

	sessionVersion := uint8(1)
	payload[8+8] = sessionVersion

	tag := GenerateHeaderTag(packetType, sequenceNumber, sessionId, sessionVersion, sessionPrivateKey)
	copy(payload[8+8+1:], tag)

	packet := make([]byte, 18+len(payload))

	packet[0] = packetType

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestClientToServerPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, secretKey []byte) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	fmt.Printf("sent client ping packet\n")

	packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

	conn.WriteToUDP(packet, &serverAddress)

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route request packet\n")

		packet := CreateRouteRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent client to server packet\n")

		packet := CreateClientToServerPacket(clientAddress, serverAddress, magic, uint64(1+i))

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateServerToClientPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte, sequenceNumber uint64) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	payload := make([]byte, 1000)

	packetType := uint8(4) // server to client packet

	binary.LittleEndian.PutUint64(payload[0:8], sequenceNumber)

	sessionPrivateKey := make([]byte, 32)
	for i := range sessionPrivateKey {
		sessionPrivateKey[i] = 100 + uint8(i)
	}

	sessionId := uint64(0x12345)
	binary.LittleEndian.PutUint64(payload[8:16], sessionId)

	sessionVersion := uint8(1)
	payload[8+8] = sessionVersion

	tag := GenerateHeaderTag(packetType, sequenceNumber, sessionId, sessionVersion, sessionPrivateKey)
	copy(payload[8+8+1:], tag)

	packet := make([]byte, 18+len(payload))

	packet[0] = packetType

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestServerToClientPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, secretKey []byte) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	fmt.Printf("sent client ping packet\n")

	packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

	conn.WriteToUDP(packet, &serverAddress)

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route request packet\n")

		packet := CreateRouteRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent server to client packet\n")

		packet := CreateServerToClientPacket(clientAddress, serverAddress, magic, uint64(1+i))

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateRelayPingPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	currentTime := uint64(time.Now().Unix())

	expireTimestamp := currentTime + 10

	pingKey, _ := base64.StdEncoding.DecodeString("ANXbw47AaWuu7sidkuw0Cq5cIXU4e8xoqJbSsFC+MT0=")

	if len(pingKey) != 32 {
		panic("ping key should be 32 bytes long")
	}

	pingToken := GeneratePingToken(expireTimestamp, clientAddress, serverAddress, pingKey)

	if len(pingToken) != 32 {
		panic("ping token should be 32 bytes long")
	}

	payload := make([]byte, 8+8+1+32)

	binary.LittleEndian.PutUint64(payload[8:], expireTimestamp)

	copy(payload[1+8+8:], pingToken)

	packet := make([]byte, 18+len(payload))

	packet[0] = 11 // relay ping packet

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestRelayPingPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent relay ping packet packet to %s\n", serverAddress.String())

		packet := CreateRelayPingPacket(clientAddress, serverAddress, magic)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateClientPingPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	currentTime := uint64(time.Now().Unix())

	expireTimestamp := currentTime + 10

	pingKey, _ := base64.StdEncoding.DecodeString("ANXbw47AaWuu7sidkuw0Cq5cIXU4e8xoqJbSsFC+MT0=")

	if len(pingKey) != 32 {
		panic("ping key should be 32 bytes long")
	}

	// IMPORTANT: Some NAT change the client port from what is expected, so set the client port to 0 in the ping token
	clientAddressWithZeroPort := clientAddress
	clientAddressWithZeroPort.Port = 0

	pingToken := GeneratePingToken(expireTimestamp, clientAddressWithZeroPort, serverAddress, pingKey)

	if len(pingToken) != 32 {
		panic("ping token should be 32 bytes long")
	}

	payload := make([]byte, 8+8+8+32)

	binary.LittleEndian.PutUint64(payload[8+8:], expireTimestamp)

	copy(payload[8+8+8:], pingToken)

	packet := make([]byte, 18+len(payload))

	packet[0] = 9 // client ping packet

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestClientPingPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent client ping packet\n")

		packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateServerPingPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	currentTime := uint64(time.Now().Unix())

	expireTimestamp := currentTime + 10

	pingKey, _ := base64.StdEncoding.DecodeString("ANXbw47AaWuu7sidkuw0Cq5cIXU4e8xoqJbSsFC+MT0=")

	if len(pingKey) != 32 {
		panic("ping key should be 32 bytes long")
	}

	pingToken := GeneratePingToken(expireTimestamp, clientAddress, serverAddress, pingKey)

	if len(pingToken) != 32 {
		panic("ping token should be 32 bytes long")
	}

	payload := make([]byte, 8+8+32)

	binary.LittleEndian.PutUint64(payload[8:], expireTimestamp)

	copy(payload[8+8:], pingToken)

	packet := make([]byte, 18+len(payload))

	packet[0] = 13 // server ping packet

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestServerPingPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent server ping packet\n")

		packet := CreateServerPingPacket(clientAddress, serverAddress, magic)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateSessionPingPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte, sequenceNumber uint64) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	payload := make([]byte, 25)

	packetType := uint8(5) // session ping packet

	binary.LittleEndian.PutUint64(payload[0:8], sequenceNumber)

	sessionPrivateKey := make([]byte, 32)
	for i := range sessionPrivateKey {
		sessionPrivateKey[i] = 100 + uint8(i)
	}

	sessionId := uint64(0x12345)
	binary.LittleEndian.PutUint64(payload[8:16], sessionId)

	sessionVersion := uint8(1)
	payload[8+8] = sessionVersion

	tag := GenerateHeaderTag(packetType, sequenceNumber, sessionId, sessionVersion, sessionPrivateKey)
	copy(payload[8+8+1:], tag)

	packet := make([]byte, 18+len(payload))

	packet[0] = packetType

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestSessionPingPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, secretKey []byte) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	fmt.Printf("sent client ping packet\n")

	packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

	conn.WriteToUDP(packet, &serverAddress)

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route request packet\n")

		packet := CreateRouteRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent session ping packet\n")

		packet := CreateSessionPingPacket(clientAddress, serverAddress, magic, uint64(1+i))

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateSessionPongPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte, sequenceNumber uint64) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	payload := make([]byte, 25)

	packetType := uint8(6) // session pong packet

	binary.LittleEndian.PutUint64(payload[0:8], sequenceNumber)

	sessionPrivateKey := make([]byte, 32)
	for i := range sessionPrivateKey {
		sessionPrivateKey[i] = 100 + uint8(i)
	}

	sessionId := uint64(0x12345)
	binary.LittleEndian.PutUint64(payload[8:16], sessionId)

	sessionVersion := uint8(1)
	payload[8+8] = sessionVersion

	tag := GenerateHeaderTag(packetType, sequenceNumber, sessionId, sessionVersion, sessionPrivateKey)
	copy(payload[8+8+1:], tag)

	packet := make([]byte, 18+len(payload))

	packet[0] = packetType

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestSessionPongPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, secretKey []byte) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	fmt.Printf("sent client ping packet\n")

	packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

	conn.WriteToUDP(packet, &serverAddress)

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route request packet\n")

		packet := CreateRouteRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent session pong packet\n")

		packet := CreateSessionPongPacket(clientAddress, serverAddress, magic, uint64(1+i))

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateRouteRequestPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte, secretKey []byte) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	payload := make([]byte, EncryptedRouteTokenBytes*2)

	packetType := uint8(1) // route request packet

	nextAddress := clientAddress
	prevAddress := clientAddress
	prevAddress.Port = 0

	sessionPrivateKey := make([]byte, 32)
	for i := range sessionPrivateKey {
		sessionPrivateKey[i] = 100 + uint8(i)
	}

	data := RouteTokenData{
		ExpireTimestamp: uint64(time.Now().Unix() + 30),
		SessionId:       0x12345,
		SessionVersion:  1,
		NextAddress:     nextAddress,
		PrevAddress:     prevAddress,
	}

	copy(data.SessionPrivateKey[:], sessionPrivateKey)

	routeToken := GenerateRouteToken(secretKey, data)

	copy(payload[0:EncryptedRouteTokenBytes], routeToken)
	copy(payload[EncryptedRouteTokenBytes:EncryptedRouteTokenBytes*2], routeToken)

	packet := make([]byte, 18+len(payload))

	packet[0] = packetType

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestRouteRequestPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, secretKey []byte) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	fmt.Printf("sent client ping packet\n")

	packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

	conn.WriteToUDP(packet, &serverAddress)

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route request packet\n")

		packet := CreateRouteRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateContinueRequestPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte, secretKey []byte) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	payload := make([]byte, EncryptedContinueTokenBytes*2)

	packetType := uint8(7) // continue request packet

	data := ContinueTokenData{
		ExpireTimestamp: uint64(time.Now().Unix() + 30),
		SessionId:       0x12345,
		SessionVersion:  1,
	}

	continueToken := GenerateContinueToken(secretKey, data)

	copy(payload[0:EncryptedContinueTokenBytes], continueToken)
	copy(payload[EncryptedContinueTokenBytes:EncryptedContinueTokenBytes*2], continueToken)

	packet := make([]byte, 18+len(payload))

	packet[0] = packetType

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestContinueRequestPacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, secretKey []byte) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	fmt.Printf("sent client ping packet\n")

	packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

	conn.WriteToUDP(packet, &serverAddress)

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route request packet\n")

		packet := CreateRouteRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent continue request packet\n")

		packet := CreateContinueRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateRouteResponsePacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte, sequenceNumber uint64) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	payload := make([]byte, 25)

	packetType := uint8(2) // route response packet

	binary.LittleEndian.PutUint64(payload[0:8], sequenceNumber)

	sessionPrivateKey := make([]byte, 32)
	for i := range sessionPrivateKey {
		sessionPrivateKey[i] = 100 + uint8(i)
	}

	sessionId := uint64(0x12345)
	binary.LittleEndian.PutUint64(payload[8:16], sessionId)

	sessionVersion := uint8(1)
	payload[8+8] = sessionVersion

	tag := GenerateHeaderTag(packetType, sequenceNumber, sessionId, sessionVersion, sessionPrivateKey)
	copy(payload[8+8+1:], tag)

	packet := make([]byte, 18+len(payload))

	packet[0] = packetType

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestRouteResponsePacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, secretKey []byte) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	fmt.Printf("sent client ping packet\n")

	packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

	conn.WriteToUDP(packet, &serverAddress)

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route request packet\n")

		packet := CreateRouteRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route response packet\n")

		packet := CreateRouteResponsePacket(clientAddress, serverAddress, magic, uint64(1+i))

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func CreateContinueResponsePacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, magic []byte, sequenceNumber uint64) []byte {

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	payload := make([]byte, 25)

	packetType := uint8(8) // continue response packet

	binary.LittleEndian.PutUint64(payload[0:8], sequenceNumber)

	sessionPrivateKey := make([]byte, 32)
	for i := range sessionPrivateKey {
		sessionPrivateKey[i] = 100 + uint8(i)
	}

	sessionId := uint64(0x12345)
	binary.LittleEndian.PutUint64(payload[8:16], sessionId)

	sessionVersion := uint8(1)
	payload[8+8] = sessionVersion

	tag := GenerateHeaderTag(packetType, sequenceNumber, sessionId, sessionVersion, sessionPrivateKey)
	copy(payload[8+8+1:], tag)

	packet := make([]byte, 18+len(payload))

	packet[0] = packetType

	a := packet[1:3]

	b := packet[3:18]

	GeneratePittle(a, fromAddressData, toAddressData, len(packet))

	GenerateChonkle(b, magic, fromAddressData, toAddressData, len(packet))

	copy(packet[18:], payload)

	return packet
}

func TestContinueResponsePacket(clientAddress net.UDPAddr, serverAddress net.UDPAddr, secretKey []byte) {

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:30000")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	magic := make([]byte, 8)
	for i := 0; i < 8; i++ {
		magic[i] = uint8(i)
	}

	fmt.Printf("sent client ping packet\n")

	packet := CreateClientPingPacket(clientAddress, serverAddress, magic)

	conn.WriteToUDP(packet, &serverAddress)

	for i := 0; i < 10; i++ {

		fmt.Printf("sent route request packet\n")

		packet := CreateRouteRequestPacket(clientAddress, serverAddress, magic, secretKey)

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	for i := 0; i < 10; i++ {

		fmt.Printf("sent continue response packet\n")

		packet := CreateContinueResponsePacket(clientAddress, serverAddress, magic, uint64(1+i))

		conn.WriteToUDP(packet, &serverAddress)

		time.Sleep(time.Millisecond * 100)
	}

	fromAddressData := GetAddressData(clientAddress)
	toAddressData := GetAddressData(serverAddress)

	for {

		buffer := make([]byte, MaxPacketBytes)

		size, from, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		if !BasicPacketFilter(buffer[:size], size) {
			fmt.Printf("basic packet filter failed\n")
			continue
		}

		if !AdvancedPacketFilter(buffer[:size], magic, toAddressData, fromAddressData, size) {
			fmt.Printf("advanced packet filter failed\n")
			continue
		}

		fmt.Printf("received %d byte packet type %d from %s\n", size, buffer[0], from.String())
	}
}

// ----------------------------------------------------------------------------------------------------

func main() {

	// clientAddress := GetEnvAddress("CLIENT_ADDRESS", "192.168.1.34:30000")
	clientAddress := GetEnvAddress("CLIENT_ADDRESS", "192.168.1.11:30000")
	serverAddress := GetEnvAddress("SERVER_ADDRESS", "192.168.1.40:40000")

	secretKey := []byte{0x22, 0x3c, 0x0c, 0xc6, 0x70, 0x7b, 0x99, 0xc4, 0xdd, 0x44, 0xb9, 0xe8, 0x3c, 0x78, 0x1c, 0xd7, 0xd3, 0x2f, 0x9b, 0xad, 0x70, 0xbf, 0x8d, 0x9f, 0xe3, 0xa6, 0xd4, 0xc7, 0xe3, 0xb2, 0x98, 0x90}

	_ = secretKey

	// TestRelayPingPacket(clientAddress, serverAddress)
	// TestClientPingPacket(clientAddress, serverAddress)
	// TestServerPingPacket(clientAddress, serverAddress)
	// TestRouteRequestPacket(clientAddress, serverAddress, secretKey)
	// TestContinueRequestPacket(clientAddress, serverAddress, secretKey)

	TestClientToServerPacket(clientAddress, serverAddress, secretKey)
	// TestServerToClientPacket(clientAddress, serverAddress, secretKey)
	// TestSessionPingPacket(clientAddress, serverAddress, secretKey)
	// TestSessionPongPacket(clientAddress, serverAddress, secretKey)
	// TestRouteResponsePacket(clientAddress, serverAddress, secretKey)
	// TestContinueResponsePacket(clientAddress, serverAddress, secretKey)
}
