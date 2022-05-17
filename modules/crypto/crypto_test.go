package crypto_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

func TestSignVerifyPacket(t *testing.T) {
	t.Parallel()

	// Note: when using these we need to offset the keys by 8 bytes since the first 8 bytes is the CustomerID
	publicKey, err := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
	assert.NoError(t, err)

	privateKey, err := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")
	assert.NoError(t, err)

	// Generate a valid, but wrong public key
	wrongPublicKey, _, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	msg := []byte("just some signed data")

	// Add in the packet type byte and hash bytes to the front of the message
	msgHeader := append([]byte{0}, make([]byte, crypto.PacketHashSize)...)
	msg = append(msgHeader, msg...)

	// Verify Signing successful when a valid key provided
	signedMsg := crypto.SignPacket(privateKey[8:], msg)
	assert.NotNil(t, signedMsg)
	assert.Len(t, signedMsg, len(msg)+crypto.PacketSignatureSize)

	// Now remove the message header from the signed message in order to properly verify it
	signedMsg = signedMsg[1+crypto.PacketHashSize:]

	// Verification should fail when wrong key provided
	assert.False(t, crypto.VerifyPacket(wrongPublicKey[8:], signedMsg))

	// Verification should fail when the message isn't signed properly
	assert.False(t, crypto.VerifyPacket(publicKey[8:], msg))

	// If the right public key + signature are provided, should succeed
	assert.True(t, crypto.VerifyPacket(publicKey[8:], signedMsg))
}

func TestSignVerifyPacketSDK5(t *testing.T) {
	t.Parallel()

	// Note: when using these we need to offset the keys by 8 bytes since the first 8 bytes is the CustomerID
	publicKey, err := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
	assert.NoError(t, err)

	privateKey, err := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")
	assert.NoError(t, err)

	// Generate a valid, but wrong public key
	wrongPublicKey, _, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	msg := []byte("just some signed data")
	serializedMsgLen := len(msg)

	// Add in the packet type byte and chonkle to the front of the message,
	// and the pittle to the end of the message, along with extra bytes for signing
	msgHeader := append([]byte{0}, make([]byte, 15)...)
	msg = append(msgHeader, msg...)
	msg = append(msg, make([]byte, 2)...)
	msg = append(msg, make([]byte, 1024)...)

	// Verify Signing successful when a valid key provided
	signedMsg := crypto.SignPacketSDK5(privateKey[8:], msg, serializedMsgLen)
	assert.NotNil(t, signedMsg)
	assert.Len(t, signedMsg, 1+15+serializedMsgLen+crypto.PacketSignatureSize+2)

	// Get the address data
	fromAddr := core.ParseAddress("127.0.0.1:32202")
	toAddr := core.ParseAddress("127.0.0.1:40000")

	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(fromAddr, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(toAddr, toAddressBuffer[:])

	// Add the chonkle and pittle
	var emptyMagic [8]byte
	signedMsgLength := len(signedMsg)
	core.GenerateChonkle(signedMsg[1:], emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, signedMsgLength)
	core.GeneratePittle(signedMsg[signedMsgLength-2:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, signedMsgLength)

	// Verification should fail when wrong key provided
	assert.False(t, crypto.VerifyPacketSDK5(wrongPublicKey[8:], signedMsg))

	// If the right public key + signature are provided, should succeed
	assert.True(t, crypto.VerifyPacketSDK5(publicKey[8:], signedMsg))
}

func TestOpenSeal(t *testing.T) {
	t.Parallel()

	// Generate a sender keypair
	senderpub, senderpriv, err := box.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	// Generate a receiver keypair
	receiverpub, receiverpriv, err := box.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	// Generate a random nonce of bytes
	nonce := make([]byte, crypto.NonceSize)
	n, err := rand.Read(nonce)
	assert.NoError(t, err)
	assert.Equal(t, crypto.NonceSize, n)

	// Define the messages to be send
	expected := []byte("this is a test message")

	// Encrypt the message with the senders private key and the receivers public key
	encrypted := crypto.Seal(expected, nonce, receiverpub[:], senderpriv[:])
	assert.Equal(t, len(expected)+box.Overhead, len(encrypted))

	// Decrypt the message with the receivers private key and the senders public key
	decrypted, ok := crypto.Open(encrypted, nonce, senderpub[:], receiverpriv[:])
	assert.True(t, ok)
	assert.Equal(t, expected, decrypted)

}

func TestGenerate(t *testing.T) {
	t.Parallel()

	t.Run("SessionID", func(t *testing.T) {
		id := crypto.GenerateSessionID()
		assert.NotZero(t, id)
	})

	t.Run("CustomerKeyPair", func(t *testing.T) {
		pub, priv, err := crypto.GenerateCustomerKeyPair()
		assert.NoError(t, err)
		assert.Equal(t, ed25519.PublicKeySize+8, len(pub))
		assert.Equal(t, ed25519.PrivateKeySize+8, len(priv))
	})
}

func TestHashID(t *testing.T) {
	t.Parallel()

	testString := "testString"
	expectedHash := uint64(886614244633029176)
	assert.Equal(t, expectedHash, crypto.HashID(testString))
}
