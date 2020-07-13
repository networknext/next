package routing

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"strings"
)

type Route struct {
	Relays []Relay `json:"relays"`
	Stats  Stats   `json:"stats"`
}

func (r Route) String() string {
	var sb strings.Builder
	sb.WriteString("stats=")
	sb.WriteString(r.Stats.String())
	sb.WriteString(" ")

	sb.WriteString("hash=")
	sb.WriteString(fmt.Sprintf("%d", r.Hash64()))
	sb.WriteString(" ")

	sb.WriteString("relays=")
	for idx, relay := range r.Relays {
		sb.WriteString(relay.Addr.String())
		if idx < len(r.Relays) {
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

func (r *Route) Decide(prevDecision Decision, lastNextStats *Stats, lastDirectStats *Stats, routeDecisions ...DecisionFunc) Decision {
	nextDecision := prevDecision
	if prevDecision.Reason == DecisionInitialSlice {
		nextDecision = Decision{}
	}
	for _, routeDecision := range routeDecisions {
		nextDecision = routeDecision(nextDecision, &r.Stats, lastNextStats, lastDirectStats)
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
