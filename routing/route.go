package routing

import (
	"encoding/binary"
	"hash/fnv"
)

type Route struct {
	Relays []Relay
	Stats  Stats
}

func (r *Route) Decide(prevDecision Decision, nnStats Stats, directStats Stats, routeDecisions ...DecisionFunc) Decision {
	nextDecision := prevDecision
	for _, routeDecision := range routeDecisions {
		nextDecision = routeDecision(nextDecision, nnStats, directStats, r.Stats)
	}
	return nextDecision
}

func (r *Route) Hash() []byte {
	fnv64 := fnv.New64()
	id := make([]byte, 8)

	for _, relay := range r.Relays {
		binary.LittleEndian.PutUint64(id, relay.ID)
		fnv64.Write(id)
	}

	return fnv64.Sum(nil)
}

func (r *Route) Hash64() uint64 {
	fnv64 := fnv.New64()
	id := make([]byte, 8)

	for _, relay := range r.Relays {
		binary.LittleEndian.PutUint64(id, relay.ID)
		fnv64.Write(id)
	}

	return fnv64.Sum64()
}
