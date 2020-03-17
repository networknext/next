package stats

import (
	"strconv"

	"github.com/networknext/backend/crypto"
)

func NewEntityID(kind string, name string) *EntityId {
	return &EntityId{
		Kind: kind,
		Name: strconv.FormatUint(crypto.HashID(name), 10),
	}
}
