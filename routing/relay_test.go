package routing_test

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestRelay(t *testing.T) {
	t.Run("NewRelay()", func(t *testing.T) {
		relay := routing.Relay{
			PublicKey: make([]byte, crypto.KeySize),
		}

		assert.Equal(t, relay, routing.NewRelay())
	})

	t.Run("UnmarshalBinary()", func(t *testing.T) {
		const (
			sellerID   = "seller ID"
			sellerName = "seller name"
			relayname  = "relay name"
			addr       = "127.0.0.1:40000"
			dcname     = "datacenter name"
		)

		var subject routing.Relay
		size := 0
		t.Run("missing ID", func(t *testing.T) {
			buff := make([]byte, size)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay ID")
		})

		t.Run("missing name", func(t *testing.T) {
			size += 8
			buff := make([]byte, size)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay name")
		})

		t.Run("missing address", func(t *testing.T) {
			size += 4 + len(relayname)
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			copy(buff[12:], relayname)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay address")
		})

		t.Run("missing seller id", func(t *testing.T) {
			size += 4 + len(addr)
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay seller ID")
		})

		t.Run("missing seller name", func(t *testing.T) {
			size += 4 + len(sellerID)
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerID)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerID)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay seller name")
		})

		t.Run("missing seller ingress price", func(t *testing.T) {
			t.Skip()
			size += 4 + len(sellerName)
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerID)))
			binary.LittleEndian.PutUint32(buff[20+len(relayname)+len(addr)+len(sellerID):], uint32(len(sellerName)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerID)
			copy(buff[32+len(relayname)+len(addr)+len(sellerID):], sellerName)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay seller ingress price")
		})

		t.Run("missing seller egress price", func(t *testing.T) {
			t.Skip()
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay seller egress price")
		})

		t.Run("missing datacenter ID", func(t *testing.T) {
			t.Skip()
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay datacenter id")
		})

		t.Run("missing datacenter name", func(t *testing.T) {
			t.Skip()
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay datacenter name")
		})

		t.Run("missing public key", func(t *testing.T) {
			t.Skip()
			size += 4 + len(dcname)

			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			binary.LittleEndian.PutUint32(buff[44+len(relayname)+len(addr)+len(sellerName):], uint32(len(dcname)))

			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			copy(buff[48+len(relayname)+len(addr)+len(sellerName):], dcname)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay public key")
		})

		t.Run("missing latitude", func(t *testing.T) {
			t.Skip()
			size += crypto.KeySize
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			binary.LittleEndian.PutUint32(buff[44+len(relayname)+len(addr)+len(sellerName):], uint32(len(dcname)))

			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			copy(buff[48+len(relayname)+len(addr)+len(sellerName):], dcname)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay latitude")
		})

		t.Run("missing longitude", func(t *testing.T) {
			t.Skip()
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			binary.LittleEndian.PutUint32(buff[44+len(relayname)+len(addr)+len(sellerName):], uint32(len(dcname)))

			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			copy(buff[48+len(relayname)+len(addr)+len(sellerName):], dcname)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay longitude")
		})

		t.Run("missing last update time", func(t *testing.T) {
			t.Skip()
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			binary.LittleEndian.PutUint32(buff[44+len(relayname)+len(addr)+len(sellerName):], uint32(len(dcname)))

			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			copy(buff[48+len(relayname)+len(addr)+len(sellerName):], dcname)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "failed to unmarshal relay last update time")
		})

		t.Run("invalid address", func(t *testing.T) {
			t.Skip()
			addr := "---not-valid---"
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			binary.LittleEndian.PutUint32(buff[44+len(relayname)+len(addr)+len(sellerName):], uint32(len(dcname)))

			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			copy(buff[48+len(relayname)+len(addr)+len(sellerName):], dcname)
			assert.EqualError(t, subject.UnmarshalBinary(buff), "invalid relay address")
		})

		t.Run("valid", func(t *testing.T) {
			t.Skip()
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[16+len(relayname)+len(addr):], uint32(len(sellerName)))
			binary.LittleEndian.PutUint32(buff[44+len(relayname)+len(addr)+len(sellerName):], uint32(len(dcname)))

			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[24+len(relayname)+len(addr):], sellerName)
			copy(buff[48+len(relayname)+len(addr)+len(sellerName):], dcname)
			assert.NoError(t, subject.UnmarshalBinary(buff))
		})
	})

	t.Run("MarshalBinary()", func(t *testing.T) {
		udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		pk := make([]byte, crypto.KeySize)
		rand.Read(pk)

		expected := routing.Relay{
			ID:        123,
			Name:      "relay name",
			Addr:      *udp,
			PublicKey: pk,
			Seller: routing.Seller{
				ID:                "12345678",
				Name:              "seller name",
				IngressPriceCents: 456,
				EgressPriceCents:  789,
			},
			Datacenter: routing.Datacenter{
				ID:   321,
				Name: "datacenter name",
			},
			Latitude:       123.456,
			Longitude:      654.321,
			LastUpdateTime: 999,
		}

		data, _ := expected.MarshalBinary()

		var actual routing.Relay
		actual.UnmarshalBinary(data)

		assert.Equal(t, expected, actual)
	})

	t.Run("Key()", func(t *testing.T) {
		relay := routing.Relay{
			ID: 123,
		}

		assert.Equal(t, "RELAY-123", relay.Key())
	})
}
