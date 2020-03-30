package definitions

import "github.com/networknext/backend/routing"

type PortalService interface {
	Relays(RelaysRequest) RelaysResponse
}

type RelaysRequest struct{}

type RelaysResponse struct {
	Relays []routing.Relay
}
