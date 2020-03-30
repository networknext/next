package rpc

import (
	"context"

	"github.com/networknext/backend/routing"
)

type Portal struct{}

func (p *Portal) Relays(ctx context.Context, relays RelaysRequest) (*RelaysResponse, error) {
	return &RelaysResponse{
		Relays: []routing.Relay{
			{ID: 1},
		},
	}, nil
}
