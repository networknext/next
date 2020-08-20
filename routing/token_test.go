package routing_test

import (
	"crypto/rand"
	"net"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

func TestEncryptNextRouteToken(t *testing.T) {
	t.Parallel()

	t.Run("failures", func(t *testing.T) {
		nodepublickey, privatekey, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		token := routing.NextRouteToken{
			Server: routing.Server{
				Addr:      net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Relays: []routing.RelayToken{
				{
					ID:        1,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.1"), Port: 13},
					PublicKey: nodepublickey[:],
				},
			},
		}
		enc, _, err := token.Encrypt(privatekey[:])
		assert.Nil(t, enc)
		assert.EqualError(t, err, "client public key cannot be nil")

		token = routing.NextRouteToken{
			Client: routing.Client{
				Addr:      net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Relays: []routing.RelayToken{
				{
					ID:        1,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.1"), Port: 13},
					PublicKey: nodepublickey[:],
				},
			},
		}
		enc, _, err = token.Encrypt(privatekey[:])
		assert.Nil(t, enc)
		assert.EqualError(t, err, "server public key cannot be nil")

		token = routing.NextRouteToken{
			Client: routing.Client{
				Addr:      net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Server: routing.Server{
				Addr:      net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Relays: []routing.RelayToken{
				{
					ID:   1,
					Addr: net.UDPAddr{IP: net.ParseIP("192.168.0.1"), Port: 13},
				},
				{
					ID:        2,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.2"), Port: 13},
					PublicKey: nodepublickey[:],
				},
			},
		}
		enc, _, err = token.Encrypt(privatekey[:])
		assert.Nil(t, enc)
		assert.EqualError(t, err, "relay public key at index 0 cannot be nil")

		token = routing.NextRouteToken{
			Client: routing.Client{
				Addr:      net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Server: routing.Server{
				Addr:      net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
		}
		enc, _, err = token.Encrypt(privatekey[:])
		assert.Nil(t, enc)
		assert.EqualError(t, err, "at least 1 relay is required")
	})

	t.Run("success", func(t *testing.T) {
		nodepublickey, privatekey, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		token := routing.NextRouteToken{
			Expires:        1,
			SessionID:      2,
			SessionVersion: 3,
			SessionFlags:   4,
			KbpsUp:         5,
			KbpsDown:       6,

			Client: routing.Client{
				Addr:      net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Server: routing.Server{
				Addr:      net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Relays: []routing.RelayToken{
				{
					ID:        1,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.1"), Port: 13},
					PublicKey: nodepublickey[:],
				},
				{
					ID:        2,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.2"), Port: 13},
					PublicKey: nodepublickey[:],
				},
				{
					ID:        3,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.3"), Port: 13},
					PublicKey: nodepublickey[:],
				},
			},
		}

		enctoken, _, err := token.Encrypt(privatekey[:])
		assert.NoError(t, err)
		assert.Equal(t, 585, len(enctoken))
	})
}

func TestEncryptContinueRouteDecision(t *testing.T) {
	nodepublickey, _, err := box.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		token := routing.ContinueRouteToken{
			Expires:        1,
			SessionID:      2,
			SessionVersion: 3,
			SessionFlags:   4,

			Client: routing.Client{
				Addr:      net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Server: routing.Server{
				Addr:      net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 13},
				PublicKey: nodepublickey[:],
			},
			Relays: []routing.RelayToken{
				{
					ID:        1,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.1"), Port: 13},
					PublicKey: nodepublickey[:],
				},
				{
					ID:        2,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.2"), Port: 13},
					PublicKey: nodepublickey[:],
				},
				{
					ID:        3,
					Addr:      net.UDPAddr{IP: net.ParseIP("192.168.0.3"), Port: 13},
					PublicKey: nodepublickey[:],
				},
			},
		}

		enctoken, _, err := token.Encrypt(crypto.RouterPrivateKey)
		assert.NoError(t, err)
		assert.Equal(t, 290, len(enctoken))
	})
}

func BenchmarkEncryptNextRouteToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		token := routing.NextRouteToken{
			Expires:        1,
			SessionID:      2,
			SessionVersion: 3,
			SessionFlags:   4,
			KbpsUp:         5,
			KbpsDown:       6,

			Client: routing.Client{
				Addr:      net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 13},
				PublicKey: make([]byte, crypto.KeySize),
			},
			Server: routing.Server{
				Addr:      net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 13},
				PublicKey: make([]byte, crypto.KeySize),
			},
			Relays: []routing.RelayToken{
				{
					ID:   1,
					Addr: net.UDPAddr{IP: net.ParseIP("192.168.0.1"), Port: 13},
				},
				{
					ID:   2,
					Addr: net.UDPAddr{IP: net.ParseIP("192.168.0.2"), Port: 13},
				},
				{
					ID:   3,
					Addr: net.UDPAddr{IP: net.ParseIP("192.168.0.3"), Port: 13},
				},
			},
		}

		token.Encrypt(crypto.RouterPrivateKey)
	}
}
