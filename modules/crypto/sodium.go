package crypto

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

func sodiumSignPacket(packetData []byte, privateKey []byte) []byte {
	signedPacketData := make([]byte, len(packetData)+C.crypto_sign_BYTES)
	for i := 0; i < len(packetData); i++ {
		signedPacketData[i] = packetData[i]
	}
	messageLength := len(packetData) - PacketHashSize - 1
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&signedPacketData[PacketHashSize+1]), C.ulonglong(messageLength))
	C.crypto_sign_final_create(&state, (*C.uchar)(&signedPacketData[len(packetData)]), nil, (*C.uchar)(&privateKey[0]))
	return signedPacketData
}

// This function assumes that the packetData already has the first 9 bytes (packet type and packet hash) removed,
// since the server backend will remove these bytes before passing it to the handlers for signature checking
func sodiumVerifyPacket(packetData []byte, publicKey []byte) bool {
	if len(packetData) < C.crypto_sign_BYTES || len(publicKey) != C.crypto_sign_PUBLICKEYBYTES {
		return false
	}

	messageLength := len(packetData) - C.crypto_sign_BYTES

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(messageLength))
	return C.crypto_sign_final_verify(&state, (*C.uchar)(&packetData[messageLength]), (*C.uchar)(&publicKey[0])) == 0
}

func sodiumHashPacket(packetData []byte, hashKey []byte) {
	messageLength := len(packetData) - PacketHashSize - 1
	if messageLength > 32 {
		messageLength = 32
	}
	C.crypto_generichash(
		(*C.uchar)(&packetData[1]),
		C.ulong(PacketHashSize),
		(*C.uchar)(&packetData[PacketHashSize+1]),
		C.ulonglong(messageLength),
		(*C.uchar)(&hashKey[0]),
		C.ulong(C.crypto_generichash_KEYBYTES),
	)
}

func sodiumIsNetworkNextPacket(packetData []byte, hashKey []byte) bool {
	packetBytes := len(packetData)
	if packetBytes <= 1+PacketHashSize {
		return false
	}
	messageLength := packetBytes - 1 - PacketHashSize
	if messageLength > 32 {
		messageLength = 32
	}
	hash := make([]byte, PacketHashSize)
	C.crypto_generichash(
		(*C.uchar)(&hash[0]),
		C.ulong(PacketHashSize),
		(*C.uchar)(&packetData[1+PacketHashSize]),
		C.ulonglong(messageLength),
		(*C.uchar)(&hashKey[0]),
		C.ulong(C.crypto_generichash_KEYBYTES),
	)
	for i := 0; i < PacketHashSize; i++ {
		if hash[i] != packetData[i+1] {
			return false
		}
	}
	return true
}

func sodiumSignPacketSDK5(packetData []byte, serializeBytes int, privateKey []byte) []byte {
	signedPacketData := packetData[:1+15+serializeBytes+int(C.crypto_sign_BYTES)+2]

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&signedPacketData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&signedPacketData[16]), C.ulonglong(serializeBytes))
	C.crypto_sign_final_create(&state, (*C.uchar)(&signedPacketData[16+serializeBytes]), nil, (*C.uchar)(&privateKey[0]))

	return signedPacketData
}
