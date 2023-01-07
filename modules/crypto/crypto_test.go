package crypto_test

import (
	"testing"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/crypto"

	"github.com/stretchr/testify/assert"
)

// todo: Test_Box

func Test_Sign(t *testing.T) {

	publicKey, privateKey := crypto.Sign_KeyPair()

	data := make([]byte, 256)
	common.RandomBytes(data)

	signature := crypto.Sign(data, privateKey)

	assert.True(t, crypto.Verify(data, publicKey, signature))
}

// todo: Test_CustomerKeyPair