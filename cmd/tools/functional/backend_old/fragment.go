package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

func RandomBytes(buffer []byte) {
	C.randombytes_buf(unsafe.Pointer(&buffer[0]), C.size_t(len(buffer)))
}

const FragmentSize = 1024
const FragmentMax = 255
const FragmentTimeout = 4 * time.Second

type UDPPacketToMaster struct {
	Type uint8
	ID   uint64
	Data []byte
}

type UDPPacketToClient struct {
	Type   uint8
	ID     uint64
	Status uint16
	Data   []byte
}

const AddressBytes = 19
const MasterTokenBytes = AddressBytes + 32

type masterToken struct {
	Address net.UDPAddr
	HMAC    []byte
}

func masterTokenCreate(addr net.UDPAddr, signPrivateKey []byte) (*masterToken, error) {
	token := &masterToken{
		Address: addr,
	}
	var err error
	token.HMAC, err = cryptoAuth(writeAddress(&token.Address), signPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign master token: %v", err)
	}
	return token, nil
}

func (token *masterToken) Write() []byte {
	buffer := make([]byte, MasterTokenBytes)
	copy(buffer, writeAddress(&token.Address))
	copy(buffer[AddressBytes:], token.HMAC)
	return buffer
}

func masterTokenRead(buffer []byte, signPrivateKey []byte) *masterToken {
	if len(buffer) < MasterTokenBytes {
		return nil
	}

	address := readAddress(buffer)
	if address == nil {
		return nil
	}

	hmac := buffer[AddressBytes:]

	if !cryptoAuthVerify(buffer[0:AddressBytes], hmac, signPrivateKey) {
		return nil
	}

	return &masterToken{
		Address: *address,
		HMAC:    hmac,
	}
}

const (
	addressNone = 0
	addressIPv4 = 1
	addressIPv6 = 2
)

func writeAddress(address *net.UDPAddr) []byte {
	buffer := make([]byte, AddressBytes)
	if address == nil {
		buffer[0] = addressNone
		return buffer
	}
	ipv4 := address.IP.To4()
	port := address.Port
	if ipv4 != nil {
		buffer[0] = addressIPv4
		buffer[1] = ipv4[0]
		buffer[2] = ipv4[1]
		buffer[3] = ipv4[2]
		buffer[4] = ipv4[3]
		buffer[5] = (byte)(port & 0xFF)
		buffer[6] = (byte)(port >> 8)
	} else {
		buffer[0] = addressIPv6
		copy(buffer[1:], address.IP)
		buffer[17] = (byte)(port & 0xFF)
		buffer[18] = (byte)(port >> 8)
	}
	return buffer
}

func readAddress(buffer []byte) *net.UDPAddr {
	addressType := buffer[0]
	if addressType == addressIPv4 {
		return &net.UDPAddr{IP: net.IPv4(buffer[1], buffer[2], buffer[3], buffer[4]), Port: ((int)(binary.LittleEndian.Uint16(buffer[5:])))}
	} else if addressType == addressIPv6 {
		return &net.UDPAddr{IP: buffer[1:], Port: ((int)(binary.LittleEndian.Uint16(buffer[17:])))}
	}
	return nil
}

func decompress(packetData []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(packetData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress zipped data: %v", err)
	}
	return decompressed, nil
}

func cryptoAuth(data []byte, key []byte) ([]byte, error) {
	if len(key) != C.crypto_auth_KEYBYTES {
		return nil, fmt.Errorf("expected %d byte key, got %d bytes", C.crypto_auth_KEYBYTES, len(key))
	}
	var signature [C.crypto_auth_BYTES]byte
	if C.crypto_auth((*C.uchar)(&signature[0]), (*C.uchar)(&data[0]), (C.ulonglong)(len(data)), (*C.uchar)(&key[0])) != 0 {
		return nil, fmt.Errorf("failed to sign data with key")
	}
	return signature[:], nil
}

func cryptoAuthVerify(data []byte, signature []byte, key []byte) bool {
	if len(key) != C.crypto_auth_KEYBYTES || len(signature) != C.crypto_auth_BYTES {
		return false
	}
	return C.crypto_auth_verify((*C.uchar)(&signature[0]), (*C.uchar)(&data[0]), (C.ulonglong)(len(data)), (*C.uchar)(&key[0])) == 0
}

func cryptoBoxSeal(data []byte, publicKey []byte) ([]byte, error) {
	if len(publicKey) < C.crypto_box_PUBLICKEYBYTES {
		return nil, fmt.Errorf("invalid crypto_box_seal public key")
	}
	encrypted := make([]byte, len(data)+C.crypto_box_SEALBYTES)
	if C.crypto_box_seal((*C.uchar)(&encrypted[0]), (*C.uchar)(&data[0]), C.ulonglong(len(data)), (*C.uchar)(&publicKey[0])) != 0 {
		return nil, fmt.Errorf("crypto_box_seal failed")
	}
	return encrypted, nil
}

func cryptoBoxSealOpen(data []byte, publicKey []byte, privateKey []byte) ([]byte, error) {
	if len(data) < C.crypto_box_SEALBYTES {
		return nil, fmt.Errorf("length of data less than crypto sealed box bytes header")
	}
	decrypted := make([]byte, len(data)-C.crypto_box_SEALBYTES)
	if len(decrypted) == 0 {
		return nil, fmt.Errorf("no data after crypto sealed box bytes header")
	}
	if C.crypto_box_seal_open((*C.uchar)(&decrypted[0]), (*C.uchar)(&data[0]), C.ulonglong(len(data)), (*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0])) != 0 {
		return nil, fmt.Errorf("failed to open crypto sealed box")
	}
	return decrypted, nil
}

func cryptoSign(data []byte, privateKey []byte) ([]byte, error) {
	if len(privateKey) != C.crypto_sign_SECRETKEYBYTES {
		return nil, fmt.Errorf("expected %d byte private key, got %d bytes", C.crypto_sign_SECRETKEYBYTES, len(privateKey))
	}
	var signature [C.crypto_sign_BYTES]byte
	if C.crypto_sign_detached((*C.uchar)(&signature[0]), nil, (*C.uchar)(&data[0]), (C.ulonglong)(len(data)), (*C.uchar)(&privateKey[0])) != 0 {
		return nil, fmt.Errorf("failed to sign data with private key")
	}
	return signature[:], nil
}

func cryptoSignVerify(data []byte, publicKey []byte, signature []byte) bool {
	if len(publicKey) != C.crypto_sign_PUBLICKEYBYTES || len(signature) != C.crypto_sign_BYTES {
		return false
	}
	return C.crypto_sign_verify_detached((*C.uchar)(&signature[0]), (*C.uchar)(&data[0]), C.ulonglong(len(data)), (*C.uchar)(&publicKey[0])) == 0
}

type fragment struct {
	data     []byte
	received bool
}

type fragmentPacket struct {
	timestamp     time.Time
	packetType    uint8
	fragments     []fragment
	fragmentTotal uint16
	status        uint16
}

type fragmentBuffer struct {
	packets map[uint64]fragmentPacket
	mutex   sync.RWMutex
}

func (buffer *fragmentBuffer) Init() {
	buffer.packets = make(map[uint64]fragmentPacket)
}

func (buffer *fragmentBuffer) Add(timestamp time.Time, packetType uint8, id uint64, fragmentIndex uint8, fragmentTotal uint8, status uint16, data []byte) []byte {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	if int(fragmentTotal) > FragmentMax {
		return nil // invalid fragment count
	}
	if fragmentIndex >= fragmentTotal {
		return nil // invalid fragment index
	}

	// look for existing entries in ring buffer
	packet, existing := buffer.packets[id]
	if existing {
		if packet.packetType != packetType {
			return nil // packet type mismatch
		}
		if packet.fragmentTotal != uint16(fragmentTotal) {
			return nil // total fragment count mismatch
		}
		if packet.status != status {
			return nil // status mismatch
		}
	} else {
		packet.fragmentTotal = uint16(fragmentTotal)
		packet.packetType = packetType
		packet.fragments = make([]fragment, FragmentMax)
		packet.status = status
	}
	packet.timestamp = timestamp
	fragment := &packet.fragments[fragmentIndex]
	fragment.data = data
	fragment.received = true
	buffer.packets[id] = packet

	// check all fragments

	completeBytes := 0

	for i := 0; i < int(fragmentTotal); i++ {
		fragment := &packet.fragments[i]
		if packet.fragments[i].received {
			completeBytes += len(fragment.data)
		} else {
			return nil // still missing a fragment
		}
	}

	// we have all fragments
	delete(buffer.packets, id)

	bytes := 0
	completePacket := make([]byte, completeBytes)
	for i := 0; i < int(fragmentTotal); i++ {
		fragment := &packet.fragments[i]
		copy(completePacket[bytes:], fragment.data)
		bytes += len(fragment.data)
	}

	return completePacket
}

func (buffer *fragmentBuffer) Cleanup(t time.Time) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	timestampThreshold := t.Add(-FragmentTimeout)
	for id, packet := range buffer.packets {
		if packet.timestamp.Before(timestampThreshold) {
			delete(buffer.packets, id)
		}
	}
}

type UDPListenerMaster struct {
	masterTokenSignPrivateKey []byte
	sealPublicKey             []byte
	sealPrivateKey            []byte
	fragmentBuffer            fragmentBuffer
	lastCleanup               time.Time
}

func UDPListenerMasterCreate(masterTokenSignPrivateKey []byte, sealPublicKey []byte, sealPrivateKey []byte) *UDPListenerMaster {
	listener := &UDPListenerMaster{}
	listener.masterTokenSignPrivateKey = masterTokenSignPrivateKey
	listener.sealPublicKey = sealPublicKey
	listener.sealPrivateKey = sealPrivateKey
	listener.fragmentBuffer.Init()
	return listener
}

func (listener *UDPListenerMaster) Pump(t time.Time) {
	if t.After(listener.lastCleanup.Add(2 * time.Second)) {
		listener.lastCleanup = t
		listener.fragmentBuffer.Cleanup(t)
	}
}

func (listener *UDPListenerMaster) Handle(t time.Time, packetData []byte) (*UDPPacketToMaster, error) {
	// 1 byte packet type
	// <encrypted>
	//     <master token>
	//         19 byte IP address
	//         8 byte timestamp
	//         64 byte MAC
	//     </master token>
	//     8 byte GUID
	//     1 byte fragment index
	//     1 byte fragment count
	//     <zipped>
	//         JSON string
	//     </zipped>
	// </encrypted>

	if len(packetData) < 1 {
		return nil, fmt.Errorf("1 byte packet too small")
	}

	packetType := packetData[0]

	const minSize = 1 + MasterTokenBytes + 8 + 2
	const maxSize = minSize + FragmentSize + C.crypto_box_SEALBYTES

	if len(packetData) < minSize {
		return nil, fmt.Errorf("%d byte packet too small, expected at least %d bytes", len(packetData), minSize)
	}

	if len(packetData) > maxSize {
		return nil, fmt.Errorf("%d byte packet too big, expected no more than %d bytes", len(packetData), maxSize)
	}

	decrypted, err := cryptoBoxSealOpen(packetData[1:], listener.sealPublicKey, listener.sealPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt client UDP packet: %v", err)
	}

	if packetType != NEXT_PACKET_TYPE_RELAY_INIT_REQUEST {
		masterToken := masterTokenRead(decrypted[0:MasterTokenBytes], listener.masterTokenSignPrivateKey)
		if masterToken == nil {
			return nil, fmt.Errorf("failed to read master token")
		}
	}

	payloadOffset := MasterTokenBytes

	id := binary.LittleEndian.Uint64(decrypted[payloadOffset:])

	fragmentIndex := decrypted[payloadOffset+8]
	fragmentTotal := decrypted[payloadOffset+9]
	payloadOffset += 8 + 2

	completePacket := listener.fragmentBuffer.Add(t, packetType, id, fragmentIndex, fragmentTotal, 0, decrypted[payloadOffset:])
	if completePacket != nil {
		decompressed, err := decompress(completePacket)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress client UDP packet: %v", err)
		}

		return &UDPPacketToMaster{
			Type: packetType,
			ID:   id,
			Data: decompressed,
		}, nil
	}

	return nil, nil
}

type UDPHandlerMaster func(*UDPPacketToMaster, *net.UDPAddr, *net.UDPConn) error

func (listener *UDPListenerMaster) Listen(
	packetsReceivedCount *int64,
	handler UDPHandlerMaster,
	address string,
) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return fmt.Errorf("failed to bind resolve bind address '%s': %v", address, err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind to UDP port: %v", err)
	}
	defer conn.Close()

	buffer := make([]byte, 4096)
	
	for {
		packetBytes, from, _ := conn.ReadFromUDP(buffer)
		if packetBytes > 0 {
			atomic.AddInt64(packetsReceivedCount, 1)
			func(buffer []byte, from *net.UDPAddr, conn *net.UDPConn) {
				packet, err := listener.Handle(time.Now(), buffer)
				if err != nil {
					fmt.Printf("error: UDP_Listen: %v", err)
					return
				}
				if packet != nil {
					err = handler(packet, from, conn)
					if err != nil {
						fmt.Printf("error: UDP_Listen: %v", err)
						return
					}
				}
			}(buffer[:packetBytes], from, conn)
		}
		listener.Pump(time.Now())
	}

	return nil
}

type UDPListenerClient struct {
	masterSignPublicKey []byte
	fragmentBuffer      fragmentBuffer
	lastCleanup         time.Time
}

func UDPListenerClientCreate(masterSignPublicKey []byte) *UDPListenerClient {
	listener := &UDPListenerClient{}
	listener.masterSignPublicKey = masterSignPublicKey
	listener.fragmentBuffer.Init()
	return listener
}

func (listener *UDPListenerClient) Pump(t time.Time) {
	if t.After(listener.lastCleanup.Add(2 * time.Second)) {
		listener.lastCleanup = t
		listener.fragmentBuffer.Cleanup(t)
	}
}

func (listener *UDPListenerClient) Handle(t time.Time, packetData []byte) (*UDPPacketToClient, error) {
	// 1 byte packet type
	// 64 byte signature
	// <signed>
	//     8 byte GUID
	//     1 byte fragment index
	//     1 byte fragment count
	//     2 byte status code
	//     <zipped>
	//         JSON string
	//     </zipped>
	// </signed>

	if len(packetData) < 1 {
		return nil, fmt.Errorf("1 byte packet too small")
	}

	packetType := packetData[0]

	const unsignedSize = 1 + C.crypto_sign_BYTES
	const headerSize = unsignedSize + 8 + 1 + 1 + 2
	const minSize = headerSize
	const maxSize = minSize + FragmentSize

	if len(packetData) < minSize {
		return nil, fmt.Errorf("%d byte packet too small, expected at least %d bytes", len(packetData), minSize)
	}

	if len(packetData) > maxSize {
		return nil, fmt.Errorf("%d byte packet too big, expected no more than %d bytes", len(packetData), maxSize)
	}

	signed := packetData[unsignedSize:]
	signature := packetData[1 : 1+C.crypto_sign_BYTES]
	payload := packetData[headerSize:]

	if !cryptoSignVerify(signed, listener.masterSignPublicKey, signature) {
		return nil, fmt.Errorf("failed to verify master packet signature")
	}

	id := binary.LittleEndian.Uint64(signed[0:8])
	fragmentIndex := signed[8]
	fragmentTotal := signed[9]
	status := binary.LittleEndian.Uint16(signed[10:12])

	completePacket := listener.fragmentBuffer.Add(t, packetType, id, fragmentIndex, fragmentTotal, status, payload)
	if completePacket != nil {
		decompressed, err := decompress(completePacket)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress client UDP packet: %v", err)
		}

		return &UDPPacketToClient{
			Type:   packetType,
			ID:     id,
			Status: status,
			Data:   decompressed,
		}, nil
	}

	return nil, nil
}

type UDPPacketToClientBuilder struct {
	signPrivateKey []byte
}

func UDPPacketToClientBuilderCreate(signPrivateKey []byte) *UDPPacketToClientBuilder {
	return &UDPPacketToClientBuilder{
		signPrivateKey: signPrivateKey,
	}
}

func (builder *UDPPacketToClientBuilder) buildFragment(packetType uint8, id uint64, fragmentIndex int, fragmentTotal int, status uint16, packetData []byte) (error, []byte) {
	// 1 byte packet type
	// 64 byte signature
	// <signed>
	//     8 byte GUID
	//     1 byte fragment index
	//     1 byte fragment count
	//     2 byte status code
	//     <zipped>
	//         JSON string
	//     </zipped>
	// </signed>

	var signed bytes.Buffer

	var scratch [8]byte
	binary.LittleEndian.PutUint64(scratch[:], id)
	signed.Write(scratch[:])

	signed.WriteByte(uint8(fragmentIndex))
	signed.WriteByte(uint8(fragmentTotal))

	binary.LittleEndian.PutUint16(scratch[:], status)
	signed.Write(scratch[:2])

	signed.Write(packetData)

	signedBytes := signed.Bytes()

	signature, err := cryptoSign(signedBytes, builder.signPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign master UDP packet: %v", err), nil
	}

	var packet bytes.Buffer
	packet.WriteByte(packetType)
	packet.Write(signature)
	packet.Write(signedBytes)

	return nil, packet.Bytes()
}

func (builder *UDPPacketToClientBuilder) Build(packet *UDPPacketToClient) ([][]byte, error) {

	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	writer.Write(packet.Data)
	writer.Close()

	data := compressed.Bytes()

	fragmentTotal := len(data) / FragmentSize
	if len(data)%FragmentSize != 0 {
		fragmentTotal += 1
	}
	if fragmentTotal > FragmentMax {
		return nil, fmt.Errorf("%d byte packet can't be sent; too big even for %d fragments", len(data), FragmentMax)
	}

	packets := make([][]byte, 0)

	for i := 0; i < fragmentTotal; i++ {
		var fragmentBytes int
		if i == fragmentTotal-1 {
			// last fragment
			fragmentBytes = len(data) - ((fragmentTotal - 1) * FragmentSize)
		} else {
			fragmentBytes = FragmentSize
		}
		fragmentStart := i * FragmentSize
		err, fragment := builder.buildFragment(
			packet.Type,
			packet.ID,
			i,
			fragmentTotal,
			packet.Status,
			data[fragmentStart:fragmentStart+fragmentBytes],
		)
		if err != nil {
			return nil, err
		}
		packets = append(packets, fragment)
	}

	return packets, nil
}

type UDPPacketToMasterBuilder struct {
	sealPublicKey     []byte
	masterTokenBuffer []byte
}

func UDPPacketToMasterBuilderCreate(sealPublicKey []byte, masterToken *masterToken) *UDPPacketToMasterBuilder {
	return &UDPPacketToMasterBuilder{
		sealPublicKey:     sealPublicKey,
		masterTokenBuffer: masterToken.Write(),
	}
}

func (builder *UDPPacketToMasterBuilder) buildFragment(packetType uint8, id uint64, fragmentIndex int, fragmentTotal int, packetData []byte) (error, []byte) {

	// 1 byte packet type
	// <encrypted>
	//     <master token>
	//         19 byte IP address
	//         8 byte timestamp
	//         32 byte MAC
	//     </master token>
	//     8 byte GUID
	//     1 byte fragment index
	//     1 byte fragment count
	//     <zipped>
	//         JSON string
	//     </zipped>
	// </encrypted>
	// 64 byte MAC (handled automatically by sodium)

	var encrypted bytes.Buffer

	encrypted.Write(builder.masterTokenBuffer)

	var scratch [8]byte
	binary.LittleEndian.PutUint64(scratch[:], id)
	encrypted.Write(scratch[:])

	encrypted.WriteByte(uint8(fragmentIndex))
	encrypted.WriteByte(uint8(fragmentTotal))
	encrypted.Write(packetData)

	encryptedBytes, err := cryptoBoxSeal(encrypted.Bytes(), builder.sealPublicKey)
	if err != nil {
		return fmt.Errorf("failed to sign master UDP packet: %v", err), nil
	}

	var packet bytes.Buffer
	packet.WriteByte(packetType)
	packet.Write(encryptedBytes)

	return nil, packet.Bytes()
}

func (builder *UDPPacketToMasterBuilder) Build(packet *UDPPacketToMaster) ([][]byte, error) {

	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	writer.Write(packet.Data)
	writer.Close()

	data := compressed.Bytes()

	fragmentTotal := len(data) / FragmentSize
	if len(data)%FragmentSize != 0 {
		fragmentTotal += 1
	}
	if fragmentTotal > FragmentMax {
		return nil, fmt.Errorf("%d byte packet can't be sent; too big even for %d fragments", len(data), FragmentMax)
	}

	packets := make([][]byte, 0)

	for i := 0; i < fragmentTotal; i++ {
		var fragmentBytes int
		if i == fragmentTotal-1 {
			// last fragment
			fragmentBytes = len(data) - ((fragmentTotal - 1) * FragmentSize)
		} else {
			fragmentBytes = FragmentSize
		}
		fragmentStart := i * FragmentSize
		err, packet := builder.buildFragment(
			packet.Type,
			packet.ID,
			i,
			fragmentTotal,
			data[fragmentStart:fragmentStart+fragmentBytes],
		)
		if err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}

	return packets, nil
}
