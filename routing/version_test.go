package routing_test

import (
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestInternal(t *testing.T) {
	internal := routing.SDKVersion{}
	assert.True(t, internal.IsInternal())

	notinternal := routing.SDKVersion{4, 5, 6}
	assert.False(t, notinternal.IsInternal())
}

func TestString(t *testing.T) {
	assert.Equal(t, "3.3.2", routing.SDKVersionMin.String())
}

func TestCompare(t *testing.T) {
	t.Parallel()

	t.Run("internal equal", func(t *testing.T) {
		a := routing.SDKVersion{0, 0, 0}
		b := routing.SDKVersion{0, 0, 0}

		assert.Equal(t, routing.SDKVersionEqual, a.Compare(b))
	})

	t.Run("internal older", func(t *testing.T) {
		a := routing.SDKVersion{1, 1, 1}
		b := routing.SDKVersion{0, 0, 0}

		assert.Equal(t, routing.SDKVersionOlder, a.Compare(b))
	})

	t.Run("internal newer", func(t *testing.T) {
		a := routing.SDKVersion{0, 0, 0}
		b := routing.SDKVersion{1, 1, 1}

		assert.Equal(t, routing.SDKVersionNewer, a.Compare(b))
	})

	t.Run("older", func(t *testing.T) {
		a := routing.SDKVersion{1, 1, 1}
		b := routing.SDKVersionMax

		assert.Equal(t, routing.SDKVersionOlder, a.Compare(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := routing.SDKVersionMin
		b := routing.SDKVersion{1, 1, 1}

		assert.Equal(t, routing.SDKVersionNewer, a.Compare(b))

		a = routing.SDKVersion{1, 2, 3}
		b = routing.SDKVersion{1, 1, 3}

		assert.Equal(t, routing.SDKVersionNewer, a.Compare(b))

		a = routing.SDKVersion{1, 2, 3}
		b = routing.SDKVersion{1, 2, 2}

		assert.Equal(t, routing.SDKVersionNewer, a.Compare(b))
	})
}

func TestAtLeast(t *testing.T) {
	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		a := routing.SDKVersionMin
		b := routing.SDKVersionMin

		assert.True(t, a.AtLeast(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := routing.SDKVersionMax
		b := routing.SDKVersionMin

		assert.True(t, a.AtLeast(b))
	})

	t.Run("older", func(t *testing.T) {
		a := routing.SDKVersionMin
		b := routing.SDKVersionMax

		assert.False(t, a.AtLeast(b))
	})

	t.Run("internal", func(t *testing.T) {
		a := routing.SDKVersion{}
		b := routing.SDKVersionMax

		assert.True(t, a.AtLeast(b))
	})
}
