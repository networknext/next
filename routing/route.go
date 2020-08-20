package routing

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net"
	"strings"
)

type Route struct {
	RelayIDs        []uint64      `json:"relayIDs"`
	RelayNames      []string      `json:"relayNames"`
	RelayAddrs      []net.UDPAddr `json:"relayAddrs"`
	RelayPublicKeys [][]byte      `json:"relayPublicKeys"`
	RelaySellers    []Seller      `json:"relaySellers"`
	Stats           Stats         `json:"stats"`
}

func (r Route) String() string {
	var sb strings.Builder
	sb.WriteString("stats=")
	sb.WriteString(r.Stats.String())
	sb.WriteString(" ")

	sb.WriteString("hash=")
	sb.WriteString(fmt.Sprintf("%d", r.Hash64()))
	sb.WriteString(" ")

	sb.WriteString("relayIDs=")
	for idx, relayID := range r.RelayIDs {
		sb.WriteString(fmt.Sprintf("%d", relayID))
		if idx < len(r.RelayIDs) {
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

	for _, relayID := range r.RelayIDs {
		binary.LittleEndian.PutUint64(id, relayID)
		fnv64.Write(id)
	}

	return fnv64.Sum(nil)
}

func (r *Route) Hash64() uint64 {
	fnv64 := fnv.New64()
	id := make([]byte, 8)

	for _, relayID := range r.RelayIDs {
		binary.LittleEndian.PutUint64(id, relayID)
		fnv64.Write(id)
	}

	return fnv64.Sum64()
}
