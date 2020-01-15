package crypto_test

import (
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	t.Run("SessionID", func(t *testing.T) {
		id := crypto.GenerateSessionID()
		assert.NotZero(t, id)
	})

	t.Run("RelayKeyPair", func(t *testing.T) {
		pub, priv, err := crypto.GenerateRelayKeyPair()
		assert.NoError(t, err)
		assert.Equal(t, crypto.PublicKeySize, len(pub))
		assert.Equal(t, crypto.PrivateKeySize, len(priv))
	})

	t.Run("CustomerKeyPair", func(t *testing.T) {
		pub, priv, err := crypto.GenerateCustomerKeyPair()
		assert.NoError(t, err)
		assert.Equal(t, crypto.PublicKeySize+8, len(pub))
		assert.Equal(t, crypto.PrivateKeySize+8, len(priv))
	})
}
