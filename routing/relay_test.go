package routing_test

import (
	"crypto/rand"
	"net"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestRelay(t *testing.T) {
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
			ID:      321,
			Name:    "datacenter name",
			Enabled: true,
		},
		Latitude:       123.456,
		Longitude:      654.321,
		LastUpdateTime: 999,
	}

	t.Run("MarshalBinary()", func(t *testing.T) {
		data, _ := expected.MarshalBinary()

		var actual routing.Relay
		actual.UnmarshalBinary(data)

		assert.Equal(t, expected, actual)
	})

	t.Run("UnmarshalBinary()", func(t *testing.T) {
		var actual routing.Relay
		size := 0
		buff := make([]byte, 0, expected.Size())
		index := 0

		t.Run("missing ID", func(t *testing.T) {
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay ID")
		})

		t.Run("missing name", func(t *testing.T) {
			size += 8
			buff = buff[:size]
			encoding.WriteUint64(buff, &index, expected.ID)
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay name")
		})

		t.Run("missing address", func(t *testing.T) {
			size += 4 + len(expected.Name)
			buff = buff[:size]
			encoding.WriteString(buff, &index, expected.Name, uint32(len(expected.Name)))
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay address")
		})

		t.Run("missing public key", func(t *testing.T) {
			addrString := expected.Addr.String()
			size += 4 + len(addrString)
			buff = buff[:size]
			encoding.WriteString(buff, &index, addrString, uint32(len(addrString)))
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay public key")
		})

		t.Run("missing seller id", func(t *testing.T) {
			size += crypto.KeySize
			buff = buff[:size]
			encoding.WriteBytes(buff, &index, expected.PublicKey, crypto.KeySize)
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay seller ID")
		})

		t.Run("missing seller name", func(t *testing.T) {
			sellerID := expected.Seller.ID
			size += 4 + len(sellerID)
			buff = buff[:size]
			encoding.WriteString(buff, &index, sellerID, uint32(len(sellerID)))
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay seller name")
		})

		t.Run("missing seller ingress price", func(t *testing.T) {
			sellerName := expected.Seller.Name
			size += 4 + len(sellerName)
			buff = buff[:size]
			encoding.WriteString(buff, &index, sellerName, uint32(len(sellerName)))
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay seller ingress price")
		})

		t.Run("missing seller egress price", func(t *testing.T) {
			size += 8
			buff = buff[:size]
			encoding.WriteUint64(buff, &index, expected.Seller.IngressPriceCents)
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay seller egress price")
		})

		t.Run("missing datacenter ID", func(t *testing.T) {
			size += 8
			buff = buff[:size]
			encoding.WriteUint64(buff, &index, expected.Seller.EgressPriceCents)
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay datacenter id")
		})

		t.Run("missing datacenter name", func(t *testing.T) {
			size += 8
			buff = buff[:size]
			encoding.WriteUint64(buff, &index, expected.Datacenter.ID)
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay datacenter name")
		})

		t.Run("missing datacenter enabled", func(t *testing.T) {
			datacenterName := expected.Datacenter.Name
			size += 4 + len(datacenterName)
			buff = buff[:size]
			encoding.WriteString(buff, &index, datacenterName, uint32(len(datacenterName)))
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay datacenter enabled")
		})

		t.Run("missing latitude", func(t *testing.T) {
			datacenterEnabled := expected.Datacenter.Enabled
			size += 1
			buff = buff[:size]
			encoding.WriteBool(buff, &index, datacenterEnabled)
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay latitude")
		})

		t.Run("missing longitude", func(t *testing.T) {
			size += 8
			buff = buff[:size]
			encoding.WriteUint64(buff, &index, uint64(expected.Latitude))
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay longitude")
		})

		t.Run("missing last update time", func(t *testing.T) {
			size += 8
			buff = buff[:size]
			encoding.WriteUint64(buff, &index, uint64(expected.Longitude))
			assert.EqualError(t, actual.UnmarshalBinary(buff), "failed to unmarshal relay last update time")
		})

		t.Run("valid", func(t *testing.T) {
			size += 8
			buff = buff[:size]
			encoding.WriteUint64(buff, &index, uint64(expected.LastUpdateTime))
			assert.NoError(t, actual.UnmarshalBinary(buff))
		})
	})

	t.Run("Key()", func(t *testing.T) {
		relay := routing.Relay{
			ID: 123,
		}

		assert.Equal(t, "RELAY-123", relay.Key())
	})
}
