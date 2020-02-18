package routing_test

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
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
			relayname = "relay name"
			addr      = "127.0.0.1:40000"
			dcname    = "datacenter name"
		)

		var subject routing.Relay
		size := 0
		t.Run("missing ID", func(t *testing.T) {
			buff := make([]byte, size)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("missing name", func(t *testing.T) {
			size += 8
			buff := make([]byte, size)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("missing address", func(t *testing.T) {
			size += 4 + len(relayname)
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			copy(buff[12:], relayname)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("missing datacenter ID", func(t *testing.T) {
			size += 4 + len(addr)
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("missing datacenter name", func(t *testing.T) {
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("missing public key", func(t *testing.T) {
			size += 4 + len(dcname)
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[24+len(relayname)+len(addr):], uint32(len(dcname)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[28+len(relayname)+len(addr):], dcname)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("missing latitude", func(t *testing.T) {
			size += crypto.KeySize
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[24+len(relayname)+len(addr):], uint32(len(dcname)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[28+len(relayname)+len(addr):], dcname)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("missing longitude", func(t *testing.T) {
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[24+len(relayname)+len(addr):], uint32(len(dcname)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[28+len(relayname)+len(addr):], dcname)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("missing last update time", func(t *testing.T) {
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[24+len(relayname)+len(addr):], uint32(len(dcname)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[28+len(relayname)+len(addr):], dcname)
			assert.Equal(t, errors.New("Invalid Relay"), subject.UnmarshalBinary(buff))
		})

		t.Run("invalid address", func(t *testing.T) {
			addr := "---not-valid---"
			size += 8
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[24+len(relayname)+len(addr):], uint32(len(dcname)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[28+len(relayname)+len(addr):], dcname)
			assert.Equal(t, errors.New("Invalid relay address"), subject.UnmarshalBinary(buff))
		})

		t.Run("valid", func(t *testing.T) {
			buff := make([]byte, size)
			binary.LittleEndian.PutUint32(buff[8:], uint32(len(relayname)))
			binary.LittleEndian.PutUint32(buff[12+len(relayname):], uint32(len(addr)))
			binary.LittleEndian.PutUint32(buff[24+len(relayname)+len(addr):], uint32(len(dcname)))
			copy(buff[12:], relayname)
			copy(buff[16+len(relayname):], addr)
			copy(buff[28+len(relayname)+len(addr):], dcname)
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
