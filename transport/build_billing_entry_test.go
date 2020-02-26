package transport_test

import (
	"testing"

	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestBuildRouteRequest(t *testing.T) {
	expected := billing.RouteRequest{}

	updatePacket := transport.SessionUpdatePacket{}

	buyer := routing.Buyer{}

	serverData := transport.ServerCacheEntry{}

	location := routing.Location{}

	storer := storage.InMemory{}

	clientRelays := []routing.Relay{}

	actual := transport.BuildRouteRequest(updatePacket, buyer, serverData, location, &storer, clientRelays)

	assert.Equal(t, expected, actual)
}
