package fake_server

import (
	"math/rand"
	"testing"

	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func getTestSessionResponse(t *testing.T, routeType int32, nearRelaysChanged bool, committed bool) transport.SessionResponsePacket {
	const sessionDataBytes = 100
	sessionDataSlice := make([]byte, sessionDataBytes)
	_, err := rand.Read(sessionDataSlice)
	assert.NoError(t, err)

	sessionData := [transport.MaxSessionDataSize]byte{}
	copy(sessionData[:], sessionDataSlice)

	var numNearRelays int32
	var nearRelayIDs []uint64
	if nearRelaysChanged {
		numNearRelays = int32(rand.Intn(6))
		nearRelayIDs = make([]uint64, numNearRelays)
		for i := int32(0); i < numNearRelays; i++ {
			nearRelayIDs[i] = rand.Uint64()
		}
	}

	response := transport.SessionResponsePacket{
		SessionDataBytes:  sessionDataBytes,
		SessionData:       sessionData,
		RouteType:         routeType,
		NearRelaysChanged: nearRelaysChanged,
		NumNearRelays:     numNearRelays,
		NearRelayIDs:      nearRelayIDs,
		Committed:         committed,
	}

	return response
}

func validateTestSession(t *testing.T, sessionPrevSlice, session Session, response transport.SessionResponsePacket) {
	assert.Equal(t, sessionPrevSlice.sliceNumber+1, session.sliceNumber)
	assert.Equal(t, response.SessionDataBytes, session.sessionDataBytes)
	assert.Equal(t, response.SessionData, session.sessionData)

	if response.RouteType == routing.RouteTypeDirect {
		assert.False(t, session.next)
	} else {
		assert.True(t, session.next)
	}

	var numNearRelays int
	if response.NearRelaysChanged {
		numNearRelays = int(response.NumNearRelays)
		assert.Equal(t, numNearRelays, int(session.numNearRelays))
		assert.Equal(t, response.NearRelayIDs, session.nearRelayIDs)
	} else {
		numNearRelays = int(sessionPrevSlice.numNearRelays)
		assert.Equal(t, numNearRelays, int(session.numNearRelays))
		assert.Equal(t, sessionPrevSlice.nearRelayIDs, session.nearRelayIDs)
	}

	assert.Equal(t, response.Committed, session.committed)

	assert.Len(t, session.nearRelayRTT, numNearRelays)
	assert.Len(t, session.nearRelayJitter, numNearRelays)
	assert.Len(t, session.nearRelayPacketLoss, numNearRelays)

	assert.NotZero(t, session.jitter)

	assert.Equal(t, uint64(600), (session.packetsSent+session.packetsLost)*uint64(session.sliceNumber))

	assert.NotZero(t, session.directRTT)

	if response.RouteType == routing.RouteTypeDirect {
		assert.Equal(t, session.jitter, session.directJitter)

		if session.packetsLost > 0 {
			assert.NotZero(t, session.directPacketLoss)
		} else {
			assert.Zero(t, session.directPacketLoss)
		}

		assert.Zero(t, session.nextRTT)
		assert.Zero(t, session.nextJitter)
		assert.Zero(t, session.nextPacketLoss)
	} else {
		assert.Equal(t, session.jitter, session.nextJitter)

		if session.packetsLost > 0 {
			assert.NotZero(t, session.nextPacketLoss)
		} else {
			assert.Zero(t, session.nextPacketLoss)
		}

		assert.NotZero(t, session.directRTT)
		assert.NotZero(t, session.directJitter)
		assert.Zero(t, session.directPacketLoss)
	}
}

func TestNewSession(t *testing.T) {
	rand.Seed(0)

	session, err := NewSession()
	assert.NoError(t, err)
	assert.NotZero(t, session)
}

func TestSessionAdvance(t *testing.T) {
	rand.Seed(0)

	t.Run("direct response near relays changed", func(t *testing.T) {
		response := getTestSessionResponse(t, routing.RouteTypeDirect, true, false)

		session, err := NewSession()
		assert.NoError(t, err)

		prevSessionSlice := session
		session.Advance(response)

		validateTestSession(t, prevSessionSlice, session, response)
	})

	t.Run("direct response", func(t *testing.T) {
		response := getTestSessionResponse(t, routing.RouteTypeDirect, false, false)

		session, err := NewSession()
		assert.NoError(t, err)

		prevSessionSlice := session
		session.Advance(response)

		validateTestSession(t, prevSessionSlice, session, response)
	})

	t.Run("new route response", func(t *testing.T) {
		response := getTestSessionResponse(t, routing.RouteTypeNew, true, true)

		session, err := NewSession()
		assert.NoError(t, err)

		prevSessionSlice := session
		session.Advance(response)

		validateTestSession(t, prevSessionSlice, session, response)
	})

	t.Run("continue route response", func(t *testing.T) {
		response := getTestSessionResponse(t, routing.RouteTypeContinue, false, true)

		session, err := NewSession()
		assert.NoError(t, err)

		prevSessionSlice := session
		session.Advance(response)

		validateTestSession(t, prevSessionSlice, session, response)
	})
}
