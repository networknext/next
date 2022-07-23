package transport_test

import (
	"testing"

	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	assert.Equal(t, "4.0.0", transport.SDKVersionMin.String())
}

func TestCompare(t *testing.T) {
	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		a := transport.SDKVersion{1, 1, 1}
		b := transport.SDKVersion{1, 1, 1}

		assert.Equal(t, transport.SDKVersionEqual, a.Compare(b))
	})

	t.Run("older", func(t *testing.T) {
		a := transport.SDKVersion{1, 1, 1}
		b := transport.SDKVersionMax

		assert.Equal(t, transport.SDKVersionOlder, a.Compare(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := transport.SDKVersionMin
		b := transport.SDKVersion{1, 1, 1}

		assert.Equal(t, transport.SDKVersionNewer, a.Compare(b))

		a = transport.SDKVersion{1, 2, 3}
		b = transport.SDKVersion{1, 1, 3}

		assert.Equal(t, transport.SDKVersionNewer, a.Compare(b))

		a = transport.SDKVersion{1, 2, 3}
		b = transport.SDKVersion{1, 2, 2}

		assert.Equal(t, transport.SDKVersionNewer, a.Compare(b))
	})
}

func TestAtLeast(t *testing.T) {
	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		a := transport.SDKVersionMin
		b := transport.SDKVersionMin

		assert.True(t, a.AtLeast(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := transport.SDKVersionMax
		b := transport.SDKVersionMin

		assert.True(t, a.AtLeast(b))
	})

	t.Run("older", func(t *testing.T) {
		a := transport.SDKVersionMin
		b := transport.SDKVersionMax

		assert.False(t, a.AtLeast(b))
	})

	t.Run("internal", func(t *testing.T) {
		a := transport.SDKVersion{}
		b := transport.SDKVersionMax

		assert.True(t, a.AtLeast(b))
	})
}
