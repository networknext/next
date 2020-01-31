package transport_test

import (
	"testing"

	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestSDKVersion(t *testing.T) {
	t.Run("internal", func(t *testing.T) {
		internal := transport.SDKVersion{}
		assert.True(t, internal.IsInternal())

		notinternal := transport.SDKVersion{4, 5, 6}
		assert.False(t, notinternal.IsInternal())
	})

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

	t.Run("string", func(t *testing.T) {
		a := transport.SDKVersionMin

		assert.Equal(t, "3.3.2", a.String())
	})
}
