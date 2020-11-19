package transport_test

import (
	"net"
	"testing"

	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func TestAnonAddr(t *testing.T) {
	t.Parallel()

	t.Run("anon addr", func(t *testing.T) {
		addr4 := net.UDPAddr{IP: net.ParseIP("68.14.255.202"), Port: 1313}
		anonAddr4 := transport.AnonymizeAddr(addr4)
		assert.Equal(t, "68.14.255.0:0", anonAddr4.String())
		assert.Equal(t, "68.14.255.0", anonAddr4.IP.String())
		assert.Equal(t, 0, anonAddr4.Port)
		assert.NotEqual(t, addr4, anonAddr4)

		addr6 := net.UDPAddr{IP: net.ParseIP("2607:f0d0:1002:51::4"), Port: 1313}
		anonAddr6 := transport.AnonymizeAddr(addr6)
		assert.Equal(t, "[2607:f0d0:1002::]:0", anonAddr6.String())
		assert.Equal(t, "2607:f0d0:1002::", anonAddr6.IP.String())
		assert.Equal(t, 0, anonAddr6.Port)
		assert.NotEqual(t, addr6, anonAddr6)
	})

	t.Run("anon addr failure", func(t *testing.T) {
		addr4 := net.UDPAddr{}
		anonAddr4 := transport.AnonymizeAddr(addr4)
		assert.Equal(t, addr4.IP, anonAddr4.IP)
	})
}

func TestFloatPrecision(t *testing.T) {
	val := -102.1683599948883057

	assert.Equal(t, -102.1683, transport.Float64Precision(val, 4))
	assert.Equal(t, -102.16, transport.Float64Precision(val, 2))
	assert.Equal(t, -102.16835, transport.Float64Precision(val, 5))
}
