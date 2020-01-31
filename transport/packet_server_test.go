package transport_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"net"
	"testing"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestServerUpdatePacket(t *testing.T) {
	t.Run("crypto/ed25519", func(t *testing.T) {
		customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
		customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

		// Create a ServerUpdatePacket like the game server does
		outgoing := transport.ServerUpdatePacket{
			Sequence:             1,
			VersionMajor:         1,
			VersionMinor:         2,
			VersionPatch:         3,
			CustomerId:           4,
			DatacenterId:         5,
			NumSessionsPending:   6,
			NumSessionsUpgraded:  7,
			ServerAddress:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerPrivateAddress: net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ServerRoutePublicKey: make([]byte, crypto.KeySize),
		}

		// Sign the packet and set it to the signature
		// using the customer's private key that is on
		// their game server
		outgoing.Signature = ed25519.Sign(customerPrivateKey[8:], outgoing.GetSignData())

		// Marshal the whole packet to binary to send it over the network
		data, err := outgoing.MarshalBinary()
		assert.NoError(t, err)

		// Unmarshal the data from binary like the server backend receives it
		var incoming transport.ServerUpdatePacket
		err = incoming.UnmarshalBinary(data)
		assert.NoError(t, err)

		// Verify the incoming packet's signed data with the signature
		// with the customer's public key we would get from configstore
		ed25519.Verify(customerPublicKey[8:], incoming.GetSignData(), incoming.Signature)
	})

	t.Run("libsodium", func(t *testing.T) {
		customerPublicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")
		customerPrivateKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn")

		// Create a ServerUpdatePacket like the game server does
		outgoing := transport.ServerUpdatePacket{
			Sequence:             1,
			VersionMajor:         1,
			VersionMinor:         2,
			VersionPatch:         3,
			CustomerId:           4,
			DatacenterId:         5,
			NumSessionsPending:   6,
			NumSessionsUpgraded:  7,
			ServerAddress:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2323},
			ServerPrivateAddress: net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 2323},
			ServerRoutePublicKey: make([]byte, crypto.KeySize),
		}

		// Sign the packet and set it to the signature
		// using the customer's private key that is on
		// their game server
		outgoing.Signature = core.CryptoSignCreate(outgoing.GetSignData(), customerPrivateKey[8:])

		// Marshal the whole packet to binary to send it over the network
		data, err := outgoing.MarshalBinary()
		assert.NoError(t, err)

		// Unmarshal the data from binary like the server backend receives it
		var incoming transport.ServerUpdatePacket
		err = incoming.UnmarshalBinary(data)
		assert.NoError(t, err)

		// Verify the incoming packet's signed data with the signature
		// with the customer's public key we would get from configstore
		core.CryptoSignVerify(incoming.GetSignData(), incoming.Signature, customerPublicKey[8:])
	})
}
