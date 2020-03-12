package stats

import (
	"strconv"
)

func NewEntityID(kind string, ID uint64) *EntityId {
	return &EntityId{
		Kind: kind,
		Name: strconv.FormatUint(ID, 10),
	}
}
