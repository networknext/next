package portal_test

import (
	"testing"

	"github.com/networknext/next/modules/portal"

	"github.com/stretchr/testify/assert"
)

const NumIterations = 1

func TestSessionData(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomSessionData()
		value := writeData.Value()
		readData := portal.SessionData{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}

func TestSliceData(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomSliceData()
		value := writeData.Value()
		readData := portal.SliceData{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}

func TestNearRelayData(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomNearRelayData()
		value := writeData.Value()
		readData := portal.NearRelayData{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}

func TestServerData(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomServerData()
		value := writeData.Value()
		readData := portal.ServerData{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}

func TestRelayData(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomRelayData()
		value := writeData.Value()
		readData := portal.RelayData{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}

func TestRelaySample(t *testing.T) {
	t.Parallel()
	for i := 0; i < NumIterations; i++ {
		writeData := portal.GenerateRandomRelaySample()
		value := writeData.Value()
		readData := portal.RelaySample{}
		readData.Parse(value)
		assert.Equal(t, *writeData, readData)
	}
}
