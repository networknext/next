package routing

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net"
	"strings"
)

type Route struct {
	NumRelays       int                    `json:"numRelays"`
	RelayIDs        [MaxRelays]uint64      `json:"relayIDs"`
	RelayNames      [MaxRelays]string      `json:"relayNames"`
	RelayAddrs      [MaxRelays]net.UDPAddr `json:"relayAddrs"`
	RelayPublicKeys [MaxRelays][]byte      `json:"relayPublicKeys"`
	RelaySellers    [MaxRelays]Seller      `json:"relaySellers"`
	Stats           Stats                  `json:"stats"`
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

func (r *Route) Hash64() uint64 {
	fnv64 := fnv.New64()
	id := make([]byte, 8)

	for i := 0; i < r.NumRelays; i++ {
		binary.LittleEndian.PutUint64(id, r.RelayIDs[i])
		fnv64.Write(id)
	}

	return fnv64.Sum64()
}
