package crypto_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/networknext/backend/modules/crypto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

func TestHashCheck(t *testing.T) {
	msg := []byte("just some data to hash")

	expected := []byte{
		// Hash
		0x58, 0x95, 0x1d, 0x93, 0xd1, 0x31, 0x6e, 0x61,

		// Message "just some data to hash"
		0x6a, 0x75, 0x73, 0x74, 0x20, 0x73, 0x6f, 0x6d, 0x65, 0x20, 0x64, 0x61, 0x74, 0x61, 0x20, 0x74, 0x6f, 0x20, 0x68, 0x61, 0x73, 0x68,
	}
	packetHash := crypto.Hash(crypto.PacketHashKey, msg)
	assert.Equal(t, crypto.PacketHashSize+len(msg), len(packetHash))
	assert.Equal(t, expected, packetHash)

	assert.True(t, crypto.Check(crypto.PacketHashKey, packetHash))
	assert.False(t, crypto.Check(crypto.PacketHashKey, []byte("short")))

	expected[0] = 0x13 // Change any part of the hash so it doesn't represet the data
	assert.False(t, crypto.Check(crypto.PacketHashKey, expected))
}

func TestSignVerify(t *testing.T) {
	// Note: when using these we need to offset the keys by 8 bytes since the first 8 bytes is the CustomerID
	publicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
	privateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

	msg := []byte("just some data to sign")

	// Generate some valid, but wrong keys
	wrongPublicKey, wrongPrivateKey, _ := crypto.GenerateCustomerKeyPair()
	wrongSig := crypto.Sign(wrongPrivateKey, msg)

	// Verify Signing fails when key length is wrong
	sig := crypto.Sign(privateKey[7:], msg)
	assert.Nil(t, sig)

	// Verify Signing successful when a valid key provided
	sig = crypto.Sign(privateKey[8:], msg)
	assert.NotNil(t, sig)
	assert.Len(t, sig, ed25519.SignatureSize)

	// Verification should fail when wrong key provided
	assert.False(t, crypto.Verify(wrongPublicKey, msg, sig))

	// Verification should fail when wrong signature provided
	assert.False(t, crypto.Verify(publicKey, msg, wrongSig))

	// If the right public key + signature are provided, should succeed
	sig = crypto.Sign(privateKey[8:], msg)
	assert.True(t, crypto.Verify(publicKey[8:], msg, sig))
}

func TestOpenSeal(t *testing.T) {
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
	testString := "testString"
	expectedHash := uint64(886614244633029176)
	assert.Equal(t, expectedHash, crypto.HashID(testString))
}
