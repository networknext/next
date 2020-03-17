package stats

import (
	"strconv"
)

func NewEntityID(kind string, relayID uint64) *EntityId {
	return &EntityId{
		Kind: kind,
		Name: strconv.FormatUint(relayID, 10),
	}
}
