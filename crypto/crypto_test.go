package crypto_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

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
