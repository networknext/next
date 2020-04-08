package jsonrpc_test

import (
	"context"
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestBuyers(t *testing.T) {
	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1"})

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.BuyersReply
		err := svc.Buyers(nil, &jsonrpc.BuyersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Buyers[0].ID, uint64(1))
		assert.Equal(t, reply.Buyers[0].Name, "local.local.1")
	})
}

func TestRelays(t *testing.T) {
	storer := storage.InMemory{}
	storer.AddRelay(context.Background(), routing.Relay{ID: 1, Name: "local.local.1"})
	storer.AddRelay(context.Background(), routing.Relay{ID: 2, Name: "local.local.2"})

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")
		assert.Equal(t, reply.Relays[1].ID, uint64(2))
		assert.Equal(t, reply.Relays[1].Name, "local.local.2")
	})

	t.Run("filter", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{Name: "local.1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Relays), 1)
		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")

		var empty jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{Name: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Relays), 0)
	})
}

func TestRelayStateUpdate(t *testing.T) {
	makeSvc := func() *jsonrpc.OpsService {
		return &jsonrpc.OpsService{
			Storage: &storage.InMemory{
				LocalRelays: []routing.Relay{
					{ID: 1, State: 0},
					{ID: 2, State: 123456},
				},
			},
		}
	}

	t.Run("found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayStateUpdate(nil, &jsonrpc.RelayStateUpdateArgs{
			RelayID:    1,
			RelayState: routing.RelayStateDisabled,
		}, &jsonrpc.RelayStateUpdateReply{})
		assert.NoError(t, err)

		relay, found := svc.Storage.Relay(1)
		assert.True(t, found)
		assert.Equal(t, routing.RelayStateDisabled, relay.State)

		relay, found = svc.Storage.Relay(2)
		assert.True(t, found)
		assert.Equal(t, routing.RelayState(123456), relay.State)
	})

	t.Run("not found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayStateUpdate(nil, &jsonrpc.RelayStateUpdateArgs{
			RelayID:    987654321,
			RelayState: routing.RelayStateDisabled,
		}, &jsonrpc.RelayStateUpdateReply{})
		assert.Error(t, err)

		relay, found := svc.Storage.Relay(1)
		assert.True(t, found)
		assert.Equal(t, routing.RelayState(0), relay.State)

		relay, found = svc.Storage.Relay(2)
		assert.True(t, found)
		assert.Equal(t, routing.RelayState(123456), relay.State)
	})
}

func TestDatacenters(t *testing.T) {
	storer := storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 1, Name: "local.local.1"})
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 2, Name: "local.local.2"})

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(nil, &jsonrpc.DatacentersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Datacenters[0].Name, "local.local.1")
		assert.Equal(t, reply.Datacenters[1].Name, "local.local.2")
	})

	t.Run("filter", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(nil, &jsonrpc.DatacentersArgs{Name: "local.1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Datacenters), 1)
		assert.Equal(t, reply.Datacenters[0].Name, "local.local.1")

		var empty jsonrpc.DatacentersReply
		err = svc.Datacenters(nil, &jsonrpc.DatacentersArgs{Name: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Datacenters), 0)
	})
}
