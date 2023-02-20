package portal_test

import (
	"testing"

	"github.com/networknext/backend/modules/portal"

	"github.com/stretchr/testify/assert"
)

const NumIterations = 1

func TestPortalSessionData(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomSessionData()
		value := writeData.Value()
		readData := portal.SessionData{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}

func TestPortalSliceData(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomSliceData()
		value := writeData.Value()
		readData := portal.SliceData{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}

func TestPortalNearRelayData(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomNearRelayData()
		value := writeData.Value()
		readData := portal.NearRelayData{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}

// todo: test that we can handle parsing empty string and fail gracefully

// todo: test that we can handle parsing garbage data and fail gracefully

// todo: test that we can handle parsing correct structure ||||| but garbage strings gracefully (randomly corrupt an entry, post save values)
