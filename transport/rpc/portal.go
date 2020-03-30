package rpc

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
)

type Portal struct {
	RedisClient redis.Cmdable
}

func (p *Portal) Relays(ctx context.Context, _ RelaysRequest) (*RelaysResponse, error) {
	hgetallResult := p.RedisClient.HGetAll(routing.HashKeyAllRelays)
	if hgetallResult.Err() != nil && hgetallResult.Err() != redis.Nil {
		return nil, hgetallResult.Err()
	}

	var relays []routing.Relay
	for _, v := range hgetallResult.Val() {
		var relay routing.Relay
		if err := relay.UnmarshalBinary([]byte(v)); err != nil {
			continue
		}
		relays = append(relays, relay)
	}

	return &RelaysResponse{
		Relays: relays,
	}, nil
}
